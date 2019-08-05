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
	"log"

	pkgerrors "github.com/pkg/errors"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"
)

// Compile time check to see if namespacePlugin implements the correct interface
var _ plugin.Reference = namespacePlugin{}

// ExportedVariable is what we will look for when calling the plugin
var ExportedVariable namespacePlugin

type namespacePlugin struct {
}

// Create a namespace object in a specific Kubernetes cluster
func (p namespacePlugin) Create(yamlFilePath string, namespace string, client plugin.KubernetesConnector) (string, error) {
	namespaceObj := &coreV1.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.GetStandardClient().CoreV1().Namespaces().Create(namespaceObj)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Namespace error")
	}
	log.Printf("Namespace (%s) created", namespace)

	return namespace, nil
}

// Get an existing namespace hosted in a specific Kubernetes cluster
func (p namespacePlugin) Get(resource helm.KubernetesResource, namespace string, client plugin.KubernetesConnector) (string, error) {
	opts := metaV1.GetOptions{}
	ns, err := client.GetStandardClient().CoreV1().Namespaces().Get(resource.Name, opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get Namespace error")
	}

	return ns.Name, nil
}

// Delete an existing namespace hosted in a specific Kubernetes cluster
func (p namespacePlugin) Delete(resource helm.KubernetesResource, namespace string, client plugin.KubernetesConnector) error {
	deletePolicy := metaV1.DeletePropagationForeground
	opts := &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	log.Println("Deleting namespace: " + resource.Name)
	if err := client.GetStandardClient().CoreV1().Namespaces().Delete(resource.Name, opts); err != nil {
		return pkgerrors.Wrap(err, "Delete namespace error")
	}

	return nil
}

// List of existing namespaces hosted in a specific Kubernetes cluster
// This plugin ignores both gvk and namespace arguments
func (p namespacePlugin) List(gvk schema.GroupVersionKind, namespace string, client plugin.KubernetesConnector) ([]helm.KubernetesResource, error) {
	opts := metaV1.ListOptions{
		Limit: utils.ResourcesListLimit,
	}

	list, err := client.GetStandardClient().CoreV1().Namespaces().List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Namespace list error")
	}

	result := make([]helm.KubernetesResource, 0, utils.ResourcesListLimit)
	if list != nil {
		for _, ns := range list.Items {
			log.Printf("%v", ns.Name)
			result = append(result,
				helm.KubernetesResource{
					GVK: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Namespace",
					},
					Name: ns.Name,
				})
		}
	}

	return result, nil
}
