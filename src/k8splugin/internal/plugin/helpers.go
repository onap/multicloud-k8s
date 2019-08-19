/*
 * Copyright 2019 Intel Corporation, Inc
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

package plugin

import (
	"log"
	"strings"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"

	pkgerrors "github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// KubernetesConnector is an interface that is expected to be implemented
// by any code that calls the plugin framework functions.
// It implements methods that are needed by the plugins to get Kubernetes
// clients and other information needed to interface with Kubernetes
type KubernetesConnector interface {
	//GetMapper returns the RESTMapper that was created for this client
	GetMapper() meta.RESTMapper

	//GetDynamicClient returns the dynamic client that is needed for
	//unstructured REST calls to the apiserver
	GetDynamicClient() dynamic.Interface

	// GetStandardClient returns the standard client that can be used to handle
	// standard kubernetes kinds
	GetStandardClient() kubernetes.Interface

	//GetInstanceID returns the InstanceID for tracking during creation
	GetInstanceID() string
}

// Reference is the interface that is implemented
type Reference interface {
	//Create a kubernetes resource described by the yaml in yamlFilePath
	Create(yamlFilePath string, namespace string, client KubernetesConnector) (string, error)

	//Get a kubernetes resource based on the groupVersionKind and resourceName provided in resource
	Get(resource helm.KubernetesResource, namespace string, client KubernetesConnector) (string, error)

	//List all resources of the specified GroupVersionKind in the given namespace
	//If gvk is empty, the plugin will return all supported objects in the namespace
	List(gvk schema.GroupVersionKind, namespace string, client KubernetesConnector) ([]helm.KubernetesResource, error)

	//Delete a kubernetes resource described in the provided namespace
	Delete(resource helm.KubernetesResource, namespace string, client KubernetesConnector) error
}

// GetPluginByKind returns a plugin by the kind name
// If plugin does not exist, it will return the generic plugin
// TODO: Change this once we have a plugin registration mechanism
func GetPluginByKind(kind string) (Reference, error) {

	typePlugin, ok := utils.LoadedPlugins[strings.ToLower(kind)]
	if !ok {
		log.Println("No plugin for kind " + kind + " found. Using generic Plugin")
		typePlugin, ok = utils.LoadedPlugins["generic"]
		if !ok {
			return nil, pkgerrors.New("No generic plugin found")
		}
	}

	symbol, err := typePlugin.Lookup("ExportedVariable")
	if err != nil {
		return nil, pkgerrors.Wrap(err, "No ExportedVariable symbol found")
	}

	//Assert if it implements the PluginReference interface
	pluginImpl, ok := symbol.(Reference)
	if !ok {
		return nil, pkgerrors.New("ExportedVariable does not implement plugins.Reference interface type")
	}

	return pluginImpl, nil
}

// TagPodsIfPresent finds the PodTemplateSpec from any workload
// object that contains it and changes the spec to include the tag label
func TagPodsIfPresent(unstruct *unstructured.Unstructured, tag string) {

	spec, ok := unstruct.Object["spec"].(map[string]interface{})
	if !ok {
		log.Println("Error converting spec to map")
		return
	}

	template, ok := spec["template"].(map[string]interface{})
	if !ok {
		log.Println("Error converting template to map")
		return
	}

	//Attempt to convert the template to a podtemplatespec.
	//This is to check if we have any pods being created.
	podTemplateSpec := &corev1.PodTemplateSpec{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(template, podTemplateSpec)
	if err != nil {
		log.Println("Did not find a podTemplateSpec: " + err.Error())
		return
	}

	labels := podTemplateSpec.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}
	labels[config.GetConfiguration().KubernetesLabelName] = tag
	podTemplateSpec.SetLabels(labels)

	updatedTemplate, err := runtime.DefaultUnstructuredConverter.ToUnstructured(podTemplateSpec)

	//Set the label
	spec["template"] = updatedTemplate
}
