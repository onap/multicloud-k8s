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
	"k8s.io/client-go/kubernetes"

	"k8s/krd"
)

func main() {}

// CreateResource object in a specific Kubernetes resource
func CreateResource(kubedata *krd.GenericKubeResourceData, kubeclient *kubernetes.Clientset) (string, error) {
	return "externalUUID", nil
}

// ListResources of existing resources
func ListResources(limit int64, namespace string, kubeclient *kubernetes.Clientset) (*[]string, error) {
	returnVal := []string{"cloud1-default-uuid1", "cloud1-default-uuid2"}
	return &returnVal, nil
}

// DeleteResource existing resources
func DeleteResource(name string, namespace string, kubeclient *kubernetes.Clientset) error {
	return nil
}

// GetResource existing resource host
func GetResource(namespace string, client *kubernetes.Clientset) (bool, error) {
	return true, nil
}
