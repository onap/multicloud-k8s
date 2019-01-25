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
	"k8splugin/internal/rb"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

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

func (m *mockRBDefinition) List() ([]rb.Definition, error) {
	if m.Err != nil {
		return []rb.Definition{}, m.Err
	}

	return m.Items, nil
}

func (m *mockRBDefinition) Get(id string) (rb.Definition, error) {
	if m.Err != nil {
		return rb.Definition{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockRBDefinition) Delete(id string) error {
	return m.Err
}

func (m *mockRBDefinition) Upload(id string, inp []byte) error {
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
			label:        "Create without UUID",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"name":"testdomain",
				"description":"test description",
				"service-type":"firewall"
				}`)),
			expected: rb.Definition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testresourcebundle",
				Description: "test description",
				ServiceType: "firewall",
			},
			rbDefClient: &mockRBDefinition{
				//Items that will be returned by the mocked Client
				Items: []rb.Definition{
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
			vh := rbDefinitionHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("POST", "/v1/rb/definition", testCase.reader)

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
				got := rb.Definition{}
				json.NewDecoder(rr.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestRBDefListHandler(t *testing.T) {

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
			rbDefClient: &mockRBDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []rb.Definition{
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
			vh := rbDefinitionHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("GET", "/v1/rb/definition", nil)
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
				got := []rb.Definition{}
				json.NewDecoder(rr.Body).Decode(&got)

				// Since the order of returned slice is not guaranteed
				// Check both and return error if both don't match
				sort.Slice(got, func(i, j int) bool {
					return got[i].UUID < got[j].UUID
				})
				// Sort both as it is not expected that testCase.expected
				// is sorted
				sort.Slice(testCase.expected, func(i, j int) bool {
					return testCase.expected[i].UUID < testCase.expected[j].UUID
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
		label        string
		expected     rb.Definition
		inpUUID      string
		expectedCode int
		rbDefClient  *mockRBDefinition
	}{
		{
			label:        "Get Bundle Definition",
			expectedCode: http.StatusOK,
			expected: rb.Definition{
				UUID:        "123e4567-e89b-12d3-a456-426655441111",
				Name:        "testresourcebundle2",
				Description: "test description",
				ServiceType: "dns",
			},
			inpUUID: "123e4567-e89b-12d3-a456-426655441111",
			rbDefClient: &mockRBDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []rb.Definition{
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
			rbDefClient: &mockRBDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []rb.Definition{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := rbDefinitionHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("GET", "/v1/rb/definition/"+testCase.inpUUID, nil)
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
				got := rb.Definition{}
				json.NewDecoder(rr.Body).Decode(&got)

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
		inpUUID      string
		expectedCode int
		rbDefClient  *mockRBDefinition
	}{
		{
			label:        "Delete Bundle Definition",
			expectedCode: http.StatusNoContent,
			inpUUID:      "123e4567-e89b-12d3-a456-426655441111",
			rbDefClient:  &mockRBDefinition{},
		},
		{
			label:        "Delete Non-Exiting Bundle Definition",
			expectedCode: http.StatusInternalServerError,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			rbDefClient: &mockRBDefinition{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := rbDefinitionHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("GET", "/v1/rb/definition/"+testCase.inpUUID, nil)
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

func TestRBDefUploadHandler(t *testing.T) {

	testCases := []struct {
		label        string
		inpUUID      string
		body         io.Reader
		expectedCode int
		rbDefClient  *mockRBDefinition
	}{
		{
			label:        "Upload Bundle Definition Content",
			expectedCode: http.StatusOK,
			inpUUID:      "123e4567-e89b-12d3-a456-426655441111",
			body: bytes.NewBuffer([]byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			}),
			rbDefClient: &mockRBDefinition{},
		},
		{
			label:        "Upload Invalid Bundle Definition Content",
			expectedCode: http.StatusInternalServerError,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
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
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			rbDefClient:  &mockRBDefinition{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := rbDefinitionHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("POST",
				"/v1/rb/definition/"+testCase.inpUUID+"/content", testCase.body)

			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.uploadHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}
		})
	}
}
