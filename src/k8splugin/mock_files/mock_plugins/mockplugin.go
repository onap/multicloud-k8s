/*
Copyright 2018 Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ExportedVariable is what we will look for when calling the plugin
var ExportedVariable mockPlugin

type mockPlugin struct {
}

// Create object in a specific Kubernetes resource
func (p mockPlugin) Create(yamlFilePath string, namespace string, client plugin.KubernetesConnector) (string, error) {
	return "resource-name", nil
}

// List of existing resources
func (p mockPlugin) List(gvk schema.GroupVersionKind, namespace string,
	client plugin.KubernetesConnector) ([]helm.KubernetesResource, error) {
	returnVal := []helm.KubernetesResource{
		{
			Name: "resource-name-1",
		},
		{
			Name: "resource-name-2",
		},
	}
	return returnVal, nil
}

// Delete existing resources
func (p mockPlugin) Delete(resource helm.KubernetesResource, namespace string, client plugin.KubernetesConnector) error {
	return nil
}

// Get existing resource host
func (p mockPlugin) Get(resource helm.KubernetesResource, namespace string, client plugin.KubernetesConnector) (string, error) {
	return resource.Name, nil
}
