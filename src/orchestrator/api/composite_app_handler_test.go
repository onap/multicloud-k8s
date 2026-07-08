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

type mockCompositeAppManager struct {
	Items []moduleLib.CompositeApp
	Err   error
}

func (m *mockCompositeAppManager) CreateCompositeApp(c moduleLib.CompositeApp, p string) (moduleLib.CompositeApp, error) {
	if m.Err != nil {
		return moduleLib.CompositeApp{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockCompositeAppManager) GetCompositeApp(name, version, p string) (moduleLib.CompositeApp, error) {
	if m.Err != nil {
		return moduleLib.CompositeApp{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockCompositeAppManager) GetAllCompositeApps(p string) ([]moduleLib.CompositeApp, error) {
	if m.Err != nil {
		return []moduleLib.CompositeApp{}, m.Err
	}
	return m.Items, nil
}

func (m *mockCompositeAppManager) DeleteCompositeApp(name, version, p string) error {
	return m.Err
}

func init() {
	caJSONFile = "../json-schemas/composite-app.json"
}

func TestCompositeAppCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     moduleLib.CompositeApp
		expectedCode int
		client       *mockCompositeAppManager
	}{
		{
			label:        "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			client:       &mockCompositeAppManager{},
		},
		{
			label:        "Create CompositeApp",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {
					"name": "testCompositeApp",
					"description": "Test CompositeApp"
				},
				"spec": {
					"version": "v1"
				}
			}`)),
			expected: moduleLib.CompositeApp{
				Metadata: moduleLib.CompositeAppMetaData{
					Name:        "testCompositeApp",
					Description: "Test CompositeApp",
				},
				Spec: moduleLib.CompositeAppSpec{Version: "v1"},
			},
			client: &mockCompositeAppManager{
				Items: []moduleLib.CompositeApp{
					{
						Metadata: moduleLib.CompositeAppMetaData{
							Name:        "testCompositeApp",
							Description: "Test CompositeApp",
						},
						Spec: moduleLib.CompositeAppSpec{Version: "v1"},
					},
				},
			},
		},
		{
			label:        "Manager Error",
			expectedCode: http.StatusInternalServerError,
			reader: bytes.NewBuffer([]byte(`{
				"metadata": {"name": "testCompositeApp"},
				"spec": {"version": "v1"}
			}`)),
			client: &mockCompositeAppManager{Err: pkgerrors.New("Internal Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/testProject/composite-apps", testCase.reader)
			resp := executeRequest(request, NewRouter(nil, testCase.client, nil, nil, nil, nil, nil, nil, nil, nil, nil))

			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusCreated {
				got := moduleLib.CompositeApp{}
				json.NewDecoder(resp.Body).Decode(&got)
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestCompositeAppGetHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expected     moduleLib.CompositeApp
		expectedCode int
		client       *mockCompositeAppManager
	}{
		{
			label:        "Get CompositeApp",
			expectedCode: http.StatusOK,
			expected: moduleLib.CompositeApp{
				Metadata: moduleLib.CompositeAppMetaData{Name: "testCompositeApp"},
				Spec:     moduleLib.CompositeAppSpec{Version: "v1"},
			},
			client: &mockCompositeAppManager{
				Items: []moduleLib.CompositeApp{
					{
						Metadata: moduleLib.CompositeAppMetaData{Name: "testCompositeApp"},
						Spec:     moduleLib.CompositeAppSpec{Version: "v1"},
					},
				},
			},
		},
		{
			label:        "Get Non-Existing CompositeApp",
			expectedCode: http.StatusInternalServerError,
			client:       &mockCompositeAppManager{Err: pkgerrors.New("Internal Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/testProject/composite-apps/testCompositeApp/v1", nil)
			resp := executeRequest(request, NewRouter(nil, testCase.client, nil, nil, nil, nil, nil, nil, nil, nil, nil))

			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
			if resp.StatusCode == http.StatusOK {
				got := moduleLib.CompositeApp{}
				json.NewDecoder(resp.Body).Decode(&got)
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("getHandler returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestCompositeAppGetAllHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		client       *mockCompositeAppManager
	}{
		{
			label:        "Get All CompositeApps",
			expectedCode: http.StatusOK,
			client: &mockCompositeAppManager{
				Items: []moduleLib.CompositeApp{
					{Metadata: moduleLib.CompositeAppMetaData{Name: "ca1"}},
				},
			},
		},
		{
			label:        "Get All Error",
			expectedCode: http.StatusNotFound,
			client:       &mockCompositeAppManager{Err: pkgerrors.New("Internal Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/projects/testProject/composite-apps", nil)
			resp := executeRequest(request, NewRouter(nil, testCase.client, nil, nil, nil, nil, nil, nil, nil, nil, nil))
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestCompositeAppDeleteHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		client       *mockCompositeAppManager
	}{
		{
			label:        "Delete CompositeApp",
			expectedCode: http.StatusNoContent,
			client: &mockCompositeAppManager{
				Items: []moduleLib.CompositeApp{
					{Metadata: moduleLib.CompositeAppMetaData{Name: "testCompositeApp"}},
				},
			},
		},
		{
			label:        "Delete Non-Existing CompositeApp",
			expectedCode: http.StatusNotFound,
			client:       &mockCompositeAppManager{Err: pkgerrors.New("Internal Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/projects/testProject/composite-apps/testCompositeApp/v1", nil)
			resp := executeRequest(request, NewRouter(nil, testCase.client, nil, nil, nil, nil, nil, nil, nil, nil, nil))
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
