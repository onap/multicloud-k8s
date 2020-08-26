/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
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

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/types"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockControllerManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockControllerManager
	Items []controller.Controller
	Err   error
}

func (m *mockControllerManager) CreateController(inp controller.Controller, mayExist bool) (controller.Controller, error) {
	if m.Err != nil {
		return controller.Controller{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockControllerManager) GetController(name string) (controller.Controller, error) {
	if m.Err != nil {
		return controller.Controller{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockControllerManager) GetControllers() ([]controller.Controller, error) {
	if m.Err != nil {
		return []controller.Controller{}, m.Err
	}

	return m.Items, nil
}

func (m *mockControllerManager) DeleteController(name string) error {
	return m.Err
}

func (m *mockControllerManager) InitControllers() {
	return
}

func init() {
	controllerJSONFile = "../json-schemas/controller.json"
}

func TestControllerCreateHandler(t *testing.T) {
	testCases := []struct {
		label            string
		reader           io.Reader
		expected         controller.Controller
		expectedCode     int
		controllerClient *mockControllerManager
	}{
		{
			label:            "Missing Body Failure",
			expectedCode:     http.StatusBadRequest,
			controllerClient: &mockControllerManager{},
		},
		{
			label:        "Create Controller",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
                "metadata": {
				"name":"testController"
				},
				"spec": {
				"ip-address":"10.188.234.1",
				"port":8080 }
				}`)),
			expected: controller.Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: controller.ControllerSpec{
					Host: "10.188.234.1",
					Port: 8080,
				},
			},
			controllerClient: &mockControllerManager{
				//Items that will be returned by the mocked Client
				Items: []controller.Controller{
					{
						Metadata: types.Metadata{
							Name: "testController",
						},
						Spec: controller.ControllerSpec{
							Host: "10.188.234.1",
							Port: 8080,
						},
					},
				},
			},
		},
		{
			label: "Missing Controller Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			expectedCode:     http.StatusBadRequest,
			controllerClient: &mockControllerManager{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/controllers", testCase.reader)
			resp := executeRequest(request, NewRouter(nil, nil, nil, testCase.controllerClient, nil, nil, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := controller.Controller{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestControllerGetHandler(t *testing.T) {

	testCases := []struct {
		label            string
		expected         controller.Controller
		name, version    string
		expectedCode     int
		controllerClient *mockControllerManager
	}{
		{
			label:        "Get Controller",
			expectedCode: http.StatusOK,
			expected: controller.Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: controller.ControllerSpec{
					Host: "10.188.234.1",
					Port: 8080,
				},
			},
			name: "testController",
			controllerClient: &mockControllerManager{
				Items: []controller.Controller{
					{
						Metadata: types.Metadata{
							Name: "testController",
						},
						Spec: controller.ControllerSpec{
							Host: "10.188.234.1",
							Port: 8080,
						},
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Controller",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingController",
			controllerClient: &mockControllerManager{
				Items: []controller.Controller{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/controllers/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, testCase.controllerClient, nil, nil, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := controller.Controller{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestControllerDeleteHandler(t *testing.T) {

	testCases := []struct {
		label            string
		name             string
		version          string
		expectedCode     int
		controllerClient *mockControllerManager
	}{
		{
			label:            "Delete Controller",
			expectedCode:     http.StatusNoContent,
			name:             "testController",
			controllerClient: &mockControllerManager{},
		},
		{
			label:        "Delete Non-Existing Controller",
			expectedCode: http.StatusInternalServerError,
			name:         "testController",
			controllerClient: &mockControllerManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/controllers/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, testCase.controllerClient, nil, nil, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
