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
	"log"
	"os"
	"time"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/plugin"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
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
// of the cloudregion. This is written out to a file.
func (k *KubernetesClient) getKubeConfig(cloudregion string) (string, error) {

	conn := connection.NewConnectionClient()
	kubeConfigPath, err := conn.Download(cloudregion)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Downloading kubeconfig")
	}

	return kubeConfigPath, nil
}

// init loads the Kubernetes configuation values stored into the local configuration file
func (k *KubernetesClient) init(cloudregion string, iid string) error {
	if cloudregion == "" {
		return pkgerrors.New("Cloudregion is empty")
	}

	if iid == "" {
		return pkgerrors.New("Instance ID is empty")
	}

	k.instanceID = iid

	configPath, err := k.getKubeConfig(cloudregion)
	if err != nil {
		return pkgerrors.Wrap(err, "Get kubeconfig file")
	}

	//Remove kubeconfigfile after the clients are created
	defer os.Remove(configPath)

	config, err := clientcmd.BuildConfigFromFlags("", configPath)
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

func (k *KubernetesClient) ensureNamespace(namespace string) error {

	pluginImpl, err := plugin.GetPluginByKind("Namespace")
	if err != nil {
		return pkgerrors.Wrap(err, "Loading Namespace Plugin")
	}

	ns, err := pluginImpl.Get(helm.KubernetesResource{
		Name: namespace,
		GVK: schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Namespace",
		},
	}, namespace, k)

	if ns == "" {
		log.Println("Creating " + namespace + " namespace")

		_, err = pluginImpl.Create("", namespace, k)
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating "+namespace+" namespace")
		}
	}
	return nil
}

func (k *KubernetesClient) createKind(resTempl helm.KubernetesResourceTemplate,
	namespace string) (helm.KubernetesResource, error) {

	if _, err := os.Stat(resTempl.FilePath); os.IsNotExist(err) {
		return helm.KubernetesResource{}, pkgerrors.New("File " + resTempl.FilePath + "does not exists")
	}

	log.Println("Processing file: " + resTempl.FilePath)

	pluginImpl, err := plugin.GetPluginByKind(resTempl.GVK.Kind)
	if err != nil {
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Error loading plugin")
	}

	createdResourceName, err := pluginImpl.Create(resTempl.FilePath, namespace, k)
	if err != nil {
		log.Printf("Error: %s while creating: %s", err.Error(), resTempl.GVK.Kind)
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Error in plugin "+resTempl.GVK.Kind+" plugin")
	}

	log.Print(createdResourceName + " created")
	return helm.KubernetesResource{
		GVK:  resTempl.GVK,
		Name: createdResourceName,
	}, nil
}

func (k *KubernetesClient) createResources(sortedTemplates []helm.KubernetesResourceTemplate,
	namespace string) ([]helm.KubernetesResource, error) {

	err := k.ensureNamespace(namespace)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Creating Namespace")
	}

	var createdResources []helm.KubernetesResource
	for _, resTempl := range sortedTemplates {
		resCreated, err := k.createKind(resTempl, namespace)
		if err != nil {
			return nil, pkgerrors.Wrapf(err, "Error creating kind: %+v", resTempl.GVK)
		}
		createdResources = append(createdResources, resCreated)
	}

	return createdResources, nil
}

func (k *KubernetesClient) deleteKind(resource helm.KubernetesResource, namespace string) error {
	log.Println("Deleting Kind: " + resource.GVK.Kind)

	pluginImpl, err := plugin.GetPluginByKind(resource.GVK.Kind)
	if err != nil {
		return pkgerrors.Wrap(err, "Error loading plugin")
	}

	log.Println("Deleting resource: " + resource.Name)
	err = pluginImpl.Delete(resource, namespace, k)
	if err != nil {
		return pkgerrors.Wrap(err, "Error deleting "+resource.Name)
	}

	return nil
}

func (k *KubernetesClient) deleteResources(resources []helm.KubernetesResource, namespace string) error {
	//TODO: Investigate if deletion should be in a particular order
	for _, res := range resources {
		err := k.deleteKind(res, namespace)
		if err != nil {
			return pkgerrors.Wrap(err, "Deleting resources")
		}
	}

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
