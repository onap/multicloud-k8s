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

package api

import (
	"bytes"
	"encoding/json"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"k8-plugin-multicloud/csar"
	"k8-plugin-multicloud/db"
)

type mockDB struct {
	db.DatabaseConnection
}

func (c *mockDB) InitializeDatabase() error {
	return nil
}

func (c *mockDB) CheckDatabase() error {
	return nil
}

func (c *mockDB) CreateEntry(key string, value string) error {
	return nil
}

func (c *mockDB) ReadEntry(key string) (string, bool, error) {
	str := "{\"deployment\":[\"cloud1-default-uuid-sisedeploy\"],\"service\":[\"cloud1-default-uuid-sisesvc\"]}"
	return str, true, nil
}

func (c *mockDB) DeleteEntry(key string) error {
	return nil
}

func (c *mockDB) ReadAll(key string) ([]string, error) {
	returnVal := []string{"cloud1-default-uuid1", "cloud1-default-uuid2"}
	return returnVal, nil
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	router := NewRouter("")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	return recorder
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d\n", expected, actual)
	}
}

func TestVNFInstanceCreation(t *testing.T) {
	t.Run("Succesful create a VNF", func(t *testing.T) {
		payload := []byte(`{
			"cloud_region_id": "region1",
			"namespace": "test",
			"csar_id": "UUID-1",
			"oof_parameters": [{
				"key1": "value1",
				"key2": "value2",
				"key3": {}
			}],
			"network_parameters": {
				"oam_ip_address": {
					"connection_point": "string",
					"ip_address": "string",
					"workload_name": "string"
				}
			}
		}`)

		data := map[string][]string{
			"deployment": []string{"cloud1-default-uuid-sisedeploy"},
			"service":    []string{"cloud1-default-uuid-sisesvc"},
		}

		expected := &CreateVnfResponse{
			VNFID:         "test_UUID",
			CloudRegionID: "region1",
			Namespace:     "test",
			VNFComponents: data,
		}

		var result CreateVnfResponse

		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", bytes.NewBuffer(payload))

		GetVNFClient = func(configPath string) (kubernetes.Clientset, error) {
			return kubernetes.Clientset{}, nil
		}

		csar.CreateVNF = func(id string, r string, n string, kubeclient *kubernetes.Clientset) (string, map[string][]string, error) {
			return "externaluuid", data, nil
		}

		db.DBconn = &mockDB{}

		response := executeRequest(req)
		checkResponseCode(t, http.StatusCreated, response.Code)

		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("TestVNFInstanceCreation returned:\n result=%v\n expected=%v", err, expected.VNFComponents)
		}
	})
	t.Run("Missing body failure", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", nil)
		response := executeRequest(req)

		checkResponseCode(t, http.StatusBadRequest, response.Code)
	})
	t.Run("Invalid JSON request format", func(t *testing.T) {
		payload := []byte("invalid")
		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", bytes.NewBuffer(payload))
		response := executeRequest(req)
		checkResponseCode(t, http.StatusUnprocessableEntity, response.Code)
	})
	t.Run("Missing parameter failure", func(t *testing.T) {
		payload := []byte(`{
			"csar_id": "testID",
			"oof_parameters": {
				"key_values": {
					"key1": "value1",
					"key2": "value2"
				}
			},
			"vnf_instance_name": "test",
			"vnf_instance_description": "vRouter_test_description"
		}`)
		req, _ := http.NewRequest("POST", "/v1/vnf_instances/", bytes.NewBuffer(payload))
		response := executeRequest(req)
		checkResponseCode(t, http.StatusUnprocessableEntity, response.Code)
	})
}

func TestVNFInstancesRetrieval(t *testing.T) {
	t.Run("Succesful get a list of VNF", func(t *testing.T) {
		expected := &ListVnfsResponse{
			VNFs: []string{"uuid1", "uuid2"},
		}
		var result ListVnfsResponse

		req, _ := http.NewRequest("GET", "/v1/vnf_instances/cloud1/default", nil)

		db.DBconn = &mockDB{}

		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("TestVNFInstancesRetrieval returned:\n result=%v\n expected=list", err)
		}
		if !reflect.DeepEqual(*expected, result) {
			t.Fatalf("TestVNFInstancesRetrieval returned:\n result=%v\n expected=%v", result, *expected)
		}
	})
	t.Run("Get empty list", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/vnf_instances/cloudregion1/testnamespace", nil)
		db.DBconn = &mockDB{}
		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)
	})
}

func TestVNFInstanceDeletion(t *testing.T) {
	t.Run("Succesful delete a VNF", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/v1/vnf_instances/cloudregion1/testnamespace/1", nil)

		GetVNFClient = func(configPath string) (kubernetes.Clientset, error) {
			return kubernetes.Clientset{}, nil
		}

		csar.DestroyVNF = func(d map[string][]string, n string, kubeclient *kubernetes.Clientset) error {
			return nil
		}

		db.DBconn = &mockDB{}

		response := executeRequest(req)
		checkResponseCode(t, http.StatusAccepted, response.Code)

		if result := response.Body.String(); result != "" {
			t.Fatalf("TestVNFInstanceDeletion returned:\n result=%v\n expected=%v", result, "")
		}
	})
	// t.Run("Malformed delete request", func(t *testing.T) {
	// 	req, _ := http.NewRequest("DELETE", "/v1/vnf_instances/foo", nil)
	// 	response := executeRqequest(req)
	// 	checkResponseCode(t, http.StatusBadRequest, response.Code)
	// })
}

// TODO: Update this test when the UpdateVNF endpoint is fixed.
/*
func TestVNFInstanceUpdate(t *testing.T) {
	t.Run("Succesful update a VNF", func(t *testing.T) {
		payload := []byte(`{
			"cloud_region_id": "region1",
			"csar_id": "UUID-1",
			"oof_parameters": [{
				"key1": "value1",
				"key2": "value2",
				"key3": {}
			}],
			"network_parameters": {
				"oam_ip_address": {
					"connection_point": "string",
					"ip_address": "string",
					"workload_name": "string"
				}
			}
		}`)
		expected := &UpdateVnfResponse{
			DeploymentID: "1",
		}

		var result UpdateVnfResponse

		req, _ := http.NewRequest("PUT", "/v1/vnf_instances/1", bytes.NewBuffer(payload))

		GetVNFClient = func(configPath string) (krd.VNFInstanceClientInterface, error) {
			return &mockClient{
				update: func() error {
					return nil
				},
			}, nil
		}
		utils.ReadCSARFromFileSystem = func(csarID string) (*krd.KubernetesData, error) {
			kubeData := &krd.KubernetesData{
				Deployment: &appsV1.Deployment{},
				Service:    &coreV1.Service{},
			}
			return kubeData, nil
		}

		response := executeRequest(req)
		checkResponseCode(t, http.StatusCreated, response.Code)

		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("TestVNFInstanceUpdate returned:\n result=%v\n expected=%v", err, expected.DeploymentID)
		}

		if result.DeploymentID != expected.DeploymentID {
			t.Fatalf("TestVNFInstanceUpdate returned:\n result=%v\n expected=%v", result.DeploymentID, expected.DeploymentID)
		}
	})
}
*/

func TestVNFInstanceRetrieval(t *testing.T) {
	t.Run("Succesful get a VNF", func(t *testing.T) {

		data := map[string][]string{
			"deployment": []string{"cloud1-default-uuid-sisedeploy"},
			"service":    []string{"cloud1-default-uuid-sisesvc"},
		}

		expected := GetVnfResponse{
			VNFID:         "1",
			CloudRegionID: "cloud1",
			Namespace:     "default",
			VNFComponents: data,
		}

		req, _ := http.NewRequest("GET", "/v1/vnf_instances/cloud1/default/1", nil)

		GetVNFClient = func(configPath string) (kubernetes.Clientset, error) {
			return kubernetes.Clientset{}, nil
		}

		db.DBconn = &mockDB{}

		response := executeRequest(req)
		checkResponseCode(t, http.StatusOK, response.Code)

		var result GetVnfResponse

		err := json.NewDecoder(response.Body).Decode(&result)
		if err != nil {
			t.Fatalf("TestVNFInstanceRetrieval returned:\n result=%v\n expected=%v", err, expected)
		}

		if !reflect.DeepEqual(expected, result) {
			t.Fatalf("TestVNFInstanceRetrieval returned:\n result=%v\n expected=%v", result, expected)
		}
	})
	t.Run("VNF not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/vnf_instances/cloudregion1/testnamespace/1", nil)
		response := executeRequest(req)

		checkResponseCode(t, http.StatusOK, response.Code)
	})
}
