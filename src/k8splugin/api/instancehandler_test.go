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
	"sort"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockInstanceClient struct {
	app.InstanceManager
	// Items and err will be used to customize each test
	// via a localized instantiation of mockInstanceClient
	items      []app.InstanceResponse
	miniitems  []app.InstanceMiniResponse
	statusItem app.InstanceStatus
	err        error
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

func (m *mockInstanceClient) Status(id string) (app.InstanceStatus, error) {
	if m.err != nil {
		return app.InstanceStatus{}, m.err
	}

	return m.statusItem, nil
}

func (m *mockInstanceClient) List(rbname, rbversion, profilename string) ([]app.InstanceMiniResponse, error) {
	if m.err != nil {
		return []app.InstanceMiniResponse{}, m.err
	}

	return m.miniitems, nil
}

func (m *mockInstanceClient) Find(rbName string, ver string, profile string, labelKeys map[string]string) ([]app.InstanceMiniResponse, error) {
	if m.err != nil {
		return nil, m.err
	}

	return m.miniitems, nil
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
			expectedCode: http.StatusCreated,
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

			request := httptest.NewRequest("POST", "/v1/instance", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

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
			request := httptest.NewRequest("GET", "/v1/instance/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

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

func TestStatusHandler(t *testing.T) {
	testCases := []struct {
		label            string
		input            string
		expectedCode     int
		expectedResponse *app.InstanceStatus
		instClient       *mockInstanceClient
	}{
		{
			label:        "Fail to Get Status",
			input:        "HaKpys8e",
			expectedCode: http.StatusInternalServerError,
			instClient: &mockInstanceClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Succesful GET Status",
			input:        "HaKpys8e",
			expectedCode: http.StatusOK,
			expectedResponse: &app.InstanceStatus{
				Request: app.InstanceRequest{
					RBName:      "test-rbdef",
					RBVersion:   "v1",
					ProfileName: "profile1",
					CloudRegion: "region1",
				},
				Ready:         true,
				ResourceCount: 2,
				PodStatuses: []app.PodStatus{
					{
						Name:        "test-pod1",
						Namespace:   "default",
						Ready:       true,
						IPAddresses: []string{"192.168.1.1", "192.168.2.1"},
					},
					{
						Name:        "test-pod2",
						Namespace:   "default",
						Ready:       true,
						IPAddresses: []string{"192.168.3.1", "192.168.5.1"},
					},
				},
			},
			instClient: &mockInstanceClient{
				statusItem: app.InstanceStatus{
					Request: app.InstanceRequest{
						RBName:      "test-rbdef",
						RBVersion:   "v1",
						ProfileName: "profile1",
						CloudRegion: "region1",
					},
					Ready:         true,
					ResourceCount: 2,
					PodStatuses: []app.PodStatus{
						{
							Name:        "test-pod1",
							Namespace:   "default",
							Ready:       true,
							IPAddresses: []string{"192.168.1.1", "192.168.2.1"},
						},
						{
							Name:        "test-pod2",
							Namespace:   "default",
							Ready:       true,
							IPAddresses: []string{"192.168.3.1", "192.168.5.1"},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/"+testCase.input+"/status", nil)
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestInstanceListHandler(t *testing.T) {
	testCases := []struct {
		label            string
		input            string
		expectedCode     int
		queryParams      bool
		queryParamsMap   map[string]string
		expectedResponse []app.InstanceMiniResponse
		instClient       *mockInstanceClient
	}{
		{
			label:        "Fail to List Instance",
			input:        "HaKpys8e",
			expectedCode: http.StatusInternalServerError,
			instClient: &mockInstanceClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Succesful List Instances",
			expectedCode: http.StatusOK,
			expectedResponse: []app.InstanceMiniResponse{
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
				{
					ID: "HaKpys8f",
					Request: app.InstanceRequest{
						RBName:      "test-rbdef-two",
						RBVersion:   "versionsomething",
						ProfileName: "profile3",
						CloudRegion: "region1",
					},
					Namespace: "testnamespace-two",
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
					{
						ID: "HaKpys8f",
						Request: app.InstanceRequest{
							RBName:      "test-rbdef-two",
							RBVersion:   "versionsomething",
							ProfileName: "profile3",
							CloudRegion: "region1",
						},
						Namespace: "testnamespace-two",
					},
				},
			},
		},
		{
			label:       "List Instances Based on Query Parameters",
			queryParams: true,
			queryParamsMap: map[string]string{
				"rb-name": "test-rbdef1",
			},
			expectedCode: http.StatusOK,
			expectedResponse: []app.InstanceMiniResponse{
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance", nil)
			if testCase.queryParams {
				q := request.URL.Query()
				for k, v := range testCase.queryParamsMap {
					q.Add(k, v)
				}
				request.URL.RawQuery = q.Encode()
			}
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v",
					resp.StatusCode, testCase.expectedCode)
			}
			if resp.StatusCode == http.StatusOK {
				var response []app.InstanceMiniResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				if err != nil {
					t.Fatalf("Parsing the returned response got an error (%s)", err)
				}

				// Since the order of returned slice is not guaranteed
				// Sort them first and then do deepequal
				// Check both and return error if both don't match
				sort.Slice(response, func(i, j int) bool {
					return response[i].ID < response[j].ID
				})

				sort.Slice(testCase.expectedResponse, func(i, j int) bool {
					return testCase.expectedResponse[i].ID < testCase.expectedResponse[j].ID
				})

				if reflect.DeepEqual(testCase.expectedResponse, response) == false {
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
			resp := executeRequest(request, NewRouter(nil, nil, testCase.instClient, nil, nil, nil))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}
