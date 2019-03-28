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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"k8splugin/internal/app"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockInstanceClient struct {
	app.InstanceManager
	// Items and err will be used to customize each test
	// via a localized instantiation of mockInstanceClient
	items []app.InstanceResponse
	err   error
}

func (m *mockInstanceClient) Create(inp app.InstanceRequest) (app.InstanceResponse, error) {
	if m.err != nil {
		return app.InstanceResponse{}, m.err
	}

	return m.items[0], nil
}

func (m *mockInstanceClient) Get(id string) (app.InstanceResponse, error) {
	if m.err != nil {
		return app.InstanceResponse{}, m.err
	}

	return m.items[0], nil
}

func (m *mockInstanceClient) Delete(id string) error {
	return m.err
}

func executeRequest(request *http.Request, router *mux.Router) *http.Response {
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	resp := recorder.Result()
	return resp
}

func TestInstanceCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expected     app.InstanceResponse
		expectedCode int
		instClient   *mockInstanceClient
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
				"rb-name": "test-rbdef",
				"profile-name": "profile1",
				"cloud-region": "kud"
			}`)),
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			label: "Succesfully create an Instance",
			input: bytes.NewBuffer([]byte(`{
				"cloud-region": "region1",
				"rb-name": "test-rbdef",
				"rb-version": "v1",
				"profile-name": "profile1"
			}`)),
			expected: app.InstanceResponse{
				ID:          "HaKpys8e",
				RBName:      "test-rbdef",
				RBVersion:   "v1",
				ProfileName: "profile1",
				CloudRegion: "region1",
				Namespace:   "testnamespace",
				Resources: map[string][]string{
					"deployment": []string{"test-deployment"},
					"service":    []string{"test-service"},
				},
			},
			expectedCode: http.StatusCreated,
			instClient: &mockInstanceClient{
				items: []app.InstanceResponse{
					{
						ID:          "HaKpys8e",
						RBName:      "test-rbdef",
						RBVersion:   "v1",
						ProfileName: "profile1",
						CloudRegion: "region1",
						Namespace:   "testnamespace",
						Resources: map[string][]string{
							"deployment": []string{"test-deployment"},
							"service":    []string{"test-service"},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {

			request := httptest.NewRequest("POST", "/v1/instance", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient))

			if testCase.expectedCode != resp.StatusCode {
				body, _ := ioutil.ReadAll(resp.Body)
				t.Log(string(body))
				t.Fatalf("Request method returned: \n%v\n and it was expected: \n%v", resp.StatusCode, testCase.expectedCode)
			}

			if resp.StatusCode == http.StatusCreated {
				var response app.InstanceResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Parsing the returned response got an error (%s)", err)
				}
			}
		})
	}
}

func TestInstanceGetHandler(t *testing.T) {
	testCases := []struct {
		label            string
		input            string
		expectedCode     int
		expectedResponse *app.InstanceResponse
		instClient       *mockInstanceClient
	}{
		{
			label:        "Fail to retrieve Instance",
			input:        "HaKpys8e",
			expectedCode: http.StatusInternalServerError,
			instClient: &mockInstanceClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Succesful get an Instance",
			input:        "HaKpys8e",
			expectedCode: http.StatusOK,
			expectedResponse: &app.InstanceResponse{
				ID:          "HaKpys8e",
				RBName:      "test-rbdef",
				RBVersion:   "v1",
				ProfileName: "profile1",
				CloudRegion: "region1",
				Namespace:   "testnamespace",
				Resources: map[string][]string{
					"deployment": []string{"test-deployment"},
					"service":    []string{"test-service"},
				},
			},
			instClient: &mockInstanceClient{
				items: []app.InstanceResponse{
					{
						ID:          "HaKpys8e",
						RBName:      "test-rbdef",
						RBVersion:   "v1",
						ProfileName: "profile1",
						CloudRegion: "region1",
						Namespace:   "testnamespace",
						Resources: map[string][]string{
							"deployment": []string{"test-deployment"},
							"service":    []string{"test-service"},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v",
					resp.StatusCode, testCase.expectedCode)
			}
			if resp.StatusCode == http.StatusOK {
				var response app.InstanceResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
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

func TestDeleteHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        string
		expectedCode int
		instClient   *mockInstanceClient
	}{
		{
			label:        "Fail to destroy VNF",
			input:        "HaKpys8e",
			expectedCode: http.StatusInternalServerError,
			instClient: &mockInstanceClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Succesful delete a VNF",
			input:        "HaKpys8e",
			expectedCode: http.StatusAccepted,
			instClient:   &mockInstanceClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v1/instance/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
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

		req := httptest.NewRequest("PUT", "/v1/vnf_instances/1", bytes.NewBuffer(payload))

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

		if resp.DeploymentID != expected.DeploymentID {
			t.Fatalf("TestVNFInstanceUpdate returned:\n result=%v\n expected=%v", resp.DeploymentID, expected.DeploymentID)
		}
	})
}
*/
