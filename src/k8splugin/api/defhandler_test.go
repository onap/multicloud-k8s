/*
 * Copyright 2018 Intel Corporation, Inc
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
	"sort"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockRBDefinition struct {
	rb.DefinitionManager
	// Items and err will be used to customize each test
	// via a localized instantiation of mockRBDefinition
	Items []rb.Definition
	Err   error
}

func (m *mockRBDefinition) Create(inp rb.Definition) (rb.Definition, error) {
	if m.Err != nil {
		return rb.Definition{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockRBDefinition) List(name string) ([]rb.Definition, error) {
	if m.Err != nil {
		return []rb.Definition{}, m.Err
	}

	return m.Items, nil
}

func (m *mockRBDefinition) Get(name, version string) (rb.Definition, error) {
	if m.Err != nil {
		return rb.Definition{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockRBDefinition) Delete(name, version string) error {
	return m.Err
}

func (m *mockRBDefinition) Upload(name, version string, inp []byte) error {
	return m.Err
}

func TestRBDefCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     rb.Definition
		expectedCode int
		rbDefClient  *mockRBDefinition
	}{
		{
			label:        "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			rbDefClient:  &mockRBDefinition{},
		},
		{
			label:        "Create Definition",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"rb-name":"testresourcebundle",
				"rb-version":"v1",
				"chart-name":"testchart",
				"description":"test description"
				}`)),
			expected: rb.Definition{
				RBName:      "testresourcebundle",
				RBVersion:   "v1",
				ChartName:   "testchart",
				Description: "test description",
			},
			rbDefClient: &mockRBDefinition{
				//Items that will be returned by the mocked Client
				Items: []rb.Definition{
					{
						RBName:      "testresourcebundle",
						RBVersion:   "v1",
						ChartName:   "testchart",
						Description: "test description",
					},
				},
			},
		},
		{
			label: "Missing Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"rb-version":"v1",
				"chart-name":"testchart",
				"description":"test description"
				}`)),
			expectedCode: http.StatusBadRequest,
			rbDefClient:  &mockRBDefinition{},
		},
		{
			label: "Missing Version in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"rb-name":"testresourcebundle",
				"chart-name":"testchart",
				"description":"test description"
				}`)),
			expectedCode: http.StatusBadRequest,
			rbDefClient:  &mockRBDefinition{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/rb/definition", testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.rbDefClient, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := rb.Definition{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestRBDefListVersionsHandler(t *testing.T) {

	testCases := []struct {
		label        string
		expected     []rb.Definition
		expectedCode int
		rbDefClient  *mockRBDefinition
	}{
		{
			label:        "List Bundle Definitions",
			expectedCode: http.StatusOK,
			expected: []rb.Definition{
				{
					RBName:      "testresourcebundle",
					RBVersion:   "v1",
					ChartName:   "testchart",
					Description: "test description",
				},
				{
					RBName:      "testresourcebundle",
					RBVersion:   "v2",
					ChartName:   "testchart",
					Description: "test description",
				},
			},
			rbDefClient: &mockRBDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []rb.Definition{
					{
						RBName:      "testresourcebundle",
						RBVersion:   "v1",
						ChartName:   "testchart",
						Description: "test description",
					},
					{
						RBName:      "testresourcebundle",
						RBVersion:   "v2",
						ChartName:   "testchart",
						Description: "test description",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/rb/definition/testresourcebundle", nil)
			resp := executeRequest(request, NewRouter(testCase.rbDefClient, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []rb.Definition{}
				json.NewDecoder(resp.Body).Decode(&got)

				// Since the order of returned slice is not guaranteed
				// Check both and return error if both don't match
				sort.Slice(got, func(i, j int) bool {
					return got[i].RBVersion < got[j].RBVersion
				})
				// Sort both as it is not expected that testCase.expected
				// is sorted
				sort.Slice(testCase.expected, func(i, j int) bool {
					return testCase.expected[i].RBVersion < testCase.expected[j].RBVersion
				})

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestRBDefListAllHandler(t *testing.T) {

	testCases := []struct {
		label        string
		expected     []rb.Definition
		expectedCode int
		rbDefClient  *mockRBDefinition
	}{
		{
			label:        "List Bundle Definitions",
			expectedCode: http.StatusOK,
			expected: []rb.Definition{
				{
					RBName:      "resourcebundle1",
					RBVersion:   "v1",
					ChartName:   "barchart",
					Description: "test description for one",
				},
				{
					RBName:      "resourcebundle2",
					RBVersion:   "version2",
					ChartName:   "foochart",
					Description: "test description for two",
				},
			},
			rbDefClient: &mockRBDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []rb.Definition{
					{
						RBName:      "resourcebundle1",
						RBVersion:   "v1",
						ChartName:   "barchart",
						Description: "test description for one",
					},
					{
						RBName:      "resourcebundle2",
						RBVersion:   "version2",
						ChartName:   "foochart",
						Description: "test description for two",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/rb/definition", nil)
			resp := executeRequest(request, NewRouter(testCase.rbDefClient, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []rb.Definition{}
				json.NewDecoder(resp.Body).Decode(&got)

				// Since the order of returned slice is not guaranteed
				// Check both and return error if both don't match
				sort.Slice(got, func(i, j int) bool {
					return got[i].RBVersion < got[j].RBVersion
				})
				// Sort both as it is not expected that testCase.expected
				// is sorted
				sort.Slice(testCase.expected, func(i, j int) bool {
					return testCase.expected[i].RBVersion < testCase.expected[j].RBVersion
				})

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestRBDefGetHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      rb.Definition
		name, version string
		expectedCode  int
		rbDefClient   *mockRBDefinition
	}{
		{
			label:        "Get Bundle Definition",
			expectedCode: http.StatusOK,
			expected: rb.Definition{
				RBName:      "testresourcebundle",
				RBVersion:   "v1",
				ChartName:   "testchart",
				Description: "test description",
			},
			name:    "testresourcebundle",
			version: "v1",
			rbDefClient: &mockRBDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []rb.Definition{
					{
						RBName:      "testresourcebundle",
						RBVersion:   "v1",
						ChartName:   "testchart",
						Description: "test description",
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting Bundle Definition",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingbundle",
			version:      "v1",
			rbDefClient: &mockRBDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []rb.Definition{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/rb/definition/"+testCase.name+"/"+testCase.version, nil)
			resp := executeRequest(request, NewRouter(testCase.rbDefClient, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := rb.Definition{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestRBDefDeleteHandler(t *testing.T) {

	testCases := []struct {
		label        string
		name         string
		version      string
		expectedCode int
		rbDefClient  *mockRBDefinition
	}{
		{
			label:        "Delete Bundle Definition",
			expectedCode: http.StatusNoContent,
			name:         "test-rbdef",
			version:      "v1",
			rbDefClient:  &mockRBDefinition{},
		},
		{
			label:        "Delete Non-Exiting Bundle Definition",
			expectedCode: http.StatusInternalServerError,
			name:         "test-rbdef",
			version:      "v2",
			rbDefClient: &mockRBDefinition{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v1/rb/definition/"+testCase.name+"/"+testCase.version, nil)
			resp := executeRequest(request, NewRouter(testCase.rbDefClient, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestRBDefUploadHandler(t *testing.T) {

	testCases := []struct {
		label        string
		name         string
		version      string
		body         io.Reader
		expectedCode int
		rbDefClient  *mockRBDefinition
	}{
		{
			label:        "Upload Bundle Definition Content",
			expectedCode: http.StatusOK,
			name:         "test-rbdef",
			version:      "v2",
			body: bytes.NewBuffer([]byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			}),
			rbDefClient: &mockRBDefinition{},
		},
		{
			label:        "Upload Invalid Bundle Definition Content",
			expectedCode: http.StatusInternalServerError,
			name:         "test-rbdef",
			version:      "v2",
			body: bytes.NewBuffer([]byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			}),
			rbDefClient: &mockRBDefinition{
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Upload Empty Body Content",
			expectedCode: http.StatusBadRequest,
			name:         "test-rbdef",
			version:      "v2",
			rbDefClient:  &mockRBDefinition{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST",
				"/v1/rb/definition/"+testCase.name+"/"+testCase.version+"/content", testCase.body)
			resp := executeRequest(request, NewRouter(testCase.rbDefClient, nil, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
