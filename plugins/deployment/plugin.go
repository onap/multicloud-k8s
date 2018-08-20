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
	"io/ioutil"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"

	pkgerrors "github.com/pkg/errors"

	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s/krd"
)

func main() {}

// CreateResource object in a specific Kubernetes Deployment
func CreateResource(kubedata *krd.GenericKubeResourceData, kubeclient *kubernetes.Clientset) (string, error) {
	if kubedata.Namespace == "" {
		kubedata.Namespace = "default"
	}

	if _, err := os.Stat(kubedata.YamlFilePath); err != nil {
		return "", pkgerrors.New("File " + kubedata.YamlFilePath + " not found")
	}

	log.Println("Reading deployment YAML")
	rawBytes, err := ioutil.ReadFile(kubedata.YamlFilePath)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Deployment YAML file read error")
	}

	log.Println("Decoding deployment YAML")
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode(rawBytes, nil, nil)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Deserialize deployment error")
	}

	switch o := obj.(type) {
	case *appsV1.Deployment:
		kubedata.DeploymentData = o
	default:
		return "", pkgerrors.New(kubedata.YamlFilePath + " contains another resource different than Deployment")
	}

	kubedata.DeploymentData.Namespace = kubedata.Namespace
	kubedata.DeploymentData.Name = kubedata.InternalVNFID + "-" + kubedata.DeploymentData.Name

	result, err := kubeclient.AppsV1().Deployments(kubedata.Namespace).Create(kubedata.DeploymentData)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Create Deployment error")
	}

	return result.GetObjectMeta().GetName(), nil
}

// ListResources of existing deployments hosted in a specific Kubernetes Deployment
func ListResources(limit int64, namespace string, kubeclient *kubernetes.Clientset) (*[]string, error) {
	if namespace == "" {
		namespace = "default"
	}

	opts := metaV1.ListOptions{
		Limit: limit,
	}
	opts.APIVersion = "apps/v1"
	opts.Kind = "Deployment"

	list, err := kubeclient.AppsV1().Deployments(namespace).List(opts)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Deployment list error")
	}

	result := make([]string, 0, limit)
	if list != nil {
		for _, deployment := range list.Items {
			result = append(result, deployment.Name)
		}
	}

	return &result, nil
}

// DeleteResource existing deployments hosting in a specific Kubernetes Deployment
func DeleteResource(name string, namespace string, kubeclient *kubernetes.Clientset) error {
	if namespace == "" {
		namespace = "default"
	}

	log.Println("Deleting deployment: " + name)

	deletePolicy := metaV1.DeletePropagationForeground
	err := kubeclient.AppsV1().Deployments(namespace).Delete(name, &metaV1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	})

	if err != nil {
		return pkgerrors.Wrap(err, "Delete Deployment error")
	}

	return nil
}

// GetResource existing deployment hosting in a specific Kubernetes Deployment
func GetResource(name string, namespace string, kubeclient *kubernetes.Clientset) (string, error) {
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
