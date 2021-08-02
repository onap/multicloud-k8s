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
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestConvertDirectives(t *testing.T) {
	testCases := []struct {
		label    string
		input    brokerRequest
		expected map[string]string
	}{
		{
			label:    "Single variable",
			expected: map[string]string{"test": "true"},
			input: brokerRequest{SDNCDirectives: directive{[]attribute{{
				Key:   "test",
				Value: "true",
			}}}},
		},
		{
			label:    "Empty parameter",
			expected: map[string]string{"test": ""},
			input: brokerRequest{OOFDirectives: directive{[]attribute{{
				Key:   "test",
				Value: "",
			}}}},
		},
		{
			label:    "Null entry",
			input:    brokerRequest{},
			expected: make(map[string]string),
		},
		{
			label: "Complex helm overrides",
			/*
				String with int will be later treated as int in helm.TemplateClient
				(see helm/pkg/strvals/parser.go)
				If unsure, convert type in helm chart like `{{ toString $value }}` or `{{ int $value }}`
				(see http://masterminds.github.io/sprig/conversion.html)
			*/
			expected: map[string]string{"name": "{a, b, c}", "servers[0].port": "80"},
			input: brokerRequest{UserDirectives: directive{[]attribute{
				{
					Key:   "name",
					Value: "{a, b, c}",
				},
				{
					Key:   "servers[0].port",
					Value: "80",
				},
			}}},
		},
		{
			label:    "Override variables",
			expected: map[string]string{"empty": "", "sdnc": "sdnc", "user": "user", "oof": "oof"},
			input: brokerRequest{
				SDNCDirectives: directive{[]attribute{
					{
						Key:   "empty",
						Value: "sdnc",
					},
					{
						Key:   "sdnc",
						Value: "sdnc",
					},
					{
						Key:   "oof",
						Value: "sdnc",
					},
				}},
				OOFDirectives: directive{[]attribute{
					{
						Key:   "oof",
						Value: "oof",
					},
					{
						Key:   "user",
						Value: "oof",
					},
				}},
				UserDirectives: directive{[]attribute{
					{
						Key:   "user",
						Value: "user",
					},
					{
						Key:   "empty",
						Value: "",
					},
				}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			result := testCase.input.convertDirectives()
			if !reflect.DeepEqual(result, testCase.expected) {
				t.Fatalf("Unexpected result. Wanted '%v', retrieved '%v'",
					testCase.expected, result)
			}
		})
	}
}

func TestBrokerCreateHandler(t *testing.T) {
	testCases := []struct {
		label         string
		input         io.Reader
		expected      brokerPOSTResponse
		expectedError string
		expectedCode  int
		instClient    *mockInstanceClient
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
			label:         "Missing vf-module-*-id parameter",
			expectedError: "vf-module-model-customization-id is empty",
			expectedCode:  http.StatusBadRequest,
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-invariant-id": "123456qwerty",
				"vf-module-model-version-id": "123qweasdzxc",
				"generic-vnf-id": "dummy-vnf-id",
				"vf-module-id": "dummy-vfm-id",
				"template_data": {
					"stack_name": "dummy-stack-name"
				},
				"sdnc_directives": {
					"attributes": [
						{
							"attribute_name": "k8s-rb-profile-name",
							"attribute_value": "dummy-profile"
						}
					]
				}
			}`)),
		},
		{
			label:         "Missing stack name parameter",
			expectedError: "stack_name is missing from template_data",
			expectedCode:  http.StatusBadRequest,
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-customization-id": "84sdfkio938",
				"vf-module-model-invariant-id": "123456qwerty",
				"vf-module-model-version-id": "123qweasdzxc",
				"generic-vnf-id": "dummy-vnf-id",
				"vf-module-id": "dummy-vfm-id",
				"template_data": {
				},
				"sdnc_directives": {
					"attributes": [
						{
							"attribute_name": "k8s-rb-profile-name",
							"attribute_value": "dummy-profile"
						}
					]
				}
			}`)),
		},
		{
			label:         "Missing profile name directive",
			expectedError: "k8s-rb-profile-name is missing from directives",
			expectedCode:  http.StatusBadRequest,
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-customization-id": "84sdfkio938",
				"vf-module-model-invariant-id": "123456qwerty",
				"vf-module-model-version-id": "123qweasdzxc",
				"generic-vnf-id": "dummy-vnf-id",
				"vf-module-id": "dummy-vfm-id",
				"template_data": {
					"stack_name": "dummy-stack-name"
				},
				"sdnc_directives": {
					"attributes": [
					]
				}
			}`)),
		},
		{
			label:         "Missing vf-module-id parameter",
			expectedError: "vf-module-id is empty",
			expectedCode:  http.StatusBadRequest,
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-customization-id": "84sdfkio938",
				"vf-module-model-invariant-id": "123456qwerty",
				"vf-module-model-version-id": "123qweasdzxc",
				"generic-vnf-id": "dummy-vnf-id",
				"template_data": {
					"stack_name": "dummy-stack-name"
				},
				"sdnc_directives": {
					"attributes": [
						{
							"attribute_name": "k8s-rb-profile-name",
							"attribute_value": "dummy-profile"
						}
					]
				}
			}`)),
		},
		{
			label: "Succesfully create an Instance",
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-customization-id": "84sdfkio938",
				"vf-module-model-invariant-id": "123456qwerty",
				"vf-module-model-version-id": "123qweasdzxc",
				"generic-vnf-id": "dummy-vnf-id",
				"vf-module-id": "dummy-vfm-id",
				"template_data": {
					"stack_name": "dummy-stack-name"
				},
				"sdnc_directives": {
					"attributes": [
						{
							"attribute_name": "k8s-rb-profile-name",
							"attribute_value": "dummy-profile"
						}
					]
				}
			}`)),
			expected: brokerPOSTResponse{
				WorkloadID:     "HaKpys8e",
				TemplateType:   "heat",
				WorkloadStatus: "CREATE_COMPLETE",
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
						ID: "HaKpys8e",
						Request: app.InstanceRequest{
							RBName:      "123456qwerty",
							RBVersion:   "123qweasdzxc",
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

			request := httptest.NewRequest("POST", "/cloudowner/cloudregion/infra_workload", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil, nil, nil))
			defer resp.Body.Close()

			if testCase.expectedCode != resp.StatusCode {
				body, _ := ioutil.ReadAll(resp.Body)
				t.Log(string(body))
				t.Fatalf("Request method returned code '%v', but '%v' was expected",
					resp.StatusCode, testCase.expectedCode)
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
			} else if testCase.expectedError != "" {
				body, err := ioutil.ReadAll(resp.Body)
				if err == nil {
					if !strings.Contains(string(body), testCase.expectedError) {
						t.Fatalf("Request method returned body '%s', but '%s' wasn't found",
							body, testCase.expectedError)
					}
				} else {
					t.Fatalf("Request method returned malformed body")
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
				WorkloadStatus: "CREATE_COMPLETE",
			},
			instClient: &mockInstanceClient{
				items: []app.InstanceResponse{
					{
						ID: "HaKpys8e",
						Request: app.InstanceRequest{
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
			request := httptest.NewRequest("GET", "/cloudowner/cloudregion/infra_workload/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil, nil, nil))

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

func TestBrokerFindHandler(t *testing.T) {
	testCases := []struct {
		label            string
		input            string
		expectedCode     int
		expectedResponse brokerGETResponse
		instClient       *mockInstanceClient
	}{
		{
			label:        "Successful find an Instance",
			input:        "test-vf-module-name",
			expectedCode: http.StatusOK,
			expectedResponse: brokerGETResponse{
				TemplateType:   "heat",
				WorkloadID:     "HaKpys8e",
				WorkloadStatus: "CREATE_COMPLETE",
				WorkloadStatusReason: map[string]interface{}{
					"stacks": []map[string]interface{}{
						{
							"stack_status": "CREATE_COMPLETE",
							"id":           "HaKpys8e",
						},
					},
				},
			},
			instClient: &mockInstanceClient{
				miniitems: []app.InstanceMiniResponse{
					{
						ID: "HaKpys8e",
						Request: app.InstanceRequest{
							RBName:      "test-rbdef",
							RBVersion:   "v1",
							ProfileName: "profile1",
							CloudRegion: "region1",
						},
						Namespace: "testnamespace",
					},
				},
			},
		},
		{
			label:        "Fail to find an Instance",
			input:        "test-vf-module-name-1",
			expectedCode: http.StatusOK,
			expectedResponse: brokerGETResponse{
				TemplateType:   "heat",
				WorkloadID:     "",
				WorkloadStatus: "GET_COMPLETE",
				WorkloadStatusReason: map[string]interface{}{
					"stacks": []map[string]interface{}{},
				},
			},
			instClient: &mockInstanceClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/cloudowner/cloudregion/infra_workload?name="+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil, nil, nil))

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
				if testCase.expectedResponse.WorkloadID != response.WorkloadID {
					t.Fatalf("TestGetHandler returned:\n result=%v\n expected=%v",
						response.WorkloadID, testCase.expectedResponse.WorkloadID)
				}
				tcStacks := testCase.expectedResponse.WorkloadStatusReason["stacks"].([]map[string]interface{})
				if len(tcStacks) != 0 {
					//We expect only one response in this testcase.
					resStacks := response.WorkloadStatusReason["stacks"].([]interface{})[0].(map[string]interface{})
					if !reflect.DeepEqual(tcStacks[0], resStacks) {
						t.Fatalf("TestGetHandler returned:\n result=%v\n expected=%v",
							resStacks, tcStacks)
					}
				}
			}
		})
	}
}

func TestBrokerDeleteHandler(t *testing.T) {
	testCases := []struct {
		label            string
		input            string
		expectedCode     int
		expectedResponse brokerDELETEResponse
		instClient       *mockInstanceClient
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
			expectedResponse: brokerDELETEResponse{
				TemplateType:   "heat",
				WorkloadID:     "HaKpys8e",
				WorkloadStatus: "DELETE_COMPLETE",
			},
			instClient: &mockInstanceClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/cloudowner/cloudregion/infra_workload/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil, nil, nil))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
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
