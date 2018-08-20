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

package csar

import (
	"io/ioutil"
	"log"
	"os"

	"k8s.io/client-go/kubernetes"

	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"k8s.io/apimachinery/pkg/util/uuid"

	"k8-plugin-multicloud/krd"
)

// CreateVNF reads the CSAR files from the files system and creates them one by one
var CreateVNF = func(csarID string, cloudRegionID string, namespace string, kubeclient *kubernetes.Clientset) (string, map[string][]string, error) {
	namespacePlugin, ok := krd.LoadedPlugins["namespace"]
	if !ok {
		return "", nil, pkgerrors.New("No plugin for namespace resource found")
	}

	symGetNamespaceFunc, err := namespacePlugin.Lookup("GetResource")
	if err != nil {
		return "", nil, pkgerrors.Wrap(err, "Error fetching namespace plugin")
	}

	present, err := symGetNamespaceFunc.(func(string, *kubernetes.Clientset) (bool, error))(
		namespace, kubeclient)
	if err != nil {
		return "", nil, pkgerrors.Wrap(err, "Error in plugin namespace plugin")
	}

	if present == false {
		symGetNamespaceFunc, err := namespacePlugin.Lookup("CreateResource")
		if err != nil {
			return "", nil, pkgerrors.Wrap(err, "Error fetching namespace plugin")
		}

		err = symGetNamespaceFunc.(func(string, *kubernetes.Clientset) error)(
			namespace, kubeclient)
		if err != nil {
			return "", nil, pkgerrors.Wrap(err, "Error creating "+namespace+" namespace")
		}
	}

	var path string

	// uuid
	externalVNFID := string(uuid.NewUUID())

	// cloud1-default-uuid
	internalVNFID := cloudRegionID + "-" + namespace + "-" + externalVNFID

	csarDirPath := os.Getenv("CSAR_DIR") + "/" + csarID
	metadataYAMLPath := csarDirPath + "/metadata.yaml"

	seqFile, err := ReadMetadataFile(metadataYAMLPath)
	if err != nil {
		return "", nil, pkgerrors.Wrap(err, "Error while reading Metadata File: "+metadataYAMLPath)
	}

	resourceYAMLNameMap := make(map[string][]string)

	for _, resource := range seqFile.ResourceTypePathMap {
		for resourceName, resourceFileNames := range resource {
			// Load/Use Deployment data/client

			var resourceNameList []string

			for _, filename := range resourceFileNames {
				path = csarDirPath + "/" + filename

				_, err = os.Stat(path)
				if os.IsNotExist(err) {
					return "", nil, pkgerrors.New("File " + path + "does not exists")
				}

				log.Println("Processing file: " + path)

				genericKubeData := &krd.GenericKubeResourceData{
					YamlFilePath:  path,
					Namespace:     namespace,
					InternalVNFID: internalVNFID,
				}

				typePlugin, ok := krd.LoadedPlugins[resourceName]
				if !ok {
					return "", nil, pkgerrors.New("No plugin for resource " + resourceName + " found")
				}

				symCreateResourceFunc, err := typePlugin.Lookup("CreateResource")
				if err != nil {
					return "", nil, pkgerrors.Wrap(err, "Error fetching "+resourceName+" plugin")
				}

				// cloud1-default-uuid-sisedeploy
				internalResourceName, err := symCreateResourceFunc.(func(*krd.GenericKubeResourceData, *kubernetes.Clientset) (string, error))(
					genericKubeData, kubeclient)
				if err != nil {
					return "", nil, pkgerrors.Wrap(err, "Error in plugin "+resourceName+" plugin")
				}

				// ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
				resourceNameList = append(resourceNameList, internalResourceName)

				/*
					{
						"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
					}
				*/
				resourceYAMLNameMap[resourceName] = resourceNameList
			}
		}
	}

	/*
		uuid,
		{
			"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
			"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
		},
		nil
	*/
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
		typePlugin, ok := krd.LoadedPlugins[resourceName]
		if !ok {
			return pkgerrors.New("No plugin for resource " + resourceName + " found")
		}

		symDeleteResourceFunc, err := typePlugin.Lookup("DeleteResource")
		if err != nil {
			return pkgerrors.Wrap(err, "Error fetching "+resourceName+" plugin")
		}

		for _, resourceName := range resourceList {

			log.Println("Deleting resource: " + resourceName)

			err = symDeleteResourceFunc.(func(string, string, *kubernetes.Clientset) error)(
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
	ResourceTypePathMap []map[string][]string `yaml:"resources"`
}

// ReadMetadataFile reads the metadata yaml to return the order or reads
var ReadMetadataFile = func(yamlFilePath string) (MetadataFile, error) {
	var seqFile MetadataFile

	if _, err := os.Stat(yamlFilePath); err == nil {
		log.Println("Reading metadata YAML: " + yamlFilePath)
		rawBytes, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return seqFile, pkgerrors.Wrap(err, "Metadata YAML file read error")
		}

		err = yaml.Unmarshal(rawBytes, &seqFile)
		if err != nil {
			return seqFile, pkgerrors.Wrap(err, "Metadata YAML file read error")
		}
	}

	return seqFile, nil
}
