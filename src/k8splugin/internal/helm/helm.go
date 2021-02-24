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
	"k8s.io/helm/pkg/strvals"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"

	"github.com/ghodss/yaml"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/util/validation"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/hooks"
	"k8s.io/helm/pkg/manifest"
	"k8s.io/helm/pkg/proto/hapi/chart"
	protorelease "k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/releaseutil"
	"k8s.io/helm/pkg/renderutil"
	"k8s.io/helm/pkg/tiller"
	"k8s.io/helm/pkg/timeconv"
)

// Template is the interface for all helm templating commands
// Any backend implementation will implement this interface and will
// access the functionality via this.
// FIXME Template is not referenced anywhere
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

// Define hooks that are honored by k8splugin
var honoredEvents = map[string]protorelease.Hook_Event{
	hooks.ReleaseTestSuccess: protorelease.Hook_RELEASE_TEST_SUCCESS,
	hooks.ReleaseTestFailure: protorelease.Hook_RELEASE_TEST_FAILURE,
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

// Checks whether resource is a hook and if it is, returns hook struct
//Logic is based on private method
//file *manifestFile) sort(result *result) error
//of helm/pkg/tiller package
func isHook(path, resource string) (*protorelease.Hook, error) {

	var entry releaseutil.SimpleHead
	err := yaml.Unmarshal([]byte(resource), &entry)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Loading resource to YAML")
	}
	//If resource has no metadata it can't be a hook
	if entry.Metadata == nil ||
		entry.Metadata.Annotations == nil ||
		len(entry.Metadata.Annotations) == 0 {
		return nil, nil
	}
	//Determine hook weight
	hookWeight, err := strconv.Atoi(entry.Metadata.Annotations[hooks.HookWeightAnno])
	if err != nil {
		hookWeight = 0
	}
	//Prepare hook obj
	resultHook := &protorelease.Hook{
		Name:           entry.Metadata.Name,
		Kind:           entry.Kind,
		Path:           path,
		Manifest:       resource,
		Events:         []protorelease.Hook_Event{},
		Weight:         int32(hookWeight),
		DeletePolicies: []protorelease.Hook_DeletePolicy{},
	}
	//Determine hook's events
	hookTypes, ok := entry.Metadata.Annotations[hooks.HookAnno]
	if !ok {
		return resultHook, nil
	}
	for _, hookType := range strings.Split(hookTypes, ",") {
		hookType = strings.ToLower(strings.TrimSpace(hookType))
		e, ok := honoredEvents[hookType]
		if ok {
			resultHook.Events = append(resultHook.Events, e)
		}
	}
	return resultHook, nil
}

// GenerateKubernetesArtifacts a mapping of type to fully evaluated helm template
func (h *TemplateClient) GenerateKubernetesArtifacts(inputPath string, valueFiles []string,
	values []string) ([]KubernetesResourceTemplate, []*protorelease.Hook, error) {

	var outputDir, chartPath, namespace, releaseName string
	var retData []KubernetesResourceTemplate
	var hookList []*protorelease.Hook

	releaseName = h.releaseName
	namespace = h.kubeNameSpace

	// verify chart path exists
	if _, err := os.Stat(inputPath); err == nil {
		if chartPath, err = filepath.Abs(inputPath); err != nil {
			return retData, hookList, err
		}
	} else {
		return retData, hookList, err
	}

	//Create a temp directory in the system temp folder
	outputDir, err := ioutil.TempDir("", "helm-tmpl-")
	if err != nil {
		return retData, hookList, pkgerrors.Wrap(err, "Got error creating temp dir")
	}

	if namespace == "" {
		namespace = "default"
	}

	// get combined values and create config
	rawVals, err := h.processValues(valueFiles, values)
	if err != nil {
		return retData, hookList, err
	}
	config := &chart.Config{Raw: string(rawVals), Values: map[string]*chart.Value{}}

	if msgs := validation.IsDNS1123Label(releaseName); releaseName != "" && len(msgs) > 0 {
		return retData, hookList, fmt.Errorf("release name %s is not a valid DNS label: %s", releaseName, strings.Join(msgs, ";"))
	}

	// Check chart requirements to make sure all dependencies are present in /charts
	c, err := chartutil.Load(chartPath)
	if err != nil {
		return retData, hookList, pkgerrors.Errorf("Got error: %s", err.Error())
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
		return retData, hookList, err
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

		// Iterating over map can yield different order at times
		// so first we'll sort keys
		sortedKeys := make([]string, len(rmap))
		for k1, _ := range rmap {
			sortedKeys = append(sortedKeys, k1)
		}
		// This makes empty files have the lowest indices
		sort.Strings(sortedKeys)

		for k1, v1 := range sortedKeys {
			key := fmt.Sprintf("%s-%d", k, k1)
			newRenderedTemplates[key] = rmap[v1]
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
		if h.emptyRegex.MatchString(data) {
			continue
		}

		hook, _ := isHook(b, data)
		// if hook is not nil, then append it to hooks list and continue
		// if it's not, disregard error
		if hook != nil {
			hookList = append(hookList, hook)
			continue
		}

		mfilePath := filepath.Join(outputDir, m.Name)
		utils.EnsureDirectory(mfilePath)
		err = ioutil.WriteFile(mfilePath, []byte(data), 0666)
		if err != nil {
			return retData, hookList, err
		}

		gvk, err := getGroupVersionKind(data)
		if err != nil {
			return retData, hookList, err
		}

		kres := KubernetesResourceTemplate{
			GVK:      gvk,
			FilePath: mfilePath,
		}
		retData = append(retData, kres)
	}
	return retData, hookList, nil
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
