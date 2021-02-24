/*
 * Copyright 2018 Intel Corporation, Inc
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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/validation"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/releaseutil"
	"helm.sh/helm/v3/pkg/cli"
	helmOptions "helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
)

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
	emptyRegex    *regexp.Regexp
	kubeVersion   string
	kubeNameSpace string
	releaseName   string
}

// NewTemplateClient returns a new instance of TemplateClient
func NewTemplateClient(k8sversion, namespace, releasename string) *TemplateClient {
	return &TemplateClient{
		// emptyRegex defines template content that could be considered empty yaml-wise
		emptyRegex: regexp.MustCompile(`(?m)\A(^(\s*#.*|\s*)$\n?)*\z`),
		// defaultKubeVersion is the default value of --kube-version flag
		kubeVersion:   k8sversion,
		kubeNameSpace: namespace,
		releaseName:   releasename,
	}
}

// Combines valueFiles and values into a single values stream.
// values takes precedence over valueFiles
func (h *TemplateClient) processValues(valueFiles []string, values []string) (map[string]interface{}, error) {
	settings := cli.New()
	providers := getter.All(settings)
	options := helmOptions.Options {
		ValueFiles: valueFiles,
		Values: values,
	}
	base, err := options.MergeValues(providers)
	if err != nil {
	   return nil, err
	}

	return base, nil
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

	if namespace == "" {
		namespace = "default"
	}

	// get combined values and create config
	rawVals, err := h.processValues(valueFiles, values)
	if err != nil {
		return retData, err
	}

	if msgs := validation.IsDNS1123Label(releaseName); releaseName != "" && len(msgs) > 0 {
		return retData, fmt.Errorf("release name %s is not a valid DNS label: %s", releaseName, strings.Join(msgs, ";"))
	}

	// Initialize the install client
	client := action.NewInstall(&action.Configuration{})
	client.DryRun = true
	client.ClientOnly = true
	client.ReleaseName = releaseName
	client.IncludeCRDs = true

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return retData, err
	}

	if chartRequested.Metadata.Type != "" && chartRequested.Metadata.Type != "application" {
		return nil, fmt.Errorf(
			"chart %q has an unsupported type and is not installable: %q",
			chartRequested.Metadata.Name,
			chartRequested.Metadata.Type,
		)
	}

	client.Namespace = namespace
	release, err := client.Run(chartRequested, rawVals)
	if err != nil {
		return nil, err
	}
	// SplitManifests returns integer-sortable so that manifests get output
	// in the same order as the input by `BySplitManifestsOrder`.
	rmap := releaseutil.SplitManifests(release.Manifest)
	keys := make([]string, 0, len(rmap))
	for k := range rmap {
		keys = append(keys, k)
	}
	// Sort Keys to get Sort Order
	sort.Sort(releaseutil.BySplitManifestsOrder(keys))
	for _, k := range keys {
		data := rmap[k]
		b := filepath.Base(k)
		if b == "NOTES.txt" {
			continue
		}
		if strings.HasPrefix(b, "_") {
			continue
		}
		// blank template after execution
		if h.emptyRegex.MatchString(data) {
			continue
		}
		mfilePath := filepath.Join(outputDir, k)
		utils.EnsureDirectory(mfilePath)
		err = ioutil.WriteFile(mfilePath, []byte(data), 0600)
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
		return schema.GroupVersionKind{}, pkgerrors.Wrap(err, "Converting yaml to json:\n"+data)
	}

	simpleMeta := json.SimpleMetaFactory{}
	gvk, err := simpleMeta.Interpret(out)
	if err != nil {
		return schema.GroupVersionKind{}, pkgerrors.Wrap(err, "Parsing apiversion and kind")
	}

	return *gvk, nil
}
