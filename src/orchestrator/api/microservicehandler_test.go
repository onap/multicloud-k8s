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

	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockMicroserviceManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockMicroserviceManager
	Items []moduleLib.Microservice
	Err   error
}

func (m *mockMicroserviceManager) CreateMicroservice(inp moduleLib.Microservice) (moduleLib.Microservice, error) {
	if m.Err != nil {
		return moduleLib.Microservice{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockMicroserviceManager) GetMicroservice(name string) (moduleLib.Microservice, error) {
	if m.Err != nil {
		return moduleLib.Microservice{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockMicroserviceManager) DeleteMicroservice(name string) error {
	return m.Err
}

func TestMicroserviceCreateHandler(t *testing.T) {
	testCases := []struct {
		label              string
		reader             io.Reader
		expected           moduleLib.Microservice
		expectedCode       int
		microserviceClient *mockMicroserviceManager
	}{
		{
			label:              "Missing Body Failure",
			expectedCode:       http.StatusBadRequest,
			microserviceClient: &mockMicroserviceManager{},
		},
		{
			label:        "Create Microservice",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"name":"testMicroservice",
				"ip-address":"10.188.234.1",
				"port":8080
				}`)),
			expected: moduleLib.Microservice{
				Name:      "testMicroservice",
				IpAddress: "10.188.234.1",
				Port:      8080,
			},
			microserviceClient: &mockMicroserviceManager{
				//Items that will be returned by the mocked Client
				Items: []moduleLib.Microservice{
					{
						Name:      "testMicroservice",
						IpAddress: "10.188.234.1",
						Port:      8080,
					},
				},
			},
		},
		{
			label: "Missing Microservice Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			expectedCode:       http.StatusBadRequest,
			microserviceClient: &mockMicroserviceManager{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/register-microservices", testCase.reader)
			resp := executeRequest(request, NewRouter(nil, testCase.microserviceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := moduleLib.Microservice{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestMicroserviceGetHandler(t *testing.T) {

	testCases := []struct {
		label              string
		expected           moduleLib.Microservice
		name, version      string
		expectedCode       int
		microserviceClient *mockMicroserviceManager
	}{
		{
			label:        "Get Microservice",
			expectedCode: http.StatusOK,
			expected: moduleLib.Microservice{
				Name:      "testMicroservice",
				IpAddress: "10.188.234.1",
				Port:      8080,
			},
			name: "testMicroservice",
			microserviceClient: &mockMicroserviceManager{
				Items: []moduleLib.Microservice{
					{
						Name:      "testMicroservice",
						IpAddress: "10.188.234.1",
						Port:      8080,
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Microservice",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingMicroservice",
			microserviceClient: &mockMicroserviceManager{
				Items: []moduleLib.Microservice{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/register-microservices/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(nil, testCase.microserviceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := moduleLib.Microservice{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestMicroserviceDeleteHandler(t *testing.T) {

	testCases := []struct {
		label              string
		name               string
		version            string
		expectedCode       int
		microserviceClient *mockMicroserviceManager
	}{
		{
			label:              "Delete Microservice",
			expectedCode:       http.StatusNoContent,
			name:               "testMicroservice",
			microserviceClient: &mockMicroserviceManager{},
		},
		{
			label:        "Delete Non-Existing Microservice",
			expectedCode: http.StatusInternalServerError,
			name:         "testMicroservice",
			microserviceClient: &mockMicroserviceManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/register-microservices/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(nil, testCase.microserviceClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
