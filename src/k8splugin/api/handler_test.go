// +build unit

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
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hashicorp/consul/api"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"

	"k8splugin/csar"
	"k8splugin/db"
)

type mockCSAR struct {
	externalVNFID       string
	resourceYAMLNameMap map[string][]string
	err                 error
}

func (c *mockCSAR) CreateVNF(id, r, n string,
	kubeclient *kubernetes.Clientset) (string, map[string][]string, error) {
	return c.externalVNFID, c.resourceYAMLNameMap, c.err
}

func (c *mockCSAR) DestroyVNF(data map[string][]string, namespace string,
	kubeclient *kubernetes.Clientset) error {
	return c.err
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

func TestCreateHandler(t *testing.T) {
	testCases := []struct {
		label               string
		input               io.Reader
		expectedCode        int
		mockGetVNFClientErr error
		mockCreateVNF       *mockCSAR
		mockStore           *db.MockDB
	}{
		{
			label:        "Missing body failure",
			expectedCode: http.StatusBadRequest,
		},
		{
			label:        "Invalid JSON request format",
			input:        bytes.NewBuffer([]byte("invalid")),
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			label: "Missing parameter failure",
			input: bytes.NewBuffer([]byte(`{
				"csar_id": "testID",
				"oof_parameters": {
					"key_values": {
						"key1": "value1",
						"key2": "value2"
					}
				},
				"vnf_instance_name": "test",
				"vnf_instance_description": "vRouter_test_description"
			}`)),
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			label: "Fail to get the VNF client",
			input: bytes.NewBuffer([]byte(`{
				"cloud_region_id": "region1",
				"namespace": "test",
				"csar_id": "UUID-1"
			}`)),
			expectedCode:        http.StatusInternalServerError,
			mockGetVNFClientErr: pkgerrors.New("Get VNF client error"),
		},
		{
			label: "Fail to create the VNF instance",
			input: bytes.NewBuffer([]byte(`{
				"cloud_region_id": "region1",
				"namespace": "test",
				"csar_id": "UUID-1"
			}`)),
			expectedCode: http.StatusInternalServerError,
			mockCreateVNF: &mockCSAR{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label: "Fail to create a VNF DB record",
			input: bytes.NewBuffer([]byte(`{
				"cloud_region_id": "region1",
				"namespace": "test",
				"csar_id": "UUID-1"
			}`)),
			expectedCode: http.StatusInternalServerError,
			mockCreateVNF: &mockCSAR{
				resourceYAMLNameMap: map[string][]string{},
			},
			mockStore: &db.MockDB{
				Err: pkgerrors.New("Internal error"),
			},
		},
		{
			label: "Succesful create a VNF",
			input: bytes.NewBuffer([]byte(`{
				"cloud_region_id": "region1",
				"namespace": "test",
				"csar_id": "UUID-1"
			}`)),
			expectedCode: http.StatusCreated,
			mockCreateVNF: &mockCSAR{
				resourceYAMLNameMap: map[string][]string{
					"deployment": []string{"cloud1-default-uuid-sisedeploy"},
					"service":    []string{"cloud1-default-uuid-sisesvc"},
				},
			},
			mockStore: &db.MockDB{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			GetVNFClient = func(configPath string) (kubernetes.Clientset, error) {
				return kubernetes.Clientset{}, testCase.mockGetVNFClientErr
			}
			if testCase.mockCreateVNF != nil {
				csar.CreateVNF = testCase.mockCreateVNF.CreateVNF
			}
			if testCase.mockStore != nil {
				db.DBconn = testCase.mockStore
			}

			request, _ := http.NewRequest("POST", "/v1/vnf_instances/", testCase.input)
			result := executeRequest(request)

			if testCase.expectedCode != result.Code {
				t.Fatalf("Request method returned: \n%v\n and it was expected: \n%v", result.Code, testCase.expectedCode)
			}
			if result.Code == http.StatusCreated {
				var response CreateVnfResponse
				err := json.NewDecoder(result.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Parsing the returned response got an error (%s)", err)
				}
			}
		})
	}
}

func TestListHandler(t *testing.T) {
	testCases := []struct {
		label            string
		expectedCode     int
		expectedResponse []string
		mockStore        *db.MockDB
	}{
		{
			label:        "Fail to retrieve DB records",
			expectedCode: http.StatusInternalServerError,
			mockStore: &db.MockDB{
				Err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Get result from DB non-records",
			expectedCode: http.StatusNotFound,
			mockStore:    &db.MockDB{},
		},
		{
			label:            "Get empty list",
			expectedCode:     http.StatusOK,
			expectedResponse: []string{""},
			mockStore: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key:   "",
						Value: []byte("{}"),
					},
				},
			},
		},
		{
			label:            "Succesful get a list of VNF",
			expectedCode:     http.StatusOK,
			expectedResponse: []string{"uid1", "uid2"},
			mockStore: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key:   "uuid1",
						Value: []byte("{}"),
					},
					&api.KVPair{
						Key:   "uuid2",
						Value: []byte("{}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			if testCase.mockStore != nil {
				db.DBconn = testCase.mockStore
			}

			request, _ := http.NewRequest("GET", "/v1/vnf_instances/cloud1/default", nil)
			result := executeRequest(request)

			if testCase.expectedCode != result.Code {
				t.Fatalf("Request method returned: \n%v\n and it was expected: \n%v",
					result.Code, testCase.expectedCode)
			}
			if result.Code == http.StatusOK {
				var response ListVnfsResponse
				err := json.NewDecoder(result.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Parsing the returned response got an error (%s)", err)
				}
				if !reflect.DeepEqual(testCase.expectedResponse, response.VNFs) {
					t.Fatalf("TestListHandler returned:\n result=%v\n expected=%v",
						response.VNFs, testCase.expectedResponse)
				}
			}
		})
	}
}

func TestDeleteHandler(t *testing.T) {
	testCases := []struct {
		label               string
		expectedCode        int
		mockGetVNFClientErr error
		mockDeleteVNF       *mockCSAR
		mockStore           *db.MockDB
	}{
		{
			label:        "Fail to read a VNF DB record",
			expectedCode: http.StatusInternalServerError,
			mockStore: &db.MockDB{
				Err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Fail to find VNF record be deleted",
			expectedCode: http.StatusNotFound,
			mockStore: &db.MockDB{
				Items: api.KVPairs{},
			},
		},
		{
			label:        "Fail to unmarshal the DB record",
			expectedCode: http.StatusInternalServerError,
			mockStore: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key:   "cloudregion1-testnamespace-uuid1",
						Value: []byte("{invalid format}"),
					},
				},
			},
		},
		{
			label:               "Fail to get the VNF client",
			expectedCode:        http.StatusInternalServerError,
			mockGetVNFClientErr: pkgerrors.New("Get VNF client error"),
			mockStore: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key: "cloudregion1-testnamespace-uuid1",
						Value: []byte("{" +
							"\"deployment\": [\"deploy1\", \"deploy2\"]," +
							"\"service\": [\"svc1\", \"svc2\"]" +
							"}"),
					},
				},
			},
		},
		{
			label:        "Fail to destroy VNF",
			expectedCode: http.StatusInternalServerError,
			mockStore: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key: "cloudregion1-testnamespace-uuid1",
						Value: []byte("{" +
							"\"deployment\": [\"deploy1\", \"deploy2\"]," +
							"\"service\": [\"svc1\", \"svc2\"]" +
							"}"),
					},
				},
			},
			mockDeleteVNF: &mockCSAR{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Succesful delete a VNF",
			expectedCode: http.StatusAccepted,
			mockStore: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key: "cloudregion1-testnamespace-uuid1",
						Value: []byte("{" +
							"\"deployment\": [\"deploy1\", \"deploy2\"]," +
							"\"service\": [\"svc1\", \"svc2\"]" +
							"}"),
					},
				},
			},
			mockDeleteVNF: &mockCSAR{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			GetVNFClient = func(configPath string) (kubernetes.Clientset, error) {
				return kubernetes.Clientset{}, testCase.mockGetVNFClientErr
			}
			if testCase.mockStore != nil {
				db.DBconn = testCase.mockStore
			}
			if testCase.mockDeleteVNF != nil {
				csar.DestroyVNF = testCase.mockDeleteVNF.DestroyVNF
			}

			request, _ := http.NewRequest("DELETE", "/v1/vnf_instances/cloudregion1/testnamespace/uuid1", nil)
			result := executeRequest(request)

			if testCase.expectedCode != result.Code {
				t.Fatalf("Request method returned: %v and it was expected: %v", result.Code, testCase.expectedCode)
			}
		})
	}
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

func TestGetHandler(t *testing.T) {
	testCases := []struct {
		label            string
		expectedCode     int
		expectedResponse *GetVnfResponse
		mockStore        *db.MockDB
	}{
		{
			label:        "Fail to retrieve DB record",
			expectedCode: http.StatusInternalServerError,
			mockStore: &db.MockDB{
				Err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Not found DB record",
			expectedCode: http.StatusNotFound,
			mockStore:    &db.MockDB{},
		},
		{
			label:        "Fail to unmarshal the DB record",
			expectedCode: http.StatusInternalServerError,
			mockStore: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key:   "cloud1-default-1",
						Value: []byte("{invalid-format}"),
					},
				},
			},
		},
		{
			label:        "Succesful get a list of VNF",
			expectedCode: http.StatusOK,
			expectedResponse: &GetVnfResponse{
				VNFID:         "1",
				CloudRegionID: "cloud1",
				Namespace:     "default",
				VNFComponents: map[string][]string{
					"deployment": []string{"deploy1", "deploy2"},
					"service":    []string{"svc1", "svc2"},
				},
			},
			mockStore: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key: "cloud1-default-1",
						Value: []byte("{" +
							"\"deployment\": [\"deploy1\", \"deploy2\"]," +
							"\"service\": [\"svc1\", \"svc2\"]" +
							"}"),
					},
					&api.KVPair{
						Key:   "cloud1-default-2",
						Value: []byte("{}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockStore
			request, _ := http.NewRequest("GET", "/v1/vnf_instances/cloud1/default/1", nil)
			result := executeRequest(request)

			if testCase.expectedCode != result.Code {
				t.Fatalf("Request method returned: %v and it was expected: %v",
					result.Code, testCase.expectedCode)
			}
			if result.Code == http.StatusOK {
				var response GetVnfResponse
				err := json.NewDecoder(result.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Parsing the returned response got an error (%s)", err)
				}
				if !reflect.DeepEqual(testCase.expectedResponse, &response) {
					t.Fatalf("TestGetHandler returned:\n result=%v\n expected=%v",
						&response, testCase.expectedResponse)
				}
			}
		})
	}
}
