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
	"errors"

	pkgerrors "github.com/pkg/errors"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeClient loads the Kubernetes configuation values stored into the local configuration file
var GetKubeClient = func(configPath string) (kubernetes.Clientset, error) {
	var clientset *kubernetes.Clientset

	if configPath == "" {
		return *clientset, errors.New("config not passed and is not found in ~/.kube. ")
	}

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
	if err != nil {
		return *clientset, pkgerrors.Wrap(err, "setConfig: Build config from flags raised an error")
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return *clientset, err
	}

	return *clientset, nil
}
