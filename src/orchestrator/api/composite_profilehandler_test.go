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
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockCompositeProfileManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockCompositeProfileManager
	Items []moduleLib.CompositeProfile
	Err   error
}

func (m *mockCompositeProfileManager) CreateCompositeProfile(inp moduleLib.CompositeProfile, p string, ca string,
	v string) (moduleLib.CompositeProfile, error) {
	if m.Err != nil {
		return moduleLib.CompositeProfile{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockCompositeProfileManager) GetCompositeProfile(name string, projectName string,
	compositeAppName string, version string) (moduleLib.CompositeProfile, error) {
	if m.Err != nil {
		return moduleLib.CompositeProfile{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockCompositeProfileManager) GetCompositeProfiles(projectName string,
	compositeAppName string, version string) ([]moduleLib.CompositeProfile, error) {
	if m.Err != nil {
		return []moduleLib.CompositeProfile{}, m.Err
	}

	return m.Items, nil
}

func (m *mockCompositeProfileManager) DeleteCompositeProfile(name string, projectName string,
	compositeAppName string, version string) error {
	return m.Err
}

func init() {
	caprofileJSONFile = "../json-schemas/metadata.json"
}

func Test_compositeProfileHandler_createHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     moduleLib.CompositeProfile
		expectedCode int
		cProfClient  *mockCompositeProfileManager
	}{
		{
			label:        "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			cProfClient:  &mockCompositeProfileManager{},
		},
		{
			label:        "Create Composite Profile",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"metadata" : {
					"name": "testCompositeProfile",
    				"description": "Test CompositeProfile used for unit testing",
    				"userData1": "data1",
    				"userData2": "data2"
				}
			}`)),
			expected: moduleLib.CompositeProfile{
				Metadata: moduleLib.CompositeProfileMetadata{
					Name:        "testCompositeProfile",
					Description: "Test CompositeProfile used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			cProfClient: &mockCompositeProfileManager{
				//Items that will be returned by the mocked Client
				Items: []moduleLib.CompositeProfile{
					{
						Metadata: moduleLib.CompositeProfileMetadata{
							Name:        "testCompositeProfile",
							Description: "Test CompositeProfile used for unit testing",
							UserData1:   "data1",
							UserData2:   "data2",
						},
					},
				},
			},
		},
		{
			label: "Missing Composite Profile Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			expectedCode: http.StatusBadRequest,
			cProfClient:  &mockCompositeProfileManager{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/projects/{project-name}/composite-apps/{composite-app-name}/{version}/composite-profiles", testCase.reader)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, testCase.cProfClient, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := moduleLib.CompositeProfile{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}

}
