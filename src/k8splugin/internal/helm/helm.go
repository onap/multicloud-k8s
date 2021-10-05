/*
 * Copyright 2018 Intel Corporation, Inc
 * Copyright © 2021 Samsung Electronics
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
	"strings"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/utils"

	pkgerrors "github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	helmOptions "helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/releaseutil"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/validation"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
)

// Template is the interface for all helm templating commands
// Any backend implementation will implement this interface and will
// access the functionality via this.
// FIXME Template is not referenced anywhere
type Template interface {
	GenerateKubernetesArtifacts(
		chartPath string,
		valueFiles []string,
		values []string) ([]KubernetesResourceTemplate, []KubernetesResourceTemplate, []*Hook, error)
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
	options := helmOptions.Options{
		ValueFiles: valueFiles,
		Values:     values,
	}
	base, err := options.MergeValues(providers)
	if err != nil {
		return nil, err
	}

	return base, nil
}

// GenerateKubernetesArtifacts a mapping of type to fully evaluated helm template
func (h *TemplateClient) GenerateKubernetesArtifacts(inputPath string, valueFiles []string,
	values []string) ([]KubernetesResourceTemplate, []KubernetesResourceTemplate, []*Hook, error) {

	var outputDir, chartPath, namespace, releaseName string
	var retData []KubernetesResourceTemplate
	var crdData []KubernetesResourceTemplate
	var hookList []*Hook

	releaseName = h.releaseName
	namespace = h.kubeNameSpace

	// verify chart path exists
	if _, err := os.Stat(inputPath); err == nil {
		if chartPath, err = filepath.Abs(inputPath); err != nil {
			return retData, crdData, hookList, err
		}
	} else {
		return retData, crdData, hookList, err
	}

	//Create a temp directory in the system temp folder
	outputDir, err := ioutil.TempDir("", "helm-tmpl-")
	if err != nil {
		return retData, crdData, hookList, pkgerrors.Wrap(err, "Got error creating temp dir")
	}

	if namespace == "" {
		namespace = "default"
	}

	// get combined values and create config
	rawVals, err := h.processValues(valueFiles, values)
	if err != nil {
		return retData, crdData, hookList, err
	}

	if msgs := validation.IsDNS1123Label(releaseName); releaseName != "" && len(msgs) > 0 {
		return retData, crdData, hookList, fmt.Errorf("release name %s is not a valid DNS label: %s", releaseName, strings.Join(msgs, ";"))
	}

	// Initialize the install client
	client := action.NewInstall(&action.Configuration{})
	client.DryRun = true
	client.ClientOnly = true
	client.ReleaseName = releaseName
	client.IncludeCRDs = false
	client.DisableHooks = true //to ensure no duplicates in case of defined pre/post install hooks

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := loader.Load(chartPath)
	if err != nil {
		return retData, crdData, hookList, err
	}

	if chartRequested.Metadata.Type != "" && chartRequested.Metadata.Type != "application" {
		return retData, crdData, hookList, fmt.Errorf(
			"chart %q has an unsupported type and is not installable: %q",
			chartRequested.Metadata.Name,
			chartRequested.Metadata.Type,
		)
	}

	for _, crd := range chartRequested.CRDObjects() {
		if strings.HasPrefix(crd.Name, "_") {
			continue
		}
		filePath := filepath.Join(outputDir, crd.Name)
		data := string(crd.File.Data)
		// blank template after execution
		if h.emptyRegex.MatchString(data) {
			continue
		}
		utils.EnsureDirectory(filePath)
		err = ioutil.WriteFile(filePath, []byte(crd.File.Data), 0600)
		if err != nil {
			return retData, crdData, hookList, err
		}
		gvk, err := getGroupVersionKind(data)
		if err != nil {
			return retData, crdData, hookList, err
		}
		kres := KubernetesResourceTemplate{
			GVK:      gvk,
			FilePath: filePath,
		}
		crdData = append(crdData, kres)
	}
	client.Namespace = namespace
	release, err := client.Run(chartRequested, rawVals)
	if err != nil {
		return retData, crdData, hookList, err
	}
	// SplitManifests returns integer-sortable so that manifests get output
	// in the same order as the input by `BySplitManifestsOrder`.
	rmap := releaseutil.SplitManifests(release.Manifest)
	// We won't get any meaningful hooks from here
	_, m, err := releaseutil.SortManifests(rmap, nil, releaseutil.InstallOrder)
	if err != nil {
		return retData, crdData, hookList, err
	}
	for _, k := range m {
		data := k.Content
		b := filepath.Base(k.Name)
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
		mfilePath := filepath.Join(outputDir, k.Name)
		utils.EnsureDirectory(mfilePath)
		err = ioutil.WriteFile(mfilePath, []byte(k.Content), 0600)
		if err != nil {
			return retData, crdData, hookList, err
		}
		gvk, err := getGroupVersionKind(data)
		if err != nil {
			return retData, crdData, hookList, err
		}
		kres := KubernetesResourceTemplate{
			GVK:      gvk,
			FilePath: mfilePath,
		}
		retData = append(retData, kres)
	}
	for _, h := range release.Hooks {
		hFilePath := filepath.Join(outputDir, h.Name)
		utils.EnsureDirectory(hFilePath)
		err = ioutil.WriteFile(hFilePath, []byte(h.Manifest), 0600)
		if err != nil {
			return retData, crdData, hookList, err
		}
		gvk, err := getGroupVersionKind(h.Manifest)
		if err != nil {
			return retData, crdData, hookList, err
		}
		hookList = append(hookList, &Hook{*h, KubernetesResourceTemplate{gvk, hFilePath}})
	}
	return retData, crdData, hookList, nil
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

//GetReverseK8sResources reverse list of resources for delete purpose
func GetReverseK8sResources(resources []KubernetesResource) []KubernetesResource {
	reversed := []KubernetesResource{}

	for i := len(resources) - 1; i >= 0; i-- {
		reversed = append(reversed, resources[i])
	}
	return reversed
}
