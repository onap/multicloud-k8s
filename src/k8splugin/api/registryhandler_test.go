/*
Copyright 2019 Intel Corporation.
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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockRegistryClient struct {
	app.RegistryManager
	// Items and err will be used to customize each test
	// via a localized instantiation of mockInstanceClient
	items      []app.RegistryResponse
	err        error
}

func (m *mockRegistryClient) Create(inp app.RegistryRequest) (app.RegistryResponse, error) {
	if m.err != nil {
		return app.RegistryResponse{}, m.err
	}

	return m.items[0], nil
}

func (m *mockRegistryClient) Get(id string) (app.RegistryResponse, error) {
	if m.err != nil {
		return app.RegistryResponse{}, m.err
	}

	return m.items[0], nil
}

func (m *mockRegistryClient) Delete(id string) error {
	return m.err
}

func TestRegistryCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expected     app.RegistryResponse
		expectedCode int
		regClient   *mockRegistryClient
	}{
		{
			label: "Registry failure",
			expectedCode: http.StatusInternalServerError,
			regClient: &mockRegistryClient{
				items: []app.RegistryResponse{
					{
						ID: "cocky_buck",
						Request: app.RegistryRequest{
							CloudOwner:  "INTEL",
							CloudRegion: "RegionOne",
						},
					},
				},
			},
		},
		{
			label: "Succesfully Registry an kubernetes cluster",
			expectedCode: http.StatusCreated,
			regClient: &mockRegistryClient{
				items: []app.RegistryResponse{
					{
						ID: "cocky_buck",
						Request: app.RegistryRequest{
							CloudOwner:  "INTEL",
							CloudRegion: "RegionOne",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/INTEL/RegionOne/registry", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, testCase.regClient))

			if testCase.expectedCode != resp.StatusCode {
				body, _ := ioutil.ReadAll(resp.Body)
				t.Log(string(body))
				t.Logf("Request method returned: \n%v\n and it was expected: \n%v", resp.StatusCode, testCase.expectedCode)
			} else {
				if resp.StatusCode == http.StatusCreated {
					var response app.RegistryResponse
					err := json.NewDecoder(resp.Body).Decode(&response)
					if err != nil {
						t.Fatalf("Parsing the returned response got an error (%s)", err)
					}
				}
			}
		},)
	}
}

func TestRegistryGetHandler(t *testing.T) {
	testCases := []struct {
		label            string
		input            string
		expectedCode     int
		expectedResponse *app.RegistryResponse
		regClient       *mockRegistryClient
	}{
		{
			label:        "Fail to retrieve registry info",
			input:        "cocky_buck",
			expectedCode: http.StatusInternalServerError,
			regClient: &mockRegistryClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful get an registry info",
			input:        "cocky_buck",
			expectedCode: http.StatusOK,
			expectedResponse: &app.RegistryResponse{
				ID: "cocky_buck",
				Request: app.RegistryRequest{
					CloudOwner:  "INTEL",
					CloudRegion: "RegionOne",
				},
			},
			regClient: &mockRegistryClient{
				items: []app.RegistryResponse{
					{
						ID: "cocky_buck",
						Request: app.RegistryRequest{
							CloudOwner:  "INTEL",
							CloudRegion: "RegionOne",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/INTEL/RegionOne/registry/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, testCase.regClient))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v",
					resp.StatusCode, testCase.expectedCode)
			}
			if resp.StatusCode == http.StatusOK {
				var response app.RegistryResponse
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

func TestRegistryDeleteHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        string
		expectedCode int
		regClient   *mockRegistryClient
	}{
		{
			label:        "Fail to Registry a kubernetes",
			input:        "cocky_buck",
			expectedCode: http.StatusInternalServerError,
			regClient: &mockRegistryClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Succesful delete a kubernetes",
			input:        "cocky_buck",
			expectedCode: http.StatusAccepted,
			regClient:   &mockRegistryClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v1/INTEL/RegionOne/registry/"+testCase.input, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, testCase.regClient))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}
