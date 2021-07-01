/*
Copyright 2018 Intel Corporation.
Copyright © 2021 Nokia Bell Labs.
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
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/utils"
	"io/ioutil"
	"log"
	"reflect"
	"sort"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestInstanceCreate(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins
	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()
	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	// Load the mock kube config file into memory
	fd, err := ioutil.ReadFile("../../mock_files/mock_configs/mock_kube_config")
	if err != nil {
		t.Fatal("Unable to read mock_kube_config")
	}

	t.Run("Successfully create Instance", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				rb.ProfileKey{RBName: "test-rbdef", RBVersion: "v1",
					ProfileName: "profile1"}.String(): {
					"profilemetadata": []byte(
						"{\"profile-name\":\"profile1\"," +
							"\"release-name\":\"testprofilereleasename\"," +
							"\"namespace\":\"testnamespace\"," +
							"\"rb-name\":\"test-rbdef\"," +
							"\"rb-version\":\"v1\"," +
							"\"kubernetesversion\":\"1.12.3\"}"),
					// base64 encoding of vagrant/tests/vnfs/testrb/helm/profile
					"profilecontent": []byte("H4sICLmjT1wAA3Byb2ZpbGUudGFyAO1Y32/bNhD2s/6Kg/KyYZZsy" +
						"78K78lLMsxY5gRxmqIYhoKWaJsYJWokZdfo+r/vSFmunCZNBtQJ1vF7sXX36e54vDN5T" +
						"knGFlTpcEtS3jgO2ohBr2c/EXc/29Gg1+h0e1F32Ol1B1Gj3Ymifr8B7SPFc4BCaSIBG" +
						"lII/SXeY/r/KIIg8NZUKiayEaw7nt7mdOQBrAkvqBqBL1ArWULflRJbJz4SYpEt2FJSJ" +
						"QoZ21cAAlgwTnOiVyPQWFQLwVuqmCdMthKac7FNaVZWmqWjkRWRuuSvScF1gFZVwYOEr" +
						"luapjknaOazd186Z98S7tver+3j0f5v1/q/18f+7w56bdf/zwFF5ZqV/WtbH6YioVdCa" +
						"hRkJEVBVSFBvUNRmyNpesgwors0lmkqM8KNzRG8iqLIWN45GUGv57l+fkFUP9PH9GF6f" +
						"IgH+kP9b76b/o+GUb9r5J1O1I0a0D9mUBX+5/1/55g+io9/sf+DnuF1sA4Gbv+fA1++p" +
						"n0dH4+c/92oPaztv+n/fn84dOf/c+AETkW+lWy50hC1O69gguc1R6HEw5xoHAuaKIq9E" +
						"+8ELvCikCmaQJElVIJeURjnJMaPnaYJt+UoAVHYhu8Mwd+p/O9/RAtbUUBKtnj+aygUR" +
						"RNM2ZkB6PuY5hpvCzhY4L2fkSymsGF6Zd3sjIRo4u3OhJhrgmyC/ByfFnUeEG0DLrHSO" +
						"h+1WpvNJiQ23FDIZYuXVNW6mJyeT2fnAYZsX3qdcaoUSPpXwSQudr4FkmNEMZljnJxsQ" +
						"EggOPmgTgsT8UYyzbJlE5RY6A2RFK0kTGnJ5oU+SFcVH666TsCEkQz88QwmMx9+Gs8ms" +
						"ybaeDO5+eXy9Q28GV9fj6c3k/MZXF7D6eX0bHIzuZzi088wnr6FXyfTsyZQTBa6oe9za" +
						"eLHIJlJJE1M1maUHgSwEGVAKqcxW7AY15UtC7KksDS3uQyXAzmVKVNmOxWGl6AVzlKmb" +
						"VGozxcVeh7J2W01S2LOVAsHyj9ZlozgbP+74qVUk4RoMtrfMD98wCzGvEiwXHD3U5GFi" +
						"4Jzo/QhhI8fd0yFu3c/fa/d8zmZU67KsRRDefCt/Qu7YdQSw1PzNTS3W1QGnyRVef+N5" +
						"YHDKZao/4MP/ju/siEpp0SVQYbX5UNlxxJwizCFyzuMWXkLNySzIyZs4wBrTpXE23I62" +
						"wlPRZHp0qJCC7EWslxpSnS8uqgt/YmLr2btnZXaDhnwA4NPzueT8lEt126AyExPY44rS" +
						"YA1bJPl15JgRaEdM9CKv/f1YDHdE5e1cYVFdiUwoduDJC+5mBMe5nstbndCF9Zfxakpa" +
						"1aNP2LK/Xffhuc3fTNfUYlfzH8a/h97qhmVaikNPi2+nItq8exGtLA+SdW9rgUvUvqbq" +
						"YkDi6mRXNk/V1pUxy0uYsI1S+meU+XsPo2kJLnMOKZGy4J6Xt3XgZuHTayEKv3XZLjy+" +
						"yJ66WPQwcHBwcHBwcHBwcHBwcHBwcHhm8Q/mTHqWgAoAAA="),
				},
				rb.DefinitionKey{RBName: "test-rbdef", RBVersion: "v1"}.String(): {
					"defmetadata": []byte(
						"{\"rb-name\":\"test-rbdef\"," +
							"\"rb-version\":\"v1\"," +
							"\"chart-name\":\"vault-consul-dev\"," +
							"\"description\":\"testresourcebundle\"}"),
					// base64 encoding of vagrant/tests/vnfs/testrb/helm/vault-consul-dev
					"defcontent": []byte("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" +
						"cOHDgwIEDB18K/AcxEDJDAHgAAA=="),
				},
				connection.ConnectionKey{CloudRegion: "mock_connection"}.String(): {
					"metadata": []byte(
						"{\"cloud-region\":\"mock_connection\"," +
							"\"cloud-owner\":\"mock_owner\"," +
							"\"kubeconfig\": \"" + base64.StdEncoding.EncodeToString(fd) + "\"}"),
				},
			},
		}

		ic := NewInstanceClient()
		input := InstanceRequest{
			RBName:      "test-rbdef",
			RBVersion:   "v1",
			ProfileName: "profile1",
			CloudRegion: "mock_connection",
		}

		ir, err := ic.Create(input)
		if err != nil {
			t.Fatalf("TestInstanceCreate returned an error (%s)", err)
		}

		log.Println(ir)

		if len(ir.Resources) == 0 {
			t.Fatalf("TestInstanceCreate returned empty data (%+v)", ir)
		}
	})

}

func TestInstanceGet(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	t.Run("Successfully Get Instance", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				InstanceKey{ID: "HaKpys8e"}.String(): {
					"instance": []byte(
						`{
							"id":"HaKpys8e",
							"request": {
								"profile-name":"profile1",
								"rb-name":"test-rbdef",
								"rb-version":"v1",
								"cloud-region":"region1"
							},
							"namespace":"testnamespace",
							"resources": [
								{
									"GVK": {
										"Group":"apps",
										"Version":"v1",
										"Kind":"Deployment"
									},
									"Name": "deployment-1"
								},
								{
									"GVK": {
										"Group":"",
										"Version":"v1",
										"Kind":"Service"
									},
									"Name": "service-1"
								}
							]
						}`),
				},
			},
		}

		expected := InstanceResponse{
			ID: "HaKpys8e",
			Request: InstanceRequest{
				RBName:      "test-rbdef",
				RBVersion:   "v1",
				ProfileName: "profile1",
				CloudRegion: "region1",
			},
			Namespace: "testnamespace",
			Resources: []helm.KubernetesResource{
				{
					GVK: schema.GroupVersionKind{
						Group:   "apps",
						Version: "v1",
						Kind:    "Deployment"},
					Name: "deployment-1",
				},
				{
					GVK: schema.GroupVersionKind{
						Group:   "",
						Version: "v1",
						Kind:    "Service"},
					Name: "service-1",
				},
			},
		}
		ic := NewInstanceClient()
		id := "HaKpys8e"
		data, err := ic.Get(id)
		if err != nil {
			t.Fatalf("TestInstanceGet returned an error (%s)", err)
		}
		if !reflect.DeepEqual(expected, data) {
			t.Fatalf("TestInstanceGet returned:\n result=%v\n expected=%v",
				data, expected)
		}
	})

	t.Run("Get non-existing Instance", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				InstanceKey{ID: "HaKpys8e"}.String(): {
					"instance": []byte(
						`{
							"id":"HaKpys8e",
							"request": {
								"profile-name":"profile1",
								"rb-name":"test-rbdef",
								"rb-version":"v1",
								"cloud-region":"region1"
							},
							"namespace":"testnamespace",
							"resources": [
								{
									"GVK": {
										"Group":"apps",
										"Version":"v1",
										"Kind":"Deployment"
									},
									"Name": "deployment-1"
								},
								{
									"GVK": {
										"Group":"",
										"Version":"v1",
										"Kind":"Service"
									},
									"Name": "service-1"
								}
							]
						}`),
				},
			},
		}

		ic := NewInstanceClient()
		id := "non-existing"
		_, err := ic.Get(id)
		if err == nil {
			t.Fatal("Expected error, got pass", err)
		}
	})
}

func TestInstanceFind(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	items := map[string]map[string][]byte{
		InstanceKey{ID: "HaKpys8e"}.String(): {
			"instance": []byte(
				`{
					"id":"HaKpys8e",
					"request": {
						"profile-name":"profile1",
						"rb-name":"test-rbdef",
						"rb-version":"v1",
						"cloud-region":"region1",
						"labels":{
							"vf_module_id": "test-vf-module-id"
						}
					},
					"namespace":"testnamespace",
					"resources": [
						{
							"GVK": {
								"Group":"apps",
								"Version":"v1",
								"Kind":"Deployment"
							},
							"Name": "deployment-1"
						},
						{
							"GVK": {
								"Group":"",
								"Version":"v1",
								"Kind":"Service"
							},
							"Name": "service-1"
						}
					]
				}`),
		},
		InstanceKey{ID: "HaKpys8f"}.String(): {
			"instance": []byte(
				`{
					"id":"HaKpys8f",
					"request": {
						"profile-name":"profile2",
						"rb-name":"test-rbdef",
						"rb-version":"v1",
						"cloud-region":"region1"
					},
					"namespace":"testnamespace",
					"resources": [
						{
							"GVK": {
								"Group":"apps",
								"Version":"v1",
								"Kind":"Deployment"
							},
							"Name": "deployment-1"
						},
						{
							"GVK": {
								"Group":"",
								"Version":"v1",
								"Kind":"Service"
							},
							"Name": "service-1"
						}
					]
				}`),
		},
		InstanceKey{ID: "HaKpys8g"}.String(): {
			"instance": []byte(
				`{
					"id":"HaKpys8g",
					"request": {
						"profile-name":"profile1",
						"rb-name":"test-rbdef",
						"rb-version":"v2",
						"cloud-region":"region1"
					},
					"namespace":"testnamespace",
					"resources": [
						{
							"GVK": {
								"Group":"apps",
								"Version":"v1",
								"Kind":"Deployment"
							},
							"Name": "deployment-1"
						},
						{
							"GVK": {
								"Group":"",
								"Version":"v1",
								"Kind":"Service"
							},
							"Name": "service-1"
						}
					]
				}`),
		},
	}

	t.Run("Successfully Find Instance By Name", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: items,
		}

		expected := []InstanceMiniResponse{
			{
				ID: "HaKpys8e",
				Request: InstanceRequest{
					RBName:      "test-rbdef",
					RBVersion:   "v1",
					ProfileName: "profile1",
					CloudRegion: "region1",
					Labels: map[string]string{
						"vf_module_id": "test-vf-module-id",
					},
				},
				Namespace: "testnamespace",
			},
			{
				ID: "HaKpys8f",
				Request: InstanceRequest{
					RBName:      "test-rbdef",
					RBVersion:   "v1",
					ProfileName: "profile2",
					CloudRegion: "region1",
				},
				Namespace: "testnamespace",
			},
			{
				ID: "HaKpys8g",
				Request: InstanceRequest{
					RBName:      "test-rbdef",
					RBVersion:   "v2",
					ProfileName: "profile1",
					CloudRegion: "region1",
				},
				Namespace: "testnamespace",
			},
		}
		ic := NewInstanceClient()
		name := "test-rbdef"
		data, err := ic.Find(name, "", "", nil)
		if err != nil {
			t.Fatalf("TestInstanceFind returned an error (%s)", err)
		}

		// Since the order of returned slice is not guaranteed
		// Check both and return error if both don't match
		sort.Slice(data, func(i, j int) bool {
			return data[i].ID < data[j].ID
		})
		// Sort both as it is not expected that testCase.expected
		// is sorted
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].ID < expected[j].ID
		})

		if !reflect.DeepEqual(expected, data) {
			t.Fatalf("TestInstanceFind returned:\n result=%v\n expected=%v",
				data, expected)
		}
	})

	t.Run("Successfully Find Instance By Name and Label", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: items,
		}

		expected := []InstanceMiniResponse{
			{
				ID: "HaKpys8e",
				Request: InstanceRequest{
					RBName:      "test-rbdef",
					RBVersion:   "v1",
					ProfileName: "profile1",
					CloudRegion: "region1",
					Labels: map[string]string{
						"vf_module_id": "test-vf-module-id",
					},
				},
				Namespace: "testnamespace",
			},
		}
		ic := NewInstanceClient()
		name := "test-rbdef"
		labels := map[string]string{
			"vf_module_id": "test-vf-module-id",
		}
		data, err := ic.Find(name, "", "", labels)
		if err != nil {
			t.Fatalf("TestInstanceFind returned an error (%s)", err)
		}

		// Since the order of returned slice is not guaranteed
		// Check both and return error if both don't match
		sort.Slice(data, func(i, j int) bool {
			return data[i].ID < data[j].ID
		})
		// Sort both as it is not expected that testCase.expected
		// is sorted
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].ID < expected[j].ID
		})

		if !reflect.DeepEqual(expected, data) {
			t.Fatalf("TestInstanceFind returned:\n result=%v\n expected=%v",
				data, expected)
		}
	})

	t.Run("Successfully Find Instance By Name Version", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: items,
		}

		expected := []InstanceMiniResponse{
			{
				ID: "HaKpys8e",
				Request: InstanceRequest{
					RBName:      "test-rbdef",
					RBVersion:   "v1",
					ProfileName: "profile1",
					CloudRegion: "region1",
					Labels: map[string]string{
						"vf_module_id": "test-vf-module-id",
					},
				},
				Namespace: "testnamespace",
			},
			{
				ID: "HaKpys8f",
				Request: InstanceRequest{
					RBName:      "test-rbdef",
					RBVersion:   "v1",
					ProfileName: "profile2",
					CloudRegion: "region1",
				},
				Namespace: "testnamespace",
			},
		}
		ic := NewInstanceClient()
		name := "test-rbdef"
		data, err := ic.Find(name, "v1", "", nil)
		if err != nil {
			t.Fatalf("TestInstanceFind returned an error (%s)", err)
		}

		// Since the order of returned slice is not guaranteed
		// Check both and return error if both don't match
		sort.Slice(data, func(i, j int) bool {
			return data[i].ID < data[j].ID
		})
		// Sort both as it is not expected that testCase.expected
		// is sorted
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].ID < expected[j].ID
		})

		if !reflect.DeepEqual(expected, data) {
			t.Fatalf("TestInstanceFind returned:\n result=%v\n expected=%v",
				data, expected)
		}
	})

	t.Run("Successfully Find Instance By Name Version Profile", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: items,
		}

		expected := []InstanceMiniResponse{
			{
				ID: "HaKpys8e",
				Request: InstanceRequest{
					RBName:      "test-rbdef",
					RBVersion:   "v1",
					ProfileName: "profile1",
					CloudRegion: "region1",
					Labels: map[string]string{
						"vf_module_id": "test-vf-module-id",
					},
				},
				Namespace: "testnamespace",
			},
		}
		ic := NewInstanceClient()
		name := "test-rbdef"
		data, err := ic.Find(name, "v1", "profile1", nil)
		if err != nil {
			t.Fatalf("TestInstanceFind returned an error (%s)", err)
		}

		// Since the order of returned slice is not guaranteed
		// Check both and return error if both don't match
		sort.Slice(data, func(i, j int) bool {
			return data[i].ID < data[j].ID
		})
		// Sort both as it is not expected that testCase.expected
		// is sorted
		sort.Slice(expected, func(i, j int) bool {
			return expected[i].ID < expected[j].ID
		})

		if !reflect.DeepEqual(expected, data) {
			t.Fatalf("TestInstanceFind returned:\n result=%v\n expected=%v",
				data, expected)
		}
	})

	t.Run("Find non-existing Instance", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				InstanceKey{ID: "HaKpys8e"}.String(): {
					"instance": []byte(
						`{
							"profile-name":"profile1",
						  	"id":"HaKpys8e",
							"namespace":"testnamespace",
							"rb-name":"test-rbdef",
							"rb-version":"v1",
							"cloud-region":"region1",
							"resources": [
								{
									"GVK": {
										"Group":"apps",
										"Version":"v1",
										"Kind":"Deployment"
									},
									"Name": "deployment-1"
								},
								{
									"GVK": {
										"Group":"",
										"Version":"v1",
										"Kind":"Service"
									},
									"Name": "service-1"
								}
							]
						}`),
				},
			},
		}

		ic := NewInstanceClient()
		name := "non-existing"
		resp, _ := ic.Find(name, "", "", nil)
		if len(resp) != 0 {
			t.Fatalf("Expected 0 responses, but got %d", len(resp))
		}
	})
}

func TestInstanceDelete(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("TestInstanceDelete returned an error (%s)", err)
	}

	// Load the mock kube config file into memory
	fd, err := ioutil.ReadFile("../../mock_files/mock_configs/mock_kube_config")
	if err != nil {
		t.Fatal("Unable to read mock_kube_config")
	}

	t.Run("Successfully delete Instance", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				InstanceKey{ID: "HaKpys8e"}.String(): {
					"instance": []byte(
						`{
							"id":"HaKpys8e",
							"request": {
								"profile-name":"profile1",
								"rb-name":"test-rbdef",
								"rb-version":"v1",
								"cloud-region":"mock_connection"
							},
							"namespace":"testnamespace",
							"resources": [
								{
									"GVK": {
										"Group":"apps",
										"Version":"v1",
										"Kind":"Deployment"
									},
									"Name": "deployment-1"
								},
								{
									"GVK": {
										"Group":"",
										"Version":"v1",
										"Kind":"Service"
									},
									"Name": "service-1"
								}
							]
						}`),
				},
				connection.ConnectionKey{CloudRegion: "mock_connection"}.String(): {
					"metadata": []byte(
						"{\"cloud-region\":\"mock_connection\"," +
							"\"cloud-owner\":\"mock_owner\"," +
							"\"kubeconfig\": \"" + base64.StdEncoding.EncodeToString(fd) + "\"}"),
				},
			},
		}

		ic := NewInstanceClient()
		id := "HaKpys8e"
		err := ic.Delete(id)
		if err != nil {
			t.Fatalf("TestInstanceDelete returned an error (%s)", err)
		}
	})

	t.Run("Delete non-existing Instance", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				InstanceKey{ID: "HaKpys8e"}.String(): {
					"instance": []byte(
						`{
							"profile-name":"profile1",
						  	"id":"HaKpys8e",
							"namespace":"testnamespace",
							"rb-name":"test-rbdef",
							"rb-version":"v1",
							"cloud-region":"mock_kube_config",
							"resources": [
								{
									"GVK": {
										"Group":"apps",
										"Version":"v1",
										"Kind":"Deployment"
									},
									"Name": "deployment-1"
								},
								{
									"GVK": {
										"Group":"",
										"Version":"v1",
										"Kind":"Service"
									},
									"Name": "service-1"
								}
							]
						}`),
				},
			},
		}

		ic := NewInstanceClient()
		id := "non-existing"
		err := ic.Delete(id)
		if err == nil {
			t.Fatal("Expected error, got pass", err)
		}
	})
}

//TODO: add a test when pre-hook is failed (if possible)
func TestInstanceWithHookCreate(t *testing.T) {
	err := LoadMockPlugins(utils.LoadedPlugins)
	if err != nil {
		t.Fatalf("LoadMockPlugins returned an error (%s)", err)
	}

	// Load the mock kube config file into memory
	fd, err := ioutil.ReadFile("../../mock_files/mock_configs/mock_kube_config")
	if err != nil {
		t.Fatal("Unable to read mock_kube_config")
	}

	t.Run("Successfully create Instance With Hook", func(t *testing.T) {
		db.DBconn = &db.MockDB{
			Items: map[string]map[string][]byte{
				rb.ProfileKey{RBName: "test-rbdef-hook", RBVersion: "v1",
					ProfileName: "profile1"}.String(): {
					"profilemetadata": []byte(
						"{\"profile-name\":\"profile1\"," +
							"\"release-name\":\"testprofilereleasename\"," +
							"\"namespace\":\"testnamespace\"," +
							"\"rb-name\":\"test-rbdef\"," +
							"\"rb-version\":\"v1\"," +
							"\"kubernetesversion\":\"1.12.3\"}"),
					// base64 encoding of vagrant/tests/vnfs/testrb/helm/profile
					"profilecontent": []byte("H4sICCVd3FwAA3Byb2ZpbGUxLnRhcgDt1NEKgjAUxvFd7ylG98aWO" +
						"sGXiYELxLRwJvj2rbyoIPDGiuD/uzmwM9iB7Vvruvrgw7CdXHsUn6Ejm2W3aopcP9eZL" +
						"YRJM1voPN+ZndAm16kVSn9onheXMLheKeGqfdM0rq07/3bfUv9PJUkiR9+H+tSVajRym" +
						"M6+lEqN7njxoVSbU+z2deX388r9nWzkr8fGSt5d79pnLOZfm0f+dRrzb7P4DZD/LyDJA" +
						"AAAAAAAAAAAAAAA/+0Ksq1N5QAoAAA="),
				},
				rb.DefinitionKey{RBName: "test-rbdef-hook", RBVersion: "v1"}.String(): {
					"defmetadata": []byte(
						"{\"rb-name\":\"test-rbdef-hook\"," +
							"\"rb-version\":\"v1\"," +
							"\"chart-name\":\"test\"," +
							"\"description\":\"testresourcebundle\"}"),
					// base64 encoding of test helm package with hooks inside
					"defcontent": []byte("H4sICE+Q8WAAA3Rlc3QudGFyAO1aW2+jOBTOM7/CYl6HlEsIq7xV24" +
						"fVqluNdlYjrVajkQMnhS1gFjvZjbr972MDJYTQwGhMMmn9qVUaYx/o8TnfuRgGlF1NxoX" +
						"J4Xmu+LQ812x+PmNiOXzEMe3ZfD4xLdO23QlyR36uAmvKcI7QhIXs6Ly+6xcKJvZ/g+M1" +
						"0OkWJ/EY9xAbPJ/PXtx/m9tGtf+WOePjlu143gSZYzxMG298/9+hG1jhdcxQaQRoRXKU5" +
						"WBEKVdMHEM+1d6hP8KIIv6D0Z/Xv90afE6CGYMAraIYxIQb8GOcAxeSR3gZczmMoCWgDF" +
						"PKp0Up/8pCQAySLMbc6KYaDpIoXWgIhYQ8fAkgBgZfMhJH/naBdDFo0LXvAwQQvOey+E3" +
						"BKIb9HDCLSKqfW3mvAIX//xzinI3m/r3+b7nzZ/83Z57gf9tyHeX/pwDOok+QU+5NC7Sx" +
						"NJxl9VfdmppTU9cCoH4eZawYvEa/QJwgX1hMwRXCgKL0HiWcQyI/JutAS3ECi+KCtnkWV" +
						"sjSzv3fKrRR+H/NyuNkgoPyv5npzRzxOxP+b9uOyv9Ogdb+BxgSklKQGg36+N+zZ7v9tw" +
						"X/u3xM8f8p0OR/Tv70igeBhygNFuimMIWPwLQEGA4wwyJZK7n98RFNf+cZG6YwveMj6On" +
						"JqE2nmkUz7POp+uPj3tRi+OlJ57NivISYCqlI3LtPLM3AF5Mpn+EzkpcLeSLqh7cNSYNk" +
						"oToTraQ0/kWBeE/gQJH80apHFPBJynCUcuU+jxiV9uortfgowfdCV8s13S7Jf3p9gbKAJ" +
						"8mI5WuoxxjbtkZ8kiRY7NlfOg31z9+y/y3/zwhlRpmLG3+TpRwW6PF/25l7Vf5nWZaIE9" +
						"ac/6H8/xRo+v9SuNKOAH4ly4Gu37IaSy4DdEjHaUpYUQNWi/WQZ6VTGl6JAlFfoMaaw+v" +
						"GvxDdh4xP042f9I7r1c3KYlQvn+pT2SMpqtbpYcmK/kf/rAkTD1wT1RL7D2S1uo2SiC2Q" +
						"I490OjSyzz2Up+fwISc+UHq324kGaeQg7J59qOrtO9jUdHRIXDvqojFAZhwS2BEK26cns" +
						"V5/z2sLU/+sGYahjWGA9qgGaMs0QPMV2J89tv31Wd+LttdlebawvHPT7g+DdvzPQXr474" +
						"//7i7+25Yt8n/PVPH/JJBDv3tWIzv8HwjvJ996yWsM/gf6eOOxf08fskP/gXBZxneZgf9" +
						"AHSruXzZa8Z9Cvol8kHsW1Nf/K+r/sv83dx3R/5u5rjr/PQla5z8l+X4srWAgAVc2I7nt" +
						"B1lMtgmku75fRnJWLTOKLwtkces56AgOkXlutf8waPf/axVJpIDe/r9jtc5/XNszlf+fA" +
						"kf6/ztvGXgAsFswNhV8xxFA8yFlnQE0ZV7YIUBH/V+9+XOy/v/M9qxd/PfMsv/vKv8/BY" +
						"7F/2vfJ+vB7k9xUaJwC6oMaKh/dy0cVGXtph+p8d0R6iyptWvD3UbonLSky9PrxfZOWhp" +
						"RzZOGQkbonrSkSzPAi+2ftBRyYQ2UtuV9Z87YVMhY+eOL95Bmi9YQW9Q7X2GWkNLuP6V8" +
						"Sx2Q1B5La48yXFdq25XcHqS3qoKXg673f2QXAL3nf17j/M8U539zx1T5/0kg7/WLEfPYD" +
						"vHDXsB4xZlsh07eeCrb0sgYLwF9czI71AgvM5vtUMmFpbPnpl8FBQUFBQUFBQUFBQUFBQ" +
						"UFBQUFhdHwFf2f+3IAUAAA"),
				},
				connection.ConnectionKey{CloudRegion: "mock_connection"}.String(): {
					"metadata": []byte(
						"{\"cloud-region\":\"mock_connection\"," +
							"\"cloud-owner\":\"mock_owner\"," +
							"\"kubeconfig\": \"" + base64.StdEncoding.EncodeToString(fd) + "\"}"),
				},
			},
		}

		ic := NewInstanceClient()
		input := InstanceRequest{
			RBName:      "test-rbdef-hook",
			RBVersion:   "v1",
			ProfileName: "profile1",
			CloudRegion: "mock_connection",
		}

		ir, err := ic.Create(input)
		if err != nil {
			t.Fatalf("TestInstanceWithHookCreate returned an error (%s)", err)
		}

		log.Println(ir)

		if len(ir.Resources) == 0 {
			t.Fatalf("TestInstanceWithHookCreate returned empty data (%+v)", ir)
		}
	})
}
