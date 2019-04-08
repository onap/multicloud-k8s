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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"

	utils "k8splugin/internal"
	"k8splugin/internal/app"
)

type genericPlugin struct {
}

var kindToGVRMap = map[string]schema.GroupVersionResource{
	"ConfigMap":   schema.GroupVersionResource{Group: "core", Version: "v1", Resource: "configmaps"},
	"StatefulSet": schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
	"Job":         schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "jobs"},
	"Pod":         schema.GroupVersionResource{Group: "core", Version: "v1", Resource: "pods"},
	"DaemonSet":   schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "daemonsets"},
	"CustomResourceDefinition": schema.GroupVersionResource{
		Group: "apiextensions.k8s.io", Version: "v1beta1", Resource: "customresourcedefinitions",
	},
}

// Create deployment object in a specific Kubernetes cluster
func (g genericPlugin) Create(yamlFilePath string, namespace string, client *app.KubernetesClient) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	//Decode the yaml file to create a runtime.Object
	obj, err := utils.DecodeYAML(yamlFilePath, nil)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Decode deployment object error")
	}

	//Convert the runtime.Object to an unstructured Object
	unstruct := &unstructured.Unstructured{}
	err = scheme.Scheme.Convert(obj, unstruct, nil)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Converting to unstructured object")
	}

	dynClient := client.GetDynamicClient()
	mapper := client.GetMapper()

	gvk := unstruct.GroupVersionKind()
	mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Mapping kind to resource error")
	}

	gvr := mapping.Resource

	createdObj, err := dynClient.Resource(gvr).Namespace(namespace).Create(unstruct, metav1.CreateOptions{})
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create object error")
	}

	return createdObj.GetName(), nil
}

// Delete an existing deployment hosted in a specific Kubernetes cluster
func (g genericPlugin) Delete(kind string, name string, namespace string, client *app.KubernetesClient) error {
	if namespace == "" {
		namespace = "default"
	}

	deletePolicy := metav1.DeletePropagationForeground
	opts := &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	dynClient := client.GetDynamicClient()
	gvr, ok := kindToGVRMap[kind]
	if !ok {
		return pkgerrors.New("GVR not found for: " + kind)
	}

	log.Printf("Using gvr: %s, %s, %s", gvr.Group, gvr.Version, gvr.Resource)

	err := dynClient.Resource(gvr).Namespace(namespace).Delete(name, opts)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete object error")
	}

	return nil
}

// ExportedVariable is what we will look for when calling the generic plugin
var ExportedVariable genericPlugin
