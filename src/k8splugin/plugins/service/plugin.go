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

	"k8s.io/client-go/kubernetes"

	pkgerrors "github.com/pkg/errors"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	utils "k8splugin/internal"
)

// Create a service object in a specific Kubernetes cluster
func Create(data *utils.ResourceData, client kubernetes.Interface) (string, error) {
	namespace := data.Namespace
	if namespace == "" {
		namespace = "default"
	}
	obj, err := utils.DecodeYAML(data.YamlFilePath, nil)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Decode service object error")
	}

	service, ok := obj.(*coreV1.Service)
	if !ok {
		return "", pkgerrors.New("Decoded object contains another resource different than Service")
	}
	service.Namespace = namespace

	result, err := client.CoreV1().Services(namespace).Create(service)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Service error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// List of existing services hosted in a specific Kubernetes cluster
func List(namespace string, kubeclient kubernetes.Interface) ([]string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: utils.ResourcesListLimit,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Service"

	list, err := kubeclient.CoreV1().Services(namespace).List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Service list error")
	}

	result := make([]string, 0, utils.ResourcesListLimit)
	if list != nil {
		for _, deployment := range list.Items {
			log.Printf("%v", deployment.Name)
			result = append(result, deployment.Name)
		}
	}

	return result, nil
}

// Delete an existing service hosted in a specific Kubernetes cluster
func Delete(name string, namespace string, kubeclient kubernetes.Interface) error {
	if namespace == "" {
		namespace = "default"
	}

	deletePolicy := metaV1.DeletePropagationForeground
	opts := &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	log.Println("Deleting service: " + name)
	if err := kubeclient.CoreV1().Services(namespace).Delete(name, opts); err != nil {
		return pkgerrors.Wrap(err, "Delete service error")
	}

	return nil
}

// Get an existing service hosted in a specific Kubernetes cluster
func Get(name string, namespace string, kubeclient kubernetes.Interface) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.GetOptions{}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Service"

	service, err := kubeclient.CoreV1().Services(namespace).Get(name, opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get Deployment error")
	}

	return service.Name, nil
}
