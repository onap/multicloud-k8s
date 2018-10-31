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
	"k8splugin/vnfd"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockVNFDefinition struct {
	vnfd.VNFDefinitionInterface
	// Items and err will be used to customize each test
	// via a localized instantiation of mockVNFDefinition
	Items []vnfd.VNFDefinition
	Err   error
}

func (m *mockVNFDefinition) Create(inp vnfd.VNFDefinition) (vnfd.VNFDefinition, error) {
	if m.Err != nil {
		return vnfd.VNFDefinition{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockVNFDefinition) List() ([]vnfd.VNFDefinition, error) {
	if m.Err != nil {
		return []vnfd.VNFDefinition{}, m.Err
	}

	return m.Items, nil
}

func (m *mockVNFDefinition) Get(vnfID string) (vnfd.VNFDefinition, error) {
	if m.Err != nil {
		return vnfd.VNFDefinition{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockVNFDefinition) Delete(vnfID string) error {
	if m.Err != nil {
		return m.Err
	}

	return nil
}

func TestVnfdCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     vnfd.VNFDefinition
		expectedCode int
		vnfdClient   *mockVNFDefinition
	}{
		{
			label:        "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			vnfdClient:   &mockVNFDefinition{},
		},
		{
			label:        "Create without UUID",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"name":"testdomain",
				"description":"test description",
				"service-type":"firewall"
				}`)),
			expected: vnfd.VNFDefinition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testvnf",
				Description: "test description",
				ServiceType: "firewall",
			},
			vnfdClient: &mockVNFDefinition{
				//Items that will be returned by the mocked Client
				Items: []vnfd.VNFDefinition{
					{
						UUID:        "123e4567-e89b-12d3-a456-426655440000",
						Name:        "testvnf",
						Description: "test description",
						ServiceType: "firewall",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := vnfdHandler{vnfdClient: testCase.vnfdClient}
			req, err := http.NewRequest("POST", "/v1/vnfd", testCase.reader)

			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.vnfdCreateHandler)
			hr.ServeHTTP(rr, req)

			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}

			//Check returned body only if statusCreated
			if rr.Code == http.StatusCreated {
				got := vnfd.VNFDefinition{}
				json.NewDecoder(rr.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("vnfdCreateHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestVnfdListHandler(t *testing.T) {

	testCases := []struct {
		label        string
		expected     []vnfd.VNFDefinition
		expectedCode int
		vnfdClient   *mockVNFDefinition
	}{
		{
			label:        "List VNF Definitions",
			expectedCode: http.StatusOK,
			expected: []vnfd.VNFDefinition{
				{
					UUID:        "123e4567-e89b-12d3-a456-426655440000",
					Name:        "testvnf",
					Description: "test description",
					ServiceType: "firewall",
				},
				{
					UUID:        "123e4567-e89b-12d3-a456-426655441111",
					Name:        "testvnf2",
					Description: "test description",
					ServiceType: "dns",
				},
			},
			vnfdClient: &mockVNFDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []vnfd.VNFDefinition{
					{
						UUID:        "123e4567-e89b-12d3-a456-426655440000",
						Name:        "testvnf",
						Description: "test description",
						ServiceType: "firewall",
					},
					{
						UUID:        "123e4567-e89b-12d3-a456-426655441111",
						Name:        "testvnf2",
						Description: "test description",
						ServiceType: "dns",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := vnfdHandler{vnfdClient: testCase.vnfdClient}
			req, err := http.NewRequest("GET", "/v1/vnfd", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.vnfdListHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}

			//Check returned body only if statusOK
			if rr.Code == http.StatusOK {
				got := []vnfd.VNFDefinition{}
				json.NewDecoder(rr.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("vnfdListHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestVnfdGetHandler(t *testing.T) {

	testCases := []struct {
		label        string
		expected     vnfd.VNFDefinition
		inpUUID      string
		expectedCode int
		vnfdClient   *mockVNFDefinition
	}{
		{
			label:        "Get VNF Definition",
			expectedCode: http.StatusOK,
			expected: vnfd.VNFDefinition{
				UUID:        "123e4567-e89b-12d3-a456-426655441111",
				Name:        "testvnf2",
				Description: "test description",
				ServiceType: "dns",
			},
			inpUUID: "123e4567-e89b-12d3-a456-426655441111",
			vnfdClient: &mockVNFDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []vnfd.VNFDefinition{
					{
						UUID:        "123e4567-e89b-12d3-a456-426655441111",
						Name:        "testvnf2",
						Description: "test description",
						ServiceType: "dns",
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting VNF Definition",
			expectedCode: http.StatusInternalServerError,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			vnfdClient: &mockVNFDefinition{
				// list of definitions that will be returned by the mockclient
				Items: []vnfd.VNFDefinition{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := vnfdHandler{vnfdClient: testCase.vnfdClient}
			req, err := http.NewRequest("GET", "/v1/vnfd/"+testCase.inpUUID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.vnfdGetHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}

			//Check returned body only if statusOK
			if rr.Code == http.StatusOK {
				got := vnfd.VNFDefinition{}
				json.NewDecoder(rr.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("vnfdListHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestVnfdDeleteHandler(t *testing.T) {

	testCases := []struct {
		label        string
		inpUUID      string
		expectedCode int
		vnfdClient   *mockVNFDefinition
	}{
		{
			label:        "Delete VNF Definition",
			expectedCode: http.StatusNoContent,
			inpUUID:      "123e4567-e89b-12d3-a456-426655441111",
			vnfdClient:   &mockVNFDefinition{},
		},
		{
			label:        "Delete Non-Exiting VNF Definition",
			expectedCode: http.StatusInternalServerError,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			vnfdClient: &mockVNFDefinition{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := vnfdHandler{vnfdClient: testCase.vnfdClient}
			req, err := http.NewRequest("GET", "/v1/vnfd/"+testCase.inpUUID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.vnfdDeleteHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}
		})
	}
}
