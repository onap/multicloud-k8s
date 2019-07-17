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
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"
)

// Compile time check to see if genericPlugin implements the correct interface
var _ plugin.Reference = genericPlugin{}

// ExportedVariable is what we will look for when calling the generic plugin
var ExportedVariable genericPlugin

type genericPlugin struct {
}

// Create deployment object in a specific Kubernetes cluster
func (g genericPlugin) Create(yamlFilePath string, namespace string, client plugin.KubernetesConnector) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	//Decode the yaml file to create a runtime.Object
	unstruct := &unstructured.Unstructured{}
	//Ignore the returned obj as we expect the data in unstruct
	_, err := utils.DecodeYAML(yamlFilePath, unstruct)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Decode deployment object error")
	}

	dynClient := client.GetDynamicClient()
	mapper := client.GetMapper()

	gvk := unstruct.GroupVersionKind()
	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Mapping kind to resource error")
	}

	//Add the tracking label to all resources created here
	labels := unstruct.GetLabels()
	//Check if labels exist for this object
	if labels == nil {
		labels = map[string]string{}
	}
	labels[config.GetConfiguration().KubernetesLabelName] = client.GetInstanceID()
	unstruct.SetLabels(labels)

	// This checks if the resource we are creating has a podSpec in it
	// Eg: Deployment, StatefulSet, Job etc..
	// If a PodSpec is found, the label will be added to it too.
	plugin.TagPodsIfPresent(unstruct, client.GetInstanceID())

	gvr := mapping.Resource
	var createdObj *unstructured.Unstructured

	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		createdObj, err = dynClient.Resource(gvr).Namespace(namespace).Create(unstruct, metav1.CreateOptions{})
	case meta.RESTScopeNameRoot:
		createdObj, err = dynClient.Resource(gvr).Create(unstruct, metav1.CreateOptions{})
	default:
		return "", pkgerrors.New("Got an unknown RESTSCopeName for mapping: " + gvk.String())
	}

	if err != nil {
		return "", pkgerrors.Wrap(err, "Create object error")
	}

	return createdObj.GetName(), nil
}

// Get an existing resource hosted in a specific Kubernetes cluster
func (g genericPlugin) Get(resource helm.KubernetesResource,
	namespace string, client plugin.KubernetesConnector) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	dynClient := client.GetDynamicClient()
	mapper := client.GetMapper()

	mapping, err := mapper.RESTMapping(schema.GroupKind{
		Group: resource.GVK.Group,
		Kind:  resource.GVK.Kind,
	}, resource.GVK.Version)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Mapping kind to resource error")
	}

	gvr := mapping.Resource
	opts := metav1.GetOptions{}
	var unstruct *unstructured.Unstructured
	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		unstruct, err = dynClient.Resource(gvr).Namespace(namespace).Get(resource.Name, opts)
	case meta.RESTScopeNameRoot:
		unstruct, err = dynClient.Resource(gvr).Get(resource.Name, opts)
	default:
		return "", pkgerrors.New("Got an unknown RESTSCopeName for mapping: " + resource.GVK.String())
	}

	if err != nil {
		return "", pkgerrors.Wrap(err, "Delete object error")
	}

	return unstruct.GetName(), nil
}

// List all existing resources of the GroupVersionKind
// TODO: Implement in seperate patch
func (g genericPlugin) List(gvk schema.GroupVersionKind, namespace string,
	client plugin.KubernetesConnector) ([]helm.KubernetesResource, error) {

	var returnData []helm.KubernetesResource
	return returnData, nil
}

// Delete an existing resource hosted in a specific Kubernetes cluster
func (g genericPlugin) Delete(resource helm.KubernetesResource, namespace string, client plugin.KubernetesConnector) error {
	if namespace == "" {
		namespace = "default"
	}

	dynClient := client.GetDynamicClient()
	mapper := client.GetMapper()

	mapping, err := mapper.RESTMapping(schema.GroupKind{
		Group: resource.GVK.Group,
		Kind:  resource.GVK.Kind,
	}, resource.GVK.Version)
	if err != nil {
		return pkgerrors.Wrap(err, "Mapping kind to resource error")
	}

	gvr := mapping.Resource
	deletePolicy := metav1.DeletePropagationForeground
	opts := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	switch mapping.Scope.Name() {
	case meta.RESTScopeNameNamespace:
		err = dynClient.Resource(gvr).Namespace(namespace).Delete(resource.Name, opts)
	case meta.RESTScopeNameRoot:
		err = dynClient.Resource(gvr).Delete(resource.Name, opts)
	default:
		return pkgerrors.New("Got an unknown RESTSCopeName for mapping: " + resource.GVK.String())
	}

	if err != nil {
		return pkgerrors.Wrap(err, "Delete object error")
	}
	return nil
}
