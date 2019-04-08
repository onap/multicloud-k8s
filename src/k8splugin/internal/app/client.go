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

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/helm/pkg/tiller"
)

// KubernetesResource is the interface that is implemented
type KubernetesResource interface {
	Create(yamlFilePath string, namespace string, client *KubernetesClient) (string, error)
	Delete(kind string, name string, namespace string, client *KubernetesClient) error
}

type KubernetesClient struct {
	clientSet      *kubernetes.Clientset
	dynamicClient  dynamic.Interface
	discoverClient *discovery.DiscoveryClient
	restMapper     meta.RESTMapper
}

// GetKubeClient loads the Kubernetes configuation values stored into the local configuration file
func (k *KubernetesClient) init(configPath string) error {
	if configPath == "" {
		return pkgerrors.New("config not passed and is not found in ~/.kube. ")
	}

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

	ns, _ := symGetNamespaceFunc.(func(string, string, kubernetes.Interface) (string, error))(
		namespace, namespace, k.clientSet)

	if ns == "" {
		log.Println("Creating " + namespace + " namespace")
		symGetNamespaceFunc, err := namespacePlugin.Lookup("Create")
		if err != nil {
			return pkgerrors.Wrap(err, "Error fetching create namespace plugin")
		}
		namespaceResource := &utils.ResourceData{
			Namespace: namespace,
		}

		_, err = symGetNamespaceFunc.(func(*utils.ResourceData, kubernetes.Interface) (string, error))(
			namespaceResource, k.clientSet)
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating "+namespace+" namespace")
		}
	}
	return nil
}

func (k *KubernetesClient) createGeneric(kind string, files []string, namespace string) ([]string, error) {

	log.Println("Processing items of Kind: " + kind)

	//Check if have the mapper before loading the plugin
	err := k.updateMapper()
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Unable to create RESTMapper")
	}

	pluginObject, ok := utils.LoadedPlugins["generic"]
	if !ok {
		return nil, pkgerrors.New("No generic plugin found")
	}

	symbol, err := pluginObject.Lookup("ExportedVariable")
	if err != nil {
		return nil, pkgerrors.Wrap(err, "No ExportedVariable symbol found")
	}

	genericPlugin, ok := symbol.(KubernetesResource)
	if !ok {
		return nil, pkgerrors.New("ExportedVariable is not KubernetesResource type")
	}

	//Iterate over each file of a particular kind here
	var resourcesCreated []string
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return nil, pkgerrors.New("File " + f + "does not exists")
		}

		log.Println("Processing file: " + f)

		name, err := genericPlugin.Create(f, namespace, k)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Error in generic plugin")
		}

		resourcesCreated = append(resourcesCreated, name)
	}
	return resourcesCreated, nil
}

func (k *KubernetesClient) createKind(kind string, files []string, namespace string) ([]string, error) {

	log.Println("Processing items of Kind: " + kind)

	//Iterate over each file of a particular kind here
	var resourcesCreated []string
	for _, f := range files {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			return nil, pkgerrors.New("File " + f + "does not exists")
		}

		log.Println("Processing file: " + f)

		//Populate the namespace from profile instead of instance body
		genericKubeData := &utils.ResourceData{
			YamlFilePath: f,
			Namespace:    namespace,
		}

		typePlugin, ok := utils.LoadedPlugins[strings.ToLower(kind)]
		if !ok {
			log.Println("No plugin for kind " + kind + " found. Using generic Plugin")
			return k.createGeneric(kind, files, namespace)
		}

		symCreateResourceFunc, err := typePlugin.Lookup("Create")
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Error fetching "+kind+" plugin")
		}

		createdResourceName, err := symCreateResourceFunc.(func(*utils.ResourceData, kubernetes.Interface) (string, error))(
			genericKubeData, k.clientSet)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Error in plugin "+kind+" plugin")
		}
		log.Print(createdResourceName + " created")
		resourcesCreated = append(resourcesCreated, createdResourceName)
	}

	return resourcesCreated, nil
}

func (k *KubernetesClient) createResources(resMap map[string][]string,
	namespace string) (map[string][]string, error) {

	err := k.ensureNamespace(namespace)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Creating Namespace")
	}

	createdResourceMap := make(map[string][]string)
	// Create all the known kinds in the InstallOrder
	for _, kind := range tiller.InstallOrder {
		files, ok := resMap[kind]
		if !ok {
			log.Println("Kind " + kind + " not found. Skipping...")
			continue
		}

		resourcesCreated, err := k.createKind(kind, files, namespace)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Error creating kind: "+kind)
		}

		createdResourceMap[kind] = resourcesCreated
		delete(resMap, kind)
	}

	//Create the remaining kinds from the resMap
	for kind, files := range resMap {
		resourcesCreated, err := k.createKind(kind, files, namespace)
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Error creating kind: "+kind)
		}

		createdResourceMap[kind] = resourcesCreated
		delete(resMap, kind)
	}

	return createdResourceMap, nil
}

func (k *KubernetesClient) deleteGeneric(kind string, resources []string, namespace string) error {
	log.Println("Deleting items of Kind: " + kind)

	pluginObject, ok := utils.LoadedPlugins["generic"]
	if !ok {
		return pkgerrors.New("No generic plugin found")
	}

	symbol, err := pluginObject.Lookup("ExportedVariable")
	if err != nil {
		return pkgerrors.Wrap(err, "No ExportedVariable symbol found")
	}

	//Assert that it implements the KubernetesResource
	genericPlugin, ok := symbol.(KubernetesResource)
	if !ok {
		return pkgerrors.New("ExportedVariable is not KubernetesResource type")
	}

	for _, res := range resources {
		err = genericPlugin.Delete(kind, res, namespace, k)
		if err != nil {
			return pkgerrors.Wrap(err, "Error in generic plugin")
		}
	}

	return nil
}

func (k *KubernetesClient) deleteKind(kind string, resources []string, namespace string) error {
	log.Println("Deleting items of Kind: " + kind)

	typePlugin, ok := utils.LoadedPlugins[strings.ToLower(kind)]
	if !ok {
		return pkgerrors.New("No plugin for resource " + kind + " found")
	}

	symDeleteResourceFunc, err := typePlugin.Lookup("Delete")
	if err != nil {
		log.Println("No plugin for kind " + kind + " found. Using generic Plugin")
		return k.deleteGeneric(kind, resources, namespace)
	}

	for _, res := range resources {
		log.Println("Deleting resource: " + res)
		err = symDeleteResourceFunc.(func(string, string, kubernetes.Interface) error)(
			res, namespace, k.clientSet)
		if err != nil {
			return pkgerrors.Wrap(err, "Error destroying "+res)
		}
	}
	return nil
}

func (k *KubernetesClient) deleteResources(resMap map[string][]string, namespace string) error {
	//TODO: Investigate if deletion should be in a particular order
	for kind, resourceNames := range resMap {
		err := k.deleteKind(kind, resourceNames, namespace)
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

//GetMapper returns the RESTMapper that was created for this client
func (k *KubernetesClient) GetMapper() meta.RESTMapper {
	return k.restMapper
}

//GetDynamicClient returns the dynamic client that is needed for
//unstructured REST calls to the apiserver
func (k *KubernetesClient) GetDynamicClient() dynamic.Interface {
	return k.dynamicClient
}
