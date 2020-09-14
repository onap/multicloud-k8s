/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package helm

import (
	"bytes"
	
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"

	utils "github.com/onap/multicloud-k8s/src/orchestrator/utils"

	pkgerrors "github.com/pkg/errors"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/helm/pkg/strvals"

	"github.com/ghodss/yaml"
	logger "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/validation"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/releaseutil"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/tiller"
	"k8s.io/helm/pkg/timeconv"
)

//KubernetesResourceTemplate - Represents the template that is used to create a particular
//resource in Kubernetes
type KubernetesResourceTemplate struct {
	// Tracks the apiVersion and Kind of the resource
	GVK schema.GroupVersionKind
	// Path to the file that contains the resource info
	FilePath string
}

// Template is the interface for all helm templating commands
// Any backend implementation will implement this interface and will
// access the functionality via this.
type Template interface {
	GenerateKubernetesArtifacts(
		chartPath string,
		valueFiles []string,
		values []string) (map[string][]string, error)
}

// TemplateClient implements the Template interface
// It will also be used to maintain any localized state
type TemplateClient struct {
	whitespaceRegex *regexp.Regexp
	kubeVersion     string
	kubeNameSpace   string
	releaseName     string
	manifestName    string
}

// NewTemplateClient returns a new instance of TemplateClient
func NewTemplateClient(k8sversion, namespace, releasename, manifestFileName string) *TemplateClient {
	return &TemplateClient{
		whitespaceRegex: regexp.MustCompile(`^\s*$`),
		// defaultKubeVersion is the default value of --kube-version flag
		kubeVersion:   k8sversion,
		kubeNameSpace: namespace,
		releaseName:   releasename,
		manifestName:  manifestFileName,
	}
}

// Combines valueFiles and values into a single values stream.
// values takes precedence over valueFiles
func (h *TemplateClient) processValues(valueFiles []string, values []string) ([]byte, error) {
	base := map[string]interface{}{}

	//Values files that are used for overriding the chart
	for _, filePath := range valueFiles {
		currentMap := map[string]interface{}{}

		var bytes []byte
		var err error
		if strings.TrimSpace(filePath) == "-" {
			bytes, err = ioutil.ReadAll(os.Stdin)
		} else {
			bytes, err = ioutil.ReadFile(filePath)
		}

		if err != nil {
			return []byte{}, err
		}

		if err := yaml.Unmarshal(bytes, &currentMap); err != nil {
			return []byte{}, fmt.Errorf("failed to parse %s: %s", filePath, err)
		}
		// Merge with the previous map
		base = h.mergeValues(base, currentMap)
	}

	//User specified value. Similar to ones provided by -x
	for _, value := range values {
		if err := strvals.ParseInto(value, base); err != nil {
			return []byte{}, fmt.Errorf("failed parsing --set data: %s", err)
		}
	}

	return yaml.Marshal(base)
}

func (h *TemplateClient) mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}
		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = h.mergeValues(destMap, nextMap)
	}
	return dest
}

// GenerateKubernetesArtifacts a mapping of type to fully evaluated helm template
func (h *TemplateClient) GenerateKubernetesArtifacts(inputPath string, valueFiles []string,
	values []string) ([]KubernetesResourceTemplate, error) {

	var outputDir, chartPath, namespace, releaseName string
	var retData []KubernetesResourceTemplate

	releaseName = h.releaseName
	namespace = h.kubeNameSpace

	// verify chart path exists
	if _, err := os.Stat(inputPath); err == nil {
		if chartPath, err = filepath.Abs(inputPath); err != nil {
			return retData, err
		}
	} else {
		return retData, err
	}

	//Create a temp directory in the system temp folder
	outputDir, err := ioutil.TempDir("", "helm-tmpl-")
	if err != nil {
		return retData, pkgerrors.Wrap(err, "Got error creating temp dir")
	}
	logger.Info(":: The o/p dir:: ", logger.Fields{"OutPutDirectory ":outputDir})
	

	if namespace == "" {
		namespace = "default"
	}

	// get combined values and create config
	rawVals, err := h.processValues(valueFiles, values)
	if err != nil {
		return retData, err
	}
	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}

	if msgs := validation.IsDNS1123Label(releaseName); releaseName != "" && len(msgs) > 0 {
		return retData, fmt.Errorf("release name %s is not a valid DNS label: %s", releaseName, strings.Join(msgs, ";"))
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	c, err := chartutil.Load(chartPath)
	if err != nil {
		return retData, pkgerrors.Errorf("Got error: %s", err.Error())
	}

	renderOpts := renderutil.Options{
		ReleaseOptions: chartutil.ReleaseOptions{
			Name:      releaseName,
			IsInstall: true,
			IsUpgrade: false,
			Time:      timeconv.Now(),
			Namespace: namespace,
		},
		KubeVersion: h.kubeVersion,
	}

	renderedTemplates, err := renderutil.Render(c, config, renderOpts)
	if err != nil {
		return retData, err
	}

	newRenderedTemplates := make(map[string]string)

	//Some manifests can contain multiple yaml documents
	//This step is splitting them up into multiple files
	//Each file contains only a single k8s kind
	for k, v := range renderedTemplates {
		//Splits into manifest-0, manifest-1 etc
		if filepath.Base(k) == "NOTES.txt" {
			continue
		}
		rmap := releaseutil.SplitManifests(v)
		count := 0
		for _, v1 := range rmap {
			key := fmt.Sprintf("%s-%d", k, count)
			newRenderedTemplates[key] = v1
			count = count + 1
		}
	}

	listManifests := manifest.SplitManifests(newRenderedTemplates)
	var manifestsToRender []manifest.Manifest
	//render all manifests in the chart
	manifestsToRender = listManifests
	for _, m := range tiller.SortByKind(manifestsToRender) {
		data := m.Content
		b := filepath.Base(m.Name)
		if b == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}

		// blank template after execution
		if h.whitespaceRegex.MatchString(data) {
			continue
		}

		mfilePath := filepath.Join(outputDir, m.Name)
		utils.EnsureDirectory(mfilePath)
		err = ioutil.WriteFile(mfilePath, []byte(data), 0666)
		if err != nil {
			return retData, err
		}

		gvk, err := getGroupVersionKind(data)
		if err != nil {
			return retData, err
		}

		kres := KubernetesResourceTemplate{
			GVK:      gvk,
			FilePath: mfilePath,
		}
		retData = append(retData, kres)
	}
	return retData, nil
}

func getGroupVersionKind(data string) (schema.GroupVersionKind, error) {
	out, err := k8syaml.ToJSON([]byte(data))
	if err != nil {
		return schema.GroupVersionKind{}, pkgerrors.Wrap(err, "Converting yaml to json")
	}

	simpleMeta := json.SimpleMetaFactory{}
	gvk, err := simpleMeta.Interpret(out)
	if err != nil {
		return schema.GroupVersionKind{}, pkgerrors.Wrap(err, "Parsing apiversion and kind")
	}

	return *gvk, nil
}

// Resolver is an interface exposes the helm related functionalities
type Resolver interface {
	Resolve(appContent, appProfileContent []byte, overrideValuesOfAppStr []string, appName string) ([]KubernetesResourceTemplate, error)
}


func cleanupTempFiles(fp string)  error {
	sa := strings.Split(fp, "/")
	dp:= "/" + sa[1] + "/" + sa[2] + "/"
	err := os.RemoveAll(dp)
	if err != nil {
		log.Error("Error while deleting dir", log.Fields{"Dir: ":dp})
		return err
	}
	log.Info("Clean up k8s-ext-tmp-dir::", log.Fields{"Dir: ": dp})
	return nil
}


// Resolve function
func (h *TemplateClient) Resolve(appContent []byte, appProfileContent []byte, overrideValuesOfAppStr []string, appName string) ([]KubernetesResourceTemplate, error) {

	var sortedTemplates []KubernetesResourceTemplate

	//chartBasePath is the tmp path where the appContent(rawHelmCharts) is extracted.
	chartBasePath, err := utils.ExtractTarBall(bytes.NewBuffer(appContent))
	defer cleanupTempFiles(chartBasePath)
	if err != nil {
		logger.Error("Error while extracting appContent", logger.Fields{})
		return sortedTemplates, pkgerrors.Wrap(err, "Error while extracting appContent")
	}
	logger.Info("The chartBasePath ::", logger.Fields{"chartBasePath":chartBasePath})

	//prPath is the tmp path where the appProfileContent is extracted.
	prPath, err := utils.ExtractTarBall(bytes.NewBuffer(appProfileContent))
	defer cleanupTempFiles(prPath)
	if err != nil {
		logger.Error("Error while extracting Profile Content", logger.Fields{})
		return sortedTemplates, pkgerrors.Wrap(err, "Error while extracting Profile Content")
	}
	logger.Info("The profile path:: ", logger.Fields{"Profile Path":prPath})

	prYamlClient, err := ProcessProfileYaml(prPath, h.manifestName)
	if err != nil {
		logger.Error("Error while processing Profile Manifest", logger.Fields{})
		return sortedTemplates, pkgerrors.Wrap(err, "Error while processing Profile Manifest")
	}
	logger.Info("Got the profileYamlClient..", logger.Fields{})

	err = prYamlClient.CopyConfigurationOverrides(chartBasePath)
	if err != nil {
		logger.Error("Error while copying configresources to chart", logger.Fields{})
		return sortedTemplates, pkgerrors.Wrap(err, "Error while copying configresources to chart")
	}

	chartPath := filepath.Join(chartBasePath, appName)
	sortedTemplates, err = h.GenerateKubernetesArtifacts(chartPath, []string{prYamlClient.GetValues()}, overrideValuesOfAppStr)
	if err != nil {
		logger.Error("Error while generating final k8s yaml", logger.Fields{})
		return sortedTemplates, pkgerrors.Wrap(err, "Error while generating final k8s yaml")
	}
	return sortedTemplates, nil
}
