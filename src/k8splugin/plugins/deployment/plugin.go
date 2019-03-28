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

	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	utils "k8splugin/internal"
)

// Create deployment object in a specific Kubernetes cluster
func Create(data *utils.ResourceData, client kubernetes.Interface) (string, error) {
	namespace := data.Namespace
	if namespace == "" {
		namespace = "default"
	}
	obj, err := utils.DecodeYAML(data.YamlFilePath, nil)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Decode deployment object error")
	}

	deployment, ok := obj.(*appsV1.Deployment)
	if !ok {
		return "", pkgerrors.New("Decoded object contains another resource different than Deployment")
	}
	deployment.Namespace = namespace
	result, err := client.AppsV1().Deployments(namespace).Create(deployment)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Deployment error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// List of existing deployments hosted in a specific Kubernetes cluster
func List(namespace string, kubeclient kubernetes.Interface) ([]string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: utils.ResourcesListLimit,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Deployment"

	list, err := kubeclient.AppsV1().Deployments(namespace).List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Deployment list error")
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

// Delete an existing deployment hosted in a specific Kubernetes cluster
func Delete(name string, namespace string, kubeclient kubernetes.Interface) error {
	if namespace == "" {
		namespace = "default"
	}

	deletePolicy := metaV1.DeletePropagationForeground
	opts := &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	log.Println("Deleting deployment: " + name)
	if err := kubeclient.AppsV1().Deployments(namespace).Delete(name, opts); err != nil {
		return pkgerrors.Wrap(err, "Delete Deployment error")
	}

	return nil
}

// Get an existing deployment hosted in a specific Kubernetes cluster
func Get(name string, namespace string, kubeclient kubernetes.Interface) (string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.GetOptions{}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Deployment"

	deployment, err := kubeclient.AppsV1().Deployments(namespace).Get(name, opts)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Get Deployment error")
	}

	return deployment.Name, nil
}
