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
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/internal/project"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockProjectManager struct {
	// Items and err will be used to customize each test
	// via a localized instantiation of mockProjectManager
	Items []project.Project
	Err   error
}

func (m *mockProjectManager) Create(inp project.Project) (project.Project, error) {
	if m.Err != nil {
		return project.Project{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockProjectManager) Get(name string) (project.Project, error) {
	if m.Err != nil {
		return project.Project{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockProjectManager) Delete(name string) error {
	return m.Err
}

func TestProjectCreateHandler(t *testing.T) {
	testCases := []struct {
		label         string
		reader        io.Reader
		expected      project.Project
		expectedCode  int
		projectClient *mockProjectManager
	}{
		{
			label:         "Missing Body Failure",
			expectedCode:  http.StatusBadRequest,
			projectClient: &mockProjectManager{},
		},
		{
			label:        "Create Project",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"project-name":"testProject",
				"description":"Test Project used for unit testing"
				}`)),
			expected: project.Project{
				ProjectName: "testProject",
				Description: "Test Project used for unit testing",
			},
			projectClient: &mockProjectManager{
				//Items that will be returned by the mocked Client
				Items: []project.Project{
					{
						ProjectName: "testProject",
						Description: "Test Project used for unit testing",
					},
				},
			},
		},
		{
			label: "Missing Project Name in Request Body",
			reader: bytes.NewBuffer([]byte(`{
				"description":"test description"
				}`)),
			expectedCode:  http.StatusBadRequest,
			projectClient: &mockProjectManager{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v2/project", testCase.reader)
			resp := executeRequest(request, NewRouter(testCase.projectClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := project.Project{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestProjectGetHandler(t *testing.T) {

	testCases := []struct {
		label         string
		expected      project.Project
		name, version string
		expectedCode  int
		projectClient *mockProjectManager
	}{
		{
			label:        "Get Project",
			expectedCode: http.StatusOK,
			expected: project.Project{
				ProjectName: "testProject",
				Description: "A Test project for unit testing",
			},
			name: "testProject",
			projectClient: &mockProjectManager{
				Items: []project.Project{
					{
						ProjectName: "testProject",
						Description: "A Test project for unit testing",
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting Project",
			expectedCode: http.StatusInternalServerError,
			name:         "nonexistingproject",
			projectClient: &mockProjectManager{
				Items: []project.Project{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v2/project/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.projectClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := project.Project{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestProjectDeleteHandler(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		version       string
		expectedCode  int
		projectClient *mockProjectManager
	}{
		{
			label:         "Delete Project",
			expectedCode:  http.StatusNoContent,
			name:          "testProject",
			projectClient: &mockProjectManager{},
		},
		{
			label:        "Delete Non-Exiting Project",
			expectedCode: http.StatusInternalServerError,
			name:         "testProject",
			projectClient: &mockProjectManager{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v2/project/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(testCase.projectClient))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
