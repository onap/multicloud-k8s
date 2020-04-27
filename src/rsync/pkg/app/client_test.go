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
	"io/ioutil"
	"os"
	"plugin"
	"reflect"
	"testing"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

func LoadMockPlugins(krdLoadedPlugins map[string]*plugin.Plugin) error {
	if _, err := os.Stat("../../mock_files/mock_plugins/mockplugin.so"); os.IsNotExist(err) {
		return pkgerrors.New("mockplugin.so does not exist. Please compile mockplugin.go to generate")
	}

	mockPlugin, err := plugin.Open("../../mock_files/mock_plugins/mockplugin.so")
	if err != nil {
		return pkgerrors.Wrap(err, "Opening mock plugins")
	}

	krdLoadedPlugins["namespace"] = mockPlugin
	krdLoadedPlugins["generic"] = mockPlugin
	krdLoadedPlugins["service"] = mockPlugin

	return nil
}

func TestInit(t *testing.T) {
	t.Run("Successfully create Kube Client", func(t *testing.T) {
		// Load the mock kube config file into memory
		fd, err := ioutil.ReadFile("../../mock_files/mock_configs/mock_kube_config")
		if err != nil {
			t.Fatal("Unable to read mock_kube_config")
		}

		fdbase64 := base64.StdEncoding.EncodeToString(fd)

		// Create mock db with connectivity information in it
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				connection.ConnectionKey{CloudRegion: "mock_connection"}.String(): {
					"metadata": []byte(
						"{\"cloud-region\":\"mock_connection\"," +
							"\"cloud-owner\":\"mock_owner\"," +
							"\"kubeconfig\": \"" + fdbase64 + "\"}"),
				},
			},
		}

		kubeClient := KubernetesClient{}
		// Refer to the connection via its name
		err = kubeClient.init("mock_connection", "abcdefg")
		if err != nil {
			t.Fatalf("TestGetKubeClient returned an error (%s)", err)
		}

		name := reflect.TypeOf(kubeClient.clientSet).Elem().Name()
		if name != "Clientset" {
			t.Fatalf("TestGetKubeClient returned :\n result=%v\n expected=%v", name, "Clientset")
		}

	})
}

func TestCreateResources(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	k8 := KubernetesClient{
		clientSet: &kubernetes.Clientset{},
	}

	t.Run("Successfully delete resources", func(t *testing.T) {
		data := []helm.KubernetesResourceTemplate{
			{
				GVK: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment"},
				FilePath: "../../mock_files/mock_yamls/deployment.yaml",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Service"},
				FilePath: "../../mock_files/mock_yamls/service.yaml",
			},
		}

		_, err := k8.createResources(data, "testnamespace")
		if err != nil {
			t.Fatalf("TestCreateResources returned an error (%s)", err)
		}
	})
}

func TestDeleteResources(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	k8 := KubernetesClient{
		clientSet: &kubernetes.Clientset{},
	}

	t.Run("Successfully delete resources", func(t *testing.T) {
		data := []helm.KubernetesResource{
			{
				GVK: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment"},
				Name: "deployment-1",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment"},
				Name: "deployment-2",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Service"},
				Name: "service-1",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Service"},
				Name: "service-2",
			},
		}

		err := k8.deleteResources(data, "test")
		if err != nil {
			t.Fatalf("TestCreateVNF returned an error (%s)", err)
		}
	})
}
