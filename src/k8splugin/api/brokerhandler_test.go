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

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"

	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestConvertAttribute(t *testing.T) {
	testCases := []struct {
		label    string
		input    map[string]interface{}
		expected [2]string
	}{
		{
			label: "Correct conversion",
			input: map[string]interface{}{
				"attribute_name":  "test",
				"attribute_value": "value",
			},
			expected: [2]string{"test", "value"},
		},
		{
			label: "Empty valuee",
			input: map[string]interface{}{
				"attribute_name":  "test2",
				"attribute_value": "",
			},
			expected: [2]string{"test2", ""},
		},
		{
			label: "Malformed attribute",
			input: map[string]interface{}{
				"attribute_nam":   "test123",
				"attribute_value": "value123",
			},
			expected: [2]string{"", ""},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			var result [2]string
			result[0], result[1] = convertAttribute(testCase.input)
			if result != testCase.expected {
				t.Fatalf("Expected %v, aquired %v", testCase.expected, result)
			}
		})
	}
}

func TestExtractAttributes(t *testing.T) {
	testCases := []struct {
		label    string
		input    map[string]interface{}
		expected map[string]string
	}{
		{
			label:    "Single variable",
			expected: map[string]string{"test": "true"},
			input: map[string]interface{}{"attributes": []map[string]interface{}{{
				"attribute_name":  "test",
				"attribute_value": "true",
			}}},
		},
		{
			label:    "Empty parameter",
			expected: map[string]string{"test": ""},
			input: map[string]interface{}{"attributes": []map[string]interface{}{{
				"attribute_name":  "test",
				"attribute_value": "",
			}}},
		},
		{
			label:    "Null entry",
			input:    map[string]interface{}{"attributes": nil},
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
			input: map[string]interface{}{"attributes": []map[string]interface{}{
				{
					"attribute_name":  "name",
					"attribute_value": "{a, b, c}",
				},
				{
					"attribute_name":  "servers[0].port",
					"attribute_value": "80",
				},
			}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			result := make(map[string]string)
			extractAttributes(testCase.input, &result)
			if !reflect.DeepEqual(result, testCase.expected) {
				t.Fatalf("Unexpected result. Wanted '%v', retrieved '%v'",
					testCase.expected, result)
			}
		})
	}
}

func TestGetOverrideParams(t *testing.T) {

	testCases := []struct {
		label    string
		expected map[string]string
		input    brokerRequest
	}{
		{
			label:    "Excessive parameters",
			expected: map[string]string{"var1": "1", "var2": "2"},
			input: brokerRequest{
				SDNCDirectives: map[string]interface{}{
					"attributes": []map[string]interface{}{
						{
							"attribute_name":  "var1",
							"attribute_value": "1",
						},
						{
							"attribute_name":  "var2",
							"attribute_value": "2",
						},
						{
							"attribute_name":  "k8s-rb-profile-name",
							"attribute_value": "default",
						},
					},
				},
			},
		},
		{
			label:    "Override variables",
			expected: map[string]string{"empty": "", "sdnc": "sdnc", "user": "user", "oof": "oof"},
			input: brokerRequest{
				SDNCDirectives: map[string]interface{}{
					"attributes": []map[string]interface{}{
						{
							"attribute_name":  "empty",
							"attribute_value": "sdnc",
						},
						{
							"attribute_name":  "sdnc",
							"attribute_value": "sdnc",
						},
						{
							"attribute_name":  "oof",
							"attribute_value": "sdnc",
						},
					},
				},
				OOFDirectives: map[string]interface{}{
					"attributes": []map[string]interface{}{
						{
							"attribute_name":  "oof",
							"attribute_value": "oof",
						},
						{
							"attribute_name":  "user",
							"attribute_value": "oof",
						},
					},
				},
				UserDirectives: map[string]interface{}{
					"attributes": []map[string]interface{}{
						{
							"attribute_name":  "user",
							"attribute_value": "user",
						},
						{
							"attribute_name":  "empty",
							"attribute_value": "",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			result := testCase.input.getOverrideParams()
			if !reflect.DeepEqual(result, testCase.expected) {
				t.Fatalf("Unexpected result. Wanted '%v', retrieved '%v'",
					testCase.expected, result)
			}
		})
	}
}

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
			label: "Missing vf-module-*-id parameter",
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-customization-id": "84sdfkio938",
				"vf-module-model-invariant-id": "123456qwerty",
				"sdnc_directives": {
					"attributes": [
						{
							"attribute_name": "vf_module_name",
							"attribute_value": "test-vf-module-name"
						},
						{
							"attribute_name": "k8s-rb-profile-name",
							"attribute_value": "profile1"
						}
					]
				}
			}`)),
			expectedCode: http.StatusBadRequest,
		},
		{
			label: "Missing parameter from sdnc_directives",
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-customization-id": "84sdfkio938",
				"vf-module-model-invariant-id": "123456qwerty",
				"vf-module-model-version-id": "123qweasdzxc",
				"sdnc_directives": {
					"attributes": [
						{
							"attribute_name": "vf_module_name",
							"attribute_value": "test-vf-module-name"
						}
					]
				}
			}`)),
			expectedCode: http.StatusBadRequest,
		},
		{
			label: "Succesfully create an Instance",
			input: bytes.NewBuffer([]byte(`{
				"vf-module-model-customization-id": "84sdfkio938",
				"vf-module-model-invariant-id": "123456qwerty",
				"vf-module-model-version-id": "123qweasdzxc",
				"sdnc_directives": {
					"attributes": [
						{
							"attribute_name": "vf_module_name",
							"attribute_value": "test-vf-module-name"
						},
						{
							"attribute_name": "k8s-rb-profile-name",
							"attribute_value": "profile1"
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
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

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
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

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
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

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
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

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
