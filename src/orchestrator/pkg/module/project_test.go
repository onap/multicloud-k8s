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

package module

import (
	"reflect"
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

func TestCreateProject(t *testing.T) {
	testCases := []struct {
		label         string
		inp           Project
		expectedError string
		mockdb        *db.MockDB
		expected      Project
	}{
		{
			label: "Create Project",
			inp: Project{
				MetaData: ProjectMetaData{
					Name:        "testProject",
					Description: "A sample Project used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expected: Project{
				MetaData: ProjectMetaData{
					Name:        "testProject",
					Description: "A sample Project used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expectedError: "",
			mockdb:        &db.MockDB{},
		},
		{
			label:         "Failed Create Project",
			expectedError: "Error Creating Project",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Project"),
			},
		},
		{
			label: "Create Existing Project",
			inp: Project{
				MetaData: ProjectMetaData{
					Name:        "testProject",
					Description: "A sample Project used for unit testing",
					UserData1:   "data1",
					UserData2:   "data2",
				},
			},
			expectedError: "Project already exists",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProjectKey{ProjectName: "testProject"}.String(): {
						"projectmetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"testProject\"," +
								"\"Description\":\"Test project for unit testing\"," +
								"\"UserData1\":\"userData1\"," +
								"\"UserData2\":\"userData2\"}" +
								"}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewProjectClient()
			got, err := impl.CreateProject(testCase.inp, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Create returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestUpdateProject(t *testing.T) {
	testCases := []struct {
		label         string
		inp           Project
		expectedError string
		mockdb        *db.MockDB
		expected      Project
	}{
		{
			label: "Update Project",
			inp: Project{
				MetaData: ProjectMetaData{
					Name:        "testProject",
					Description: "Test project for unit testing",
					UserData1:   "update userData1",
					UserData2:   "update userData2",
				},
			},
			expected: Project{
				MetaData: ProjectMetaData{
					Name:        "testProject",
					Description: "Test project for unit testing",
					UserData1:   "update userData1",
					UserData2:   "update userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProjectKey{ProjectName: "testProject"}.String(): {
						"projectmetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"testProject\"," +
								"\"Description\":\"Test project for unit testing\"," +
								"\"UserData1\":\"userData1\"," +
								"\"UserData2\":\"userData2\"}" +
								"}"),
					},
				},
			},
		},
		{
			label: "Failed Update Project",
			inp: Project{
				MetaData: ProjectMetaData{
					Name:        "unknownProject",
					Description: "Unknown project for unit testing",
				},
			},
			expectedError: "Creating DB Entry",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Updating Project"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewProjectClient()
			got, err := impl.CreateProject(testCase.inp, true)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Update returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Update returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Update returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetProject(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      Project
	}{
		{
			label: "Get Project",
			name:  "testProject",
			expected: Project{
				MetaData: ProjectMetaData{
					Name:        "testProject",
					Description: "Test project for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProjectKey{ProjectName: "testProject"}.String(): {
						"projectmetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"testProject\"," +
								"\"Description\":\"Test project for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}" +
								"}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewProjectClient()
			got, err := impl.GetProject(testCase.name)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:  "Delete Project",
			name:   "testProject",
			mockdb: &db.MockDB{},
		},
		{
			label:         "Delete Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewProjectClient()
			err := impl.DeleteProject(testCase.name)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Delete returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Delete returned an unexpected error %s", err)
				}
			}
		})
	}
}
