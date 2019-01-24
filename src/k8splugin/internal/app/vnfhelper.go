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
	"encoding/hex"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"

	pkgerrors "github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"

	utils "k8splugin/internal"
	"k8splugin/internal/rb"
)

func generateExternalVNFID() string {
	b := make([]byte, 2)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func ensuresNamespace(namespace string, kubeclient kubernetes.Interface) error {
	namespacePlugin, ok := utils.LoadedPlugins["namespace"]
	if !ok {
		return pkgerrors.New("No plugin for namespace resource found")
	}

	symGetNamespaceFunc, err := namespacePlugin.Lookup("Get")
	if err != nil {
		return pkgerrors.Wrap(err, "Error fetching get namespace function")
	}

	ns, err := symGetNamespaceFunc.(func(string, string, kubernetes.Interface) (string, error))(
		namespace, namespace, kubeclient)
	if err != nil {
		return pkgerrors.Wrap(err, "An error ocurred during the get namespace execution")
	}

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
			namespaceResource, kubeclient)
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating "+namespace+" namespace")
		}
	}
	return nil
}

// CreateVNF reads the CSAR files from the files system and creates them one by one
var CreateVNF = func(csarID string, cloudRegionID string, rbProfileID string, kubeclient *kubernetes.Clientset) (string, map[string][]string, error) {

	overrideValues := []string{}
	rbProfileClient := rb.NewProfileClient()
	rbProfile, err := rbProfileClient.Get(rbProfileID)
	if err != nil {
		return "", nil, pkgerrors.Wrap(err, "Error getting profile info")
	}

	//Make sure that the namespace exists before trying to create any resources
	if err := ensuresNamespace(rbProfile.Namespace, kubeclient); err != nil {
		return "", nil, pkgerrors.Wrap(err, "Error while ensuring namespace: "+rbProfile.Namespace)
	}
	externalVNFID := generateExternalVNFID()
	internalVNFID := cloudRegionID + "-" + rbProfile.Namespace + "-" + externalVNFID

	metaMap, err := rb.NewProfileClient().Resolve(rbProfileID, overrideValues)
	if err != nil {
		return "", nil, pkgerrors.Wrap(err, "Error resolving helm charts")
	}

	resourceYAMLNameMap := make(map[string][]string)
	// Iterates over the resources defined in the metadata map to create kubernetes resources
	log.Printf("%d resource(s) type(s) to be processed", len(metaMap))
	for res, filePaths := range metaMap {
		//Convert resource to lower case as the map index is lowercase
		resource := strings.ToLower(res)
		log.Println("Processing items of " + string(resource) + " resource")
		var resourcesCreated []string
		for _, filepath := range filePaths {

			if _, err := os.Stat(filepath); os.IsNotExist(err) {
				return "", nil, pkgerrors.New("File " + filepath + "does not exists")
			}
			log.Println("Processing file: " + filepath)

			//Populate the namespace from profile instead of instance body
			genericKubeData := &utils.ResourceData{
				YamlFilePath: filepath,
				Namespace:    rbProfile.Namespace,
				VnfId:        internalVNFID,
			}

			typePlugin, ok := utils.LoadedPlugins[resource]
			if !ok {
				return "", nil, pkgerrors.New("No plugin for resource " + resource + " found")
			}

			symCreateResourceFunc, err := typePlugin.Lookup("Create")
			if err != nil {
				return "", nil, pkgerrors.Wrap(err, "Error fetching "+resource+" plugin")
			}

			internalResourceName, err := symCreateResourceFunc.(func(*utils.ResourceData, kubernetes.Interface) (string, error))(
				genericKubeData, kubeclient)
			if err != nil {
				return "", nil, pkgerrors.Wrap(err, "Error in plugin "+resource+" plugin")
			}
			log.Print(internalResourceName + " succesful resource created")
			resourcesCreated = append(resourcesCreated, internalResourceName)
		}
		resourceYAMLNameMap[resource] = resourcesCreated
	}

	return externalVNFID, resourceYAMLNameMap, nil
}

// DestroyVNF deletes VNFs based on data passed
var DestroyVNF = func(data map[string][]string, namespace string, kubeclient *kubernetes.Clientset) error {
	/* data:
	{
		"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
		"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
	},
	*/

	for resourceName, resourceList := range data {
		typePlugin, ok := utils.LoadedPlugins[resourceName]
		if !ok {
			return pkgerrors.New("No plugin for resource " + resourceName + " found")
		}

		symDeleteResourceFunc, err := typePlugin.Lookup("Delete")
		if err != nil {
			return pkgerrors.Wrap(err, "Error fetching "+resourceName+" plugin")
		}

		for _, resourceName := range resourceList {

			log.Println("Deleting resource: " + resourceName)

			err = symDeleteResourceFunc.(func(string, string, kubernetes.Interface) error)(
				resourceName, namespace, kubeclient)
			if err != nil {
				return pkgerrors.Wrap(err, "Error destroying "+resourceName)
			}
		}
	}

	return nil
}

// MetadataFile stores the metadata of execution
type MetadataFile struct {
	ResourceTypePathMap map[string][]string `yaml:"resources"`
}

// ReadMetadataFile reads the metadata yaml to return the order or reads
var ReadMetadataFile = func(path string) (MetadataFile, error) {
	var metadataFile MetadataFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return metadataFile, pkgerrors.Wrap(err, "Metadata YAML file does not exist")
	}

	log.Println("Reading metadata YAML: " + path)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return metadataFile, pkgerrors.Wrap(err, "Metadata YAML file read error")
	}

	err = yaml.Unmarshal(yamlFile, &metadataFile)
	if err != nil {
		return metadataFile, pkgerrors.Wrap(err, "Metadata YAML file unmarshal error")
	}
	log.Printf("metadata:\n%v", metadataFile)

	return metadataFile, nil
}
