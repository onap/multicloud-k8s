/*
Copyright 2019 Intel Corporation.
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
	"reflect"
	"testing"

	utils "github.com/onap/multicloud-k8s/src/k8splugin/internal"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
)

func TestRegistryCreate(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins
	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()
	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	t.Run("Successfully create Registry", func(t *testing.T) {

		ic := NewRegistryClient()
		input := RegistryRequest{
			CloudOwner:    "INTEL",
			CloudRegion:   "RegionOne",
		}

		ir, err := ic.Create(input)
		if err != nil {
			t.Fatalf("TestRegistryCreate returned an error (%s)", err)
		}

		log.Println(ir)
	})

}


func TestRegistryGet(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	t.Run("Successfully Get Registry", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				RegistryKey{ID: "cocky_buck"}.String(): {
					"vim": []byte(
						`{
							"id": "cocky_buck",
							"request": {
								"cloud-owner": "INTEL",
								"cloud-region": "RegionOne"
							}
						}`),
				},
			},
		}

		expected := RegistryResponse{
			ID: "cocky_buck",
			Request: RegistryRequest{
				CloudOwner:  "INTEL",
				CloudRegion: "RegionOne",
			},
		}
		ic := NewRegistryClient()
		id := "cocky_buck"
		data, err := ic.Get(id)
		if err != nil {
			t.Fatalf("TestRegistryGet returned an error (%s)", err)
		}
		if !reflect.DeepEqual(expected, data) {
			t.Fatalf("TestRegistryGet returned:\n result=%v\n expected=%v",
				data, expected)
		}
	})

	t.Run("Get non-existing Registry", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				RegistryKey{ID: "cocky_buck"}.String(): {
					"vim": []byte(
						`{
							"id":"cocky_buck",
							"request": {
								"cloud-owner":"INTEL",
								"cloud-region":"RegionOne"
							}
						}`),
				},
			},
		}

		ic := NewRegistryClient()
		id := "non-existing"
		_, err := ic.Get(id)
		if err == nil {
			t.Fatal("Expected error, got pass", err)
		}
	})
}

func TestRegistryDelete(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("TestRegistryDelete returned an error (%s)", err)
	}

	t.Run("Successfully delete Registry", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				RegistryKey{ID: "cocky_buck"}.String(): {
					"vim": []byte(
						`{
							"id":"cocky_buck",
							"request": {
								"cloud-owner":"INTEL",
								"cloud-region":"RegionOne",
							}
						}`),
				},
			},
		}

		ic := NewRegistryClient()
		id := "cocky_buck"
		err := ic.Delete(id)
		if err != nil {
			t.Fatalf("TestRegistryDelete returned an error (%s)", err)
		}
	})

	t.Run("Delete non-existing Registry", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				RegistryKey{ID: "cocky_buck"}.String(): {
					"vim": []byte(
						`{
						  	"id":"cocky_buck",
							"request": {
								"cloudowner":"INTEL",
								"cloudregion":"RegionOne",
							}
						}`),
				},
			},
		}

		ic := NewRegistryClient()
		id := "non-existing"
		err := ic.Delete(id)
		if err != nil {
			t.Log("Expected error, got pass", err)
		}
	})
}
