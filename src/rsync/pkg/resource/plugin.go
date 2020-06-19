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

package plugin

import (
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	utils "github.com/onap/multicloud-k8s/src/rsync/pkg/internal"
	"github.com/onap/multicloud-k8s/src/rsync/pkg/internal/config"
	"github.com/onap/multicloud-k8s/src/rsync/pkg/plugin"
)

// Compile time check to see if GenericPlugin implements the correct interface
var _ plugin.Reference = GenericPlugin{}

// ExportedVariable is what we will look for when calling the generic plugin
var ExportedVariable GenericPlugin

type GenericPlugin struct {
}

// Create deployment object in a specific Kubernetes cluster
func (g GenericPlugin) Create(yamlFilePath string, namespace string, client plugin.KubernetesConnector) (string, error) {
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

// Delete an existing resource hosted in a specific Kubernetes cluster
func (g GenericPlugin) Delete(yamlFilePath string, resname string, namespace string, client plugin.KubernetesConnector) error {
        if namespace == "" {
                namespace = "default"
        }

        //Decode the yaml file to create a runtime.Object
        unstruct := &unstructured.Unstructured{}
        //Ignore the returned obj as we expect the data in unstruct
        _, err := utils.DecodeYAML(yamlFilePath, unstruct)
        if err != nil {
                return pkgerrors.Wrap(err, "Decode deployment object error")
        }

        dynClient := client.GetDynamicClient()
        mapper := client.GetMapper()

        gvk := unstruct.GroupVersionKind()
        mapping, err := mapper.RESTMapping(schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}, gvk.Version)
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
                err = dynClient.Resource(gvr).Namespace(namespace).Delete(resname, opts)
        case meta.RESTScopeNameRoot:
                err = dynClient.Resource(gvr).Delete(resname, opts)
        default:
                return pkgerrors.New("Got an unknown RESTSCopeName for mappin")
        }

        if err != nil {
                return pkgerrors.Wrap(err, "Delete object error")
        }
        return nil
}
