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

	utils "k8splugin/internal"
)

func main() {}

// Create object in a specific Kubernetes resource
func Create(data *utils.ResourceData, client kubernetes.Interface) (string, error) {
	return "externalUUID", nil
}

// List of existing resources
func List(namespace string, client kubernetes.Interface) ([]string, error) {
	returnVal := []string{"cloud1-default-uuid1", "cloud1-default-uuid2"}
	return returnVal, nil
}

// Delete existing resources
func Delete(name string, namespace string, client kubernetes.Interface) error {
	return nil
}

// Get existing resource host
func Get(name string, namespace string, client kubernetes.Interface) (string, error) {
	return name, nil
}
