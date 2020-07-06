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

package app

import (
	"encoding/base64"
	"os"
	"strings"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/onap/multicloud-k8s/src/clm/pkg/cluster"
)

// KubernetesClient encapsulates the different clients' interfaces
// we need when interacting with a Kubernetes cluster
type KubernetesClient struct {
	clientSet      kubernetes.Interface
	dynamicClient  dynamic.Interface
	discoverClient *disk.CachedDiscoveryClient
	restMapper     meta.RESTMapper
	instanceID     string
}

// getKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the clustername. This is written out to a file.
func (k *KubernetesClient) getKubeConfig(clustername string, id string) ([]byte, error) {

	if !strings.Contains(clustername, "+") {
		return nil, pkgerrors.New("Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		return nil, pkgerrors.New("Not a valid cluster name")
	}
	kubeConfig, err := cluster.NewClusterClient().GetClusterContent(strs[0], strs[1])
	if err != nil {
		return nil, pkgerrors.New("Get kubeconfig failed")
	}

	dec, err := base64.StdEncoding.DecodeString(kubeConfig.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

const basePath string = "/tmp/rsync/"

func (k *KubernetesClient) GetKubeConfigFile(clustername string, id string) (string, error) {

	if !strings.Contains(clustername, "+") {
		return "", pkgerrors.New("Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		return "", pkgerrors.New("Not a valid cluster name")
	}
	kubeConfig, err := cluster.NewClusterClient().GetClusterContent(strs[0], strs[1])
	if err != nil {
		return "", pkgerrors.New("Get kubeconfig failed")
	}

	var kubeConfigPath string = basePath + id + "/" + clustername + "/"

	if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
		err = os.MkdirAll(kubeConfigPath, 0755)
		if err != nil {
			return "", err
		}
	}
	kubeConfigPath = kubeConfigPath + "config"

	f, err := os.Create(kubeConfigPath)
	defer f.Close()
	if err != nil {
		return "", err
	}
	dec, err := base64.StdEncoding.DecodeString(kubeConfig.Kubeconfig)
	if err != nil {
		return "", err
	}
	_, err = f.Write(dec)
	if err != nil {
		return "", err
	}

	return kubeConfigPath, nil
}
