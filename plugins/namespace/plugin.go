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

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func main() {}

// CreateResource is used to create a new Namespace
func CreateResource(namespace string, client *kubernetes.Clientset) error {
	namespaceStruct := &coreV1.Namespace{
		ObjectMeta: metaV1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := client.CoreV1().Namespaces().Create(namespaceStruct)
	if err != nil {
		return pkgerrors.Wrap(err, "Create Namespace error")
	}
	return nil
}

// GetResource is used to check if a given namespace actually exists in Kubernetes
func GetResource(namespace string, client *kubernetes.Clientset) (bool, error) {
	opts := metaV1.ListOptions{}

	namespaceList, err := client.CoreV1().Namespaces().List(opts)
	if err != nil {
		return false, pkgerrors.Wrap(err, "Get Namespace list error")
	}

	for _, ns := range namespaceList.Items {
		if namespace == ns.Name {
			return true, nil
		}
	}

	return false, nil
}

// DeleteResource is used to delete a namespace
func DeleteResource(namespace string, client *kubernetes.Clientset) error {
	deletePolicy := metaV1.DeletePropagationForeground

	err := client.CoreV1().Namespaces().Delete(namespace, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil {
		return pkgerrors.Wrap(err, "Delete Namespace error")
	}
	return nil
}
