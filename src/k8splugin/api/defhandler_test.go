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
	"k8splugin/resource"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockBundleDefinition struct {
	resource.BundleDefInterface
	// Items and err will be used to customize each test
	// via a localized instantiation of mockBundleDefinition
	Items []resource.BundleDefinition
	Err   error
}

func (m *mockBundleDefinition) Create(inp resource.BundleDefinition) (resource.BundleDefinition, error) {
	if m.Err != nil {
		return resource.BundleDefinition{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockBundleDefinition) List() ([]resource.BundleDefinition, error) {
	if m.Err != nil {
		return []resource.BundleDefinition{}, m.Err
	}

	return m.Items, nil
}

func (m *mockBundleDefinition) Get(id string) (resource.BundleDefinition, error) {
	if m.Err != nil {
		return resource.BundleDefinition{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockBundleDefinition) Delete(id string) error {
	return m.Err
}

func TestBundleDefCreateHandler(t *testing.T) {
	testCases := []struct {
		label           string
		reader          io.Reader
		expected        resource.BundleDefinition
		expectedCode    int
		bundleDefClient *mockBundleDefinition
	}{
		{
			label:           "Missing Body Failure",
			expectedCode:    http.StatusBadRequest,
			bundleDefClient: &mockBundleDefinition{},
		},
		{
			label:        "Create without UUID",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"name":"testdomain",
				"description":"test description",
				"service-type":"firewall"
				}`)),
			expected: resource.BundleDefinition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testresourcebundle",
				Description: "test description",
				ServiceType: "firewall",
			},
			bundleDefClient: &mockBundleDefinition{
				//Items that will be returned by the mocked Client
				Items: []resource.BundleDefinition{
					{
						UUID:        "123e4567-e89b-12d3-a456-426655440000",
						Name:        "testresourcebundle",
						Description: "test description",
						ServiceType: "firewall",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := bundleDefHandler{client: testCase.bundleDefClient}
			req, err := http.NewRequest("POST", "/v1/resource/definition", testCase.reader)

			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.createHandler)
			hr.ServeHTTP(rr, req)

			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}

			//Check returned body only if statusCreated
			if rr.Code == http.StatusCreated {
				got := resource.BundleDefinition{}
				json.NewDecoder(rr.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestBundleDefListHandler(t *testing.T) {

	testCases := []struct {
		label           string
		expected        []resource.BundleDefinition
		expectedCode    int
		bundleDefClient *mockBundleDefinition
	}{
		{
			label:        "List Bundle Definitions",
			expectedCode: http.StatusOK,
			expected: []resource.BundleDefinition{
				{
					UUID:        "123e4567-e89b-12d3-a456-426655440000",
					Name:        "testresourcebundle",
					Description: "test description",
					ServiceType: "firewall",
				},
				{
					UUID:        "123e4567-e89b-12d3-a456-426655441111",
					Name:        "testresourcebundle2",
					Description: "test description",
					ServiceType: "dns",
				},
			},
			bundleDefClient: &mockBundleDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []resource.BundleDefinition{
					{
						UUID:        "123e4567-e89b-12d3-a456-426655440000",
						Name:        "testresourcebundle",
						Description: "test description",
						ServiceType: "firewall",
					},
					{
						UUID:        "123e4567-e89b-12d3-a456-426655441111",
						Name:        "testresourcebundle2",
						Description: "test description",
						ServiceType: "dns",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := bundleDefHandler{client: testCase.bundleDefClient}
			req, err := http.NewRequest("GET", "/v1/resource/definition", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.listHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}

			//Check returned body only if statusOK
			if rr.Code == http.StatusOK {
				got := []resource.BundleDefinition{}
				json.NewDecoder(rr.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestBundleDefGetHandler(t *testing.T) {

	testCases := []struct {
		label           string
		expected        resource.BundleDefinition
		inpUUID         string
		expectedCode    int
		bundleDefClient *mockBundleDefinition
	}{
		{
			label:        "Get Bundle Definition",
			expectedCode: http.StatusOK,
			expected: resource.BundleDefinition{
				UUID:        "123e4567-e89b-12d3-a456-426655441111",
				Name:        "testresourcebundle2",
				Description: "test description",
				ServiceType: "dns",
			},
			inpUUID: "123e4567-e89b-12d3-a456-426655441111",
			bundleDefClient: &mockBundleDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []resource.BundleDefinition{
					{
						UUID:        "123e4567-e89b-12d3-a456-426655441111",
						Name:        "testresourcebundle2",
						Description: "test description",
						ServiceType: "dns",
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting Bundle Definition",
			expectedCode: http.StatusInternalServerError,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			bundleDefClient: &mockBundleDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []resource.BundleDefinition{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := bundleDefHandler{client: testCase.bundleDefClient}
			req, err := http.NewRequest("GET", "/v1/resource/definition/"+testCase.inpUUID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.getHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}

			//Check returned body only if statusOK
			if rr.Code == http.StatusOK {
				got := resource.BundleDefinition{}
				json.NewDecoder(rr.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestBundleDefDeleteHandler(t *testing.T) {

	testCases := []struct {
		label           string
		inpUUID         string
		expectedCode    int
		bundleDefClient *mockBundleDefinition
	}{
		{
			label:           "Delete Bundle Definition",
			expectedCode:    http.StatusNoContent,
			inpUUID:         "123e4567-e89b-12d3-a456-426655441111",
			bundleDefClient: &mockBundleDefinition{},
		},
		{
			label:        "Delete Non-Exiting Bundle Definition",
			expectedCode: http.StatusInternalServerError,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			bundleDefClient: &mockBundleDefinition{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := bundleDefHandler{client: testCase.bundleDefClient}
			req, err := http.NewRequest("GET", "/v1/resource/definition/"+testCase.inpUUID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.deleteHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}
		})
	}
}
