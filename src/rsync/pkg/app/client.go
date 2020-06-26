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
	"os"
	"strings"
	"time"
	"encoding/base64"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"

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

// init loads the Kubernetes configuation values stored into the local configuration file
func (k *KubernetesClient) Init(clustername string, iid string) error {
	if clustername == "" {
		return pkgerrors.New("Cloudregion is empty")
	}

	if iid == "" {
		return pkgerrors.New("Instance ID is empty")
	}

	k.instanceID = iid

	configData, err := k.getKubeConfig(clustername, iid)
	if err != nil {
		return pkgerrors.Wrap(err, "Get kubeconfig file")
	}

	config, err := clientcmd.RESTConfigFromKubeConfig(configData)
	if err != nil {
		return pkgerrors.Wrap(err, "setConfig: Build config from flags raised an error")
	}

	k.clientSet, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	k.dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating dynamic client")
	}

	k.discoverClient, err = disk.NewCachedDiscoveryClientForConfig(config, os.TempDir(), "", 10*time.Minute)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating discovery client")
	}

	k.restMapper = restmapper.NewDeferredDiscoveryRESTMapper(k.discoverClient)

	return nil
}

//GetMapper returns the RESTMapper that was created for this client
func (k *KubernetesClient) GetMapper() meta.RESTMapper {
	return k.restMapper
}

//GetDynamicClient returns the dynamic client that is needed for
//unstructured REST calls to the apiserver
func (k *KubernetesClient) GetDynamicClient() dynamic.Interface {
	return k.dynamicClient
}

// GetStandardClient returns the standard client that can be used to handle
// standard kubernetes kinds
func (k *KubernetesClient) GetStandardClient() kubernetes.Interface {
	return k.clientSet
}

//GetInstanceID returns the instanceID that is injected into all the
//resources created by the plugin
func (k *KubernetesClient) GetInstanceID() string {
	return k.instanceID
}
