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
	"k8splugin/internal/helm"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestBrokerCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expected     brokerPOSTResponse
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
				"vf-module-model-customization-id": "84sdfkio938",
				"user_directives": {
					"definition-name": "test-rbdef",
					"definition-version": "v1"				}
			}`)),
			expectedCode: http.StatusBadRequest,
		},
		{
			label: "Succesfully create an Instance",
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-customization-id": "84sdfkio938",
				"user_directives": {
					"definition-name": "test-rbdef",
					"definition-version": "v1",
					"profile-name": "profile1"
				}
			}`)),
			expected: brokerPOSTResponse{
				WorkloadID:   "HaKpys8e",
				TemplateType: "heat",
				TemplateResponse: []helm.KubernetesResource{
					{
						GVK: schema.GroupVersionKind{
							Group:   "apps",
							Version: "v1",
							Kind:    "Deployment"},
						Name: "test-deployment",
					},
					{
						GVK: schema.GroupVersionKind{
							Group:   "",
							Version: "v1",
							Kind:    "Service"},
						Name: "test-service",
					},
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
						Resources: []helm.KubernetesResource{
							{
								GVK: schema.GroupVersionKind{
									Group:   "apps",
									Version: "v1",
									Kind:    "Deployment"},
								Name: "test-deployment",
							},
							{
								GVK: schema.GroupVersionKind{
									Group:   "",
									Version: "v1",
									Kind:    "Service"},
								Name: "test-service",
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {

			request := httptest.NewRequest("POST", "/v1/cloudowner/cloudregion/infra_workload", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil))

			if testCase.expectedCode != resp.StatusCode {
				body, _ := ioutil.ReadAll(resp.Body)
				t.Log(string(body))
				t.Fatalf("Request method returned: \n%v\n and it was expected: \n%v", resp.StatusCode, testCase.expectedCode)
			}

			if resp.StatusCode == http.StatusCreated {
				var response brokerPOSTResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Parsing the returned response got an error (%s)", err)
				}
				if !reflect.DeepEqual(testCase.expected, response) {
					t.Fatalf("TestGetHandler returned:\n result=%v\n expected=%v",
						response, testCase.expected)
				}
			}
		})
	}
}

func TestBrokerGetHandler(t *testing.T) {
	testCases := []struct {
		label            string
		input            string
		expectedCode     int
		expectedResponse brokerGETResponse
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
			expectedResponse: brokerGETResponse{
				TemplateType:   "heat",
				WorkloadID:     "HaKpys8e",
				WorkloadStatus: "CREATED",
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
						Resources: []helm.KubernetesResource{
							{
								GVK: schema.GroupVersionKind{
									Group:   "apps",
									Version: "v1",
									Kind:    "Deployment"},
								Name: "test-deployment",
							},
							{
								GVK: schema.GroupVersionKind{
									Group:   "",
									Version: "v1",
									Kind:    "Service"},
								Name: "test-service",
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/cloudowner/cloudregion/infra_workload/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v",
					resp.StatusCode, testCase.expectedCode)
			}
			if resp.StatusCode == http.StatusOK {
				var response brokerGETResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Parsing the returned response got an error (%s)", err)
				}
				if !reflect.DeepEqual(testCase.expectedResponse, response) {
					t.Fatalf("TestGetHandler returned:\n result=%v\n expected=%v",
						response, testCase.expectedResponse)
				}
			}
		})
	}
}

func TestBrokerDeleteHandler(t *testing.T) {
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
			request := httptest.NewRequest("DELETE", "/v1/cloudowner/cloudregion/infra_workload/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}
