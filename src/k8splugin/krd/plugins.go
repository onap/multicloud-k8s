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

package krd

import (
	"plugin"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// LoadedPlugins stores references to the stored plugins
var LoadedPlugins = map[string]*plugin.Plugin{}

// KubeResourceClient has the signature methods to create Kubernetes reources
type KubeResourceClient interface {
	CreateResource(GenericKubeResourceData, *kubernetes.Clientset) (string, error)
	ListResources(string, string) (*[]string, error)
	DeleteResource(string, string, *kubernetes.Clientset) error
	GetResource(string, string, *kubernetes.Clientset) (string, error)
}

// GenericKubeResourceData stores all supported Kubernetes plugin types
type GenericKubeResourceData struct {
	YamlFilePath  string
	Namespace     string
	InternalVNFID string

	// Add additional Kubernetes plugins below kinds
	DeploymentData *appsV1.Deployment
	ServiceData    *coreV1.Service
}
