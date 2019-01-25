// +build integration

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
	"io/ioutil"
	"log"
	"os"
	"plugin"
	"testing"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/kubernetes"

	pkgerrors "github.com/pkg/errors"

	utils "k8splugin/internal"
)

func LoadMockPlugins(krdLoadedPlugins *map[string]*plugin.Plugin) error {
	if _, err := os.Stat("../../mock_files/mock_plugins/mockplugin.so"); os.IsNotExist(err) {
		return pkgerrors.New("mockplugin.so does not exist. Please compile mockplugin.go to generate")
	}

	mockPlugin, err := plugin.Open("../../mock_files/mock_plugins/mockplugin.so")
	if err != nil {
		return pkgerrors.Cause(err)
	}

	(*krdLoadedPlugins)["namespace"] = mockPlugin
	(*krdLoadedPlugins)["deployment"] = mockPlugin
	(*krdLoadedPlugins)["service"] = mockPlugin

	return nil
}

func TestCreateVNF(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins
	oldReadMetadataFile := ReadMetadataFile

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
		ReadMetadataFile = oldReadMetadataFile
	}()

	err := LoadMockPlugins(&utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("TestCreateVNF returned an error (%s)", err)
	}

	ReadMetadataFile = func(yamlFilePath string) (MetadataFile, error) {
		var seqFile MetadataFile

		if _, err := os.Stat(yamlFilePath); err == nil {
			rawBytes, err := ioutil.ReadFile("../../mock_files/mock_yamls/metadata.yaml")
			if err != nil {
				return seqFile, pkgerrors.Wrap(err, "Metadata YAML file read error")
			}

			err = yaml.Unmarshal(rawBytes, &seqFile)
			if err != nil {
				return seqFile, pkgerrors.Wrap(err, "Metadata YAML file unmarshall error")
			}
		}

		return seqFile, nil
	}

	kubeclient := kubernetes.Clientset{}

	t.Run("Successfully create VNF", func(t *testing.T) {
		externaluuid, data, err := CreateVNF("uuid", "cloudregion1", "test", &kubeclient)
		if err != nil {
			t.Fatalf("TestCreateVNF returned an error (%s)", err)
		}

		log.Println(externaluuid)

		if data == nil {
			t.Fatalf("TestCreateVNF returned empty data (%s)", data)
		}
	})

}

func TestDeleteVNF(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	err := LoadMockPlugins(&utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("TestCreateVNF returned an error (%s)", err)
	}

	kubeclient := kubernetes.Clientset{}

	t.Run("Successfully delete VNF", func(t *testing.T) {
		data := map[string][]string{
			"deployment": []string{"cloud1-default-uuid-sisedeploy"},
			"service":    []string{"cloud1-default-uuid-sisesvc"},
		}

		err := DestroyVNF(data, "test", &kubeclient)
		if err != nil {
			t.Fatalf("TestCreateVNF returned an error (%s)", err)
		}
	})
}

func TestReadMetadataFile(t *testing.T) {
	t.Run("Successfully read Metadata YAML file", func(t *testing.T) {
		_, err := ReadMetadataFile("../../mock_files//mock_yamls/metadata.yaml")
		if err != nil {
			t.Fatalf("TestReadMetadataFile returned an error (%s)", err)
		}
	})
}
