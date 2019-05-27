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
	"strings"

	utils "k8splugin/internal"
	"k8splugin/internal/config"
	"k8splugin/internal/connection"
	"k8splugin/internal/helm"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// PluginReference is the interface that is implemented
type PluginReference interface {
	Create(yamlFilePath string, namespace string, client *KubernetesClient) (string, error)
	Delete(resource helm.KubernetesResource, namespace string, client *KubernetesClient) error
}

type KubernetesClient struct {
	clientSet      *kubernetes.Clientset
	dynamicClient  dynamic.Interface
	discoverClient *discovery.DiscoveryClient
	restMapper     meta.RESTMapper
        cloudRegion    string
}

// getKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the cloudregion. This is written out to a file.
func (k *KubernetesClient) getKubeConfig(cloudregion string) (string, error) {
	conn := connection.NewConnectionClient()
	kubeConfigPath, err := conn.Download(cloudregion, config.GetConfiguration().KubeConfigDir)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Downloading kubeconfig")
	}

	return kubeConfigPath, nil
}

// init loads the Kubernetes configuation values stored into the local configuration file
func (k *KubernetesClient) init(cloudregion string) error {
	if cloudregion == "" {
		return pkgerrors.New("Cloudregion is empty")
	}

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

	k.discoverClient, err = discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating discovery client")
	}
	k.cloudRegion = cloudregion
	return nil
}

func (k *KubernetesClient) ensureNamespace(namespace string) error {
	namespacePlugin, ok := utils.LoadedPlugins["namespace"]
	if !ok {
		return pkgerrors.New("No plugin for namespace resource found")
	}

	symGetNamespaceFunc, err := namespacePlugin.Lookup("Get")
	if err != nil {
		return pkgerrors.Wrap(err, "Error fetching get namespace function")
	}

	ns, _ := symGetNamespaceFunc.(func(string, string, *KubernetesClient) (string, error))(
		namespace, namespace, k)

	if ns == "" {
		log.Println("Creating " + namespace + " namespace")
		symGetNamespaceFunc, err := namespacePlugin.Lookup("Create")
		if err != nil {
			return pkgerrors.Wrap(err, "Error fetching create namespace plugin")
		}
		namespaceResource := &utils.ResourceData{
			Namespace: namespace,
		}

		_, err = symGetNamespaceFunc.(func(*utils.ResourceData, *KubernetesClient) (string, error))(
			namespaceResource, k)
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating "+namespace+" namespace")
		}
	}
	return nil
}

func (k *KubernetesClient) createGeneric(resTempl helm.KubernetesResourceTemplate,
	namespace string) (helm.KubernetesResource, error) {

	log.Println("Processing Kind: " + resTempl.GVK.Kind)

	//Check if have the mapper before loading the plugin
	err := k.updateMapper()
	if err != nil {
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Unable to create RESTMapper")
	}

	pluginObject, ok := utils.LoadedPlugins["generic"]
	if !ok {
		return helm.KubernetesResource{}, pkgerrors.New("No generic plugin found")
	}

	symbol, err := pluginObject.Lookup("ExportedVariable")
	if err != nil {
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "No ExportedVariable symbol found")
	}

	//Assert if it implements the PluginReference interface
	genericPlugin, ok := symbol.(PluginReference)
	if !ok {
		return helm.KubernetesResource{}, pkgerrors.New("ExportedVariable is not PluginReference type")
	}

	if _, err := os.Stat(resTempl.FilePath); os.IsNotExist(err) {
		return helm.KubernetesResource{}, pkgerrors.New("File " + resTempl.FilePath + "does not exists")
	}

	log.Println("Processing file: " + resTempl.FilePath)

	name, err := genericPlugin.Create(resTempl.FilePath, namespace, k)
	if err != nil {
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Error in generic plugin")
	}

	return helm.KubernetesResource{
		GVK:  resTempl.GVK,
		Name: name,
	}, nil
}

func (k *KubernetesClient) createKind(resTempl helm.KubernetesResourceTemplate,
	namespace string) (helm.KubernetesResource, error) {

	log.Println("Processing Kind: " + resTempl.GVK.Kind)

	if _, err := os.Stat(resTempl.FilePath); os.IsNotExist(err) {
		return helm.KubernetesResource{}, pkgerrors.New("File " + resTempl.FilePath + "does not exists")
	}

	log.Println("Processing file: " + resTempl.FilePath)

	//Populate the namespace from profile instead of instance body
	genericKubeData := &utils.ResourceData{
		YamlFilePath: resTempl.FilePath,
		Namespace:    namespace,
	}

	typePlugin, ok := utils.LoadedPlugins[strings.ToLower(resTempl.GVK.Kind)]
	if !ok {
		log.Println("No plugin for kind " + resTempl.GVK.Kind + " found. Using generic Plugin")
		return k.createGeneric(resTempl, namespace)
	}

	symCreateResourceFunc, err := typePlugin.Lookup("Create")
	if err != nil {
		return helm.KubernetesResource{}, pkgerrors.Wrap(err, "Error fetching "+resTempl.GVK.Kind+" plugin")
	}

	createdResourceName, err := symCreateResourceFunc.(func(*utils.ResourceData, *KubernetesClient) (string, error))(
		genericKubeData, k)
	if err != nil {
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

func (k *KubernetesClient) deleteGeneric(resource helm.KubernetesResource, namespace string) error {
	log.Println("Deleting Kind: " + resource.GVK.Kind)

	pluginObject, ok := utils.LoadedPlugins["generic"]
	if !ok {
		return pkgerrors.New("No generic plugin found")
	}

	//Check if have the mapper before loading the plugin
	err := k.updateMapper()
	if err != nil {
		return pkgerrors.Wrap(err, "Unable to create RESTMapper")
	}

	symbol, err := pluginObject.Lookup("ExportedVariable")
	if err != nil {
		return pkgerrors.Wrap(err, "No ExportedVariable symbol found")
	}

	//Assert that it implements the PluginReference interface
	genericPlugin, ok := symbol.(PluginReference)
	if !ok {
		return pkgerrors.New("ExportedVariable is not PluginReference type")
	}

	err = genericPlugin.Delete(resource, namespace, k)
	if err != nil {
		return pkgerrors.Wrap(err, "Error in generic plugin")
	}

	return nil
}

func (k *KubernetesClient) deleteKind(resource helm.KubernetesResource, namespace string) error {
	log.Println("Deleting Kind: " + resource.GVK.Kind)

	typePlugin, ok := utils.LoadedPlugins[strings.ToLower(resource.GVK.Kind)]
	if !ok {
		log.Println("No plugin for kind " + resource.GVK.Kind + " found. Using generic Plugin")
		return k.deleteGeneric(resource, namespace)
	}

	symDeleteResourceFunc, err := typePlugin.Lookup("Delete")
	if err != nil {
		return pkgerrors.Wrap(err, "Error finding Delete symbol in plugin")
	}

	log.Println("Deleting resource: " + resource.Name)
	err = symDeleteResourceFunc.(func(string, string, *KubernetesClient) error)(
		resource.Name, namespace, k)
	if err != nil {
		return pkgerrors.Wrap(err, "Error destroying "+resource.Name)
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

func (k *KubernetesClient) updateMapper() error {
	//Create restMapper if not already done
	if k.restMapper != nil {
		return nil
	}

	groupResources, err := restmapper.GetAPIGroupResources(k.discoverClient)
	if err != nil {
		return pkgerrors.Wrap(err, "Get GroupResources")
	}

	k.restMapper = restmapper.NewDiscoveryRESTMapper(groupResources)
	return nil
}

//GeClient returns the kubernetes client that was created for this client
func (k *KubernetesClient) GetClient() *kubernetes.Clientset {
	return k.clientSet
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
//GetCloudRegion returns the cloud region for this client
func (k *KubernetesClient) GetCloudRegion() string {
	return k.cloudRegion
}
