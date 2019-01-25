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

// Create a namespace object in a specific Kubernetes cluster
func Create(data *utils.ResourceData, client kubernetes.Interface) (string, error) {
	namespace := &coreV1.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name: data.Namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(namespace)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Namespace error")
	}
	log.Printf("Namespace (%s) created", data.Namespace)

	return data.Namespace, nil
}

// Get an existing namespace hosted in a specific Kubernetes cluster
func Get(name string, namespace string, client kubernetes.Interface) (string, error) {
	opts := metaV1.GetOptions{}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Deployment"

	ns, err := client.CoreV1().Namespaces().Get(name, opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get Namespace error")
	}

	return ns.Name, nil
}

// Delete an existing namespace hosted in a specific Kubernetes cluster
func Delete(name string, namespace string, client kubernetes.Interface) error {
	deletePolicy := metaV1.DeletePropagationForeground
	opts := &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	log.Println("Deleting namespace: " + name)
	if err := client.CoreV1().Namespaces().Delete(name, opts); err != nil {
		return pkgerrors.Wrap(err, "Delete namespace error")
	}

	return nil
}

// List of existing namespaces hosted in a specific Kubernetes cluster
func List(namespace string, client kubernetes.Interface) ([]string, error) {
	opts := metaV1.ListOptions{
		Limit: utils.ResourcesListLimit,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Namespace"

	list, err := client.CoreV1().Namespaces().List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Namespace list error")
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
