/*
Copyright 2018 Intel Corporation.
Copyright © 2021 Nokia Bell Labs.
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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"time"
)

// ExportedVariable is what we will look for when calling the plugin
var ExportedVariable mockPlugin

type mockPlugin struct {
}

func (g mockPlugin) WatchUntilReady(
	timeout time.Duration,
	ns string,
	res helm.KubernetesResource,
	mapper meta.RESTMapper,
	restClient rest.Interface,
	objType runtime.Object,
	clientSet kubernetes.Interface) error {
	return nil
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

// Update existing resources
func (p mockPlugin) Update(yamlFilePath string, namespace string, client plugin.KubernetesConnector) (string, error) {

        return "", nil
}

