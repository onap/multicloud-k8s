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
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"
)

// Compile time check to see if servicePlugin implements the correct interface
var _ plugin.Reference = servicePlugin{}

// ExportedVariable is what we will look for when calling the plugin
var ExportedVariable servicePlugin

type servicePlugin struct {
}

// Create a service object in a specific Kubernetes cluster
func (p servicePlugin) Create(yamlFilePath string, namespace string, client plugin.KubernetesConnector) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	obj, err := utils.DecodeYAML(yamlFilePath, nil)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Decode service object error")
	}

	service, ok := obj.(*coreV1.Service)
	if !ok {
		return "", pkgerrors.New("Decoded object contains another resource different than Service")
	}
	service.Namespace = namespace

	labels := service.GetLabels()
	//Check if labels exist for this object
	if labels == nil {
		labels = map[string]string{}
	}
	labels[config.GetConfiguration().KubernetesLabelName] = client.GetInstanceID()
	service.SetLabels(labels)

	result, err := client.GetStandardClient().CoreV1().Services(namespace).Create(service)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Service error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// List of existing services hosted in a specific Kubernetes cluster
// gvk parameter is not used as this plugin is specific to services only
func (p servicePlugin) List(gvk schema.GroupVersionKind, namespace string, client plugin.KubernetesConnector) ([]helm.KubernetesResource, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: utils.ResourcesListLimit,
	}

	list, err := client.GetStandardClient().CoreV1().Services(namespace).List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Service list error")
	}

	result := make([]helm.KubernetesResource, 0, utils.ResourcesListLimit)
	if list != nil {
		for _, service := range list.Items {
			log.Printf("%v", service.Name)
			result = append(result,
				helm.KubernetesResource{
					GVK: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Service",
					},
					Name: service.GetName(),
				})
		}
	}

	return result, nil
}

// Delete an existing service hosted in a specific Kubernetes cluster
func (p servicePlugin) Delete(resource helm.KubernetesResource, namespace string, client plugin.KubernetesConnector) error {
	if namespace == "" {
		namespace = "default"
	}

	deletePolicy := metaV1.DeletePropagationForeground
	opts := &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	log.Println("Deleting service: " + resource.Name)
	if err := client.GetStandardClient().CoreV1().Services(namespace).Delete(resource.Name, opts); err != nil {
		return pkgerrors.Wrap(err, "Delete service error")
	}

	return nil
}

// Get an existing service hosted in a specific Kubernetes cluster
func (p servicePlugin) Get(resource helm.KubernetesResource, namespace string, client plugin.KubernetesConnector) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.GetOptions{}
	service, err := client.GetStandardClient().CoreV1().Services(namespace).Get(resource.Name, opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get Service error")
	}

	return service.Name, nil
}
