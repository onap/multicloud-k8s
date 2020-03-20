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
	//  pkgerrors "github.com/pkg/errors"
)

func TestCreateApp(t *testing.T) {
	testCases := []struct {
		label                  string
		inpApp                 App
		inpAppContent          AppContent
		inpProject             string
		inpCompositeAppName    string
		inpCompositeAppVersion string
		expectedError          string
		mockdb                 *db.MockDB
		expected               App
	}{
		{
			label: "Create App",
			inpApp: App{
				Metadata: AppMetaData{
					Name:        "testApp",
					Description: "A sample app used for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},

			inpAppContent: AppContent{
				FileContent: "Sample file content",
			},
			inpProject:             "testProject",
			inpCompositeAppName:    "testCompositeApp",
			inpCompositeAppVersion: "v1",
			expected: App{
				Metadata: AppMetaData{
					Name:        "testApp",
					Description: "A sample app used for unit testing",
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
								"\"metadata\": {" +
								"\"Name\": \"testProject\"," +
								"\"Description\": \"Test project for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}" +
								"}"),
					},
					CompositeAppKey{CompositeAppName: "testCompositeApp", Version: "v1", Project: "testProject"}.String(): {
						"compositeappmetadata": []byte(
							"{" +
								"\"metadata\":{" +
								"\"Name\":\"testCompositeApp\"," +
								"\"Description\":\"Test CompositeApp for unit testing\"," +
								"\"UserData1\":\"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\":{" +
								"\"Version\":\"v1\"}" +
								"}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewAppClient()
			got, err := impl.CreateApp(testCase.inpApp, testCase.inpAppContent, testCase.inpProject, testCase.inpCompositeAppName, testCase.inpCompositeAppVersion)
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

func TestGetApp(t *testing.T) {

	testCases := []struct {
		label                  string
		inpApp                 string
		inpProject             string
		inpCompositeAppName    string
		inpCompositeAppVersion string
		expectedError          string
		mockdb                 *db.MockDB
		expected               App
	}{
		{
			label:                  "Get Composite App",
			inpApp:                 "testApp",
			inpProject:             "testProject",
			inpCompositeAppName:    "testCompositeApp",
			inpCompositeAppVersion: "v1",
			expected: App{
				Metadata: AppMetaData{
					Name:        "testApp",
					Description: "Test App for unit testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					AppKey{App: "testApp", Project: "testProject", CompositeApp: "testCompositeApp", CompositeAppVersion: "v1"}.String(): {
						"appmetadata": []byte(
							"{" +
								"\"metadata\": {" +
								"\"Name\": \"testApp\"," +
								"\"Description\": \"Test App for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}" +
								"}"),
						"appcontent": []byte(
							"{" +
								"\"FileContent\": \"sample file content\"" +
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
			impl := NewAppClient()
			got, err := impl.GetApp(testCase.inpApp, testCase.inpProject, testCase.inpCompositeAppName, testCase.inpCompositeAppVersion)
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

func TestGetAppContent(t *testing.T) {

	testCases := []struct {
		label                  string
		inpApp                 string
		inpProject             string
		inpCompositeAppName    string
		inpCompositeAppVersion string
		expectedError          string
		mockdb                 *db.MockDB
		expected               AppContent
	}{
		{
			label:                  "Get App content",
			inpApp:                 "testApp",
			inpProject:             "testProject",
			inpCompositeAppName:    "testCompositeApp",
			inpCompositeAppVersion: "v1",
			expected: AppContent{
				FileContent: "Samplefilecontent",
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					AppKey{App: "testApp", Project: "testProject", CompositeApp: "testCompositeApp", CompositeAppVersion: "v1"}.String(): {
						"appmetadata": []byte(
							"{" +
								"\"metadata\": {" +
								"\"Name\": \"testApp\"," +
								"\"Description\": \"Test App for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}" +
								"}"),
						"appcontent": []byte(
							"{" +
								"\"FileContent\": \"Samplefilecontent\"" +
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
			impl := NewAppClient()
			got, err := impl.GetAppContent(testCase.inpApp, testCase.inpProject, testCase.inpCompositeAppName, testCase.inpCompositeAppVersion)
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

func TestDeleteApp(t *testing.T) {

	testCases := []struct {
		label                  string
		inpApp                 string
		inpProject             string
		inpCompositeAppName    string
		inpCompositeAppVersion string
		expectedError          string
		mockdb                 *db.MockDB
	}{
		{
			label:                  "Delete App",
			inpApp:                 "testApp",
			inpProject:             "testProject",
			inpCompositeAppName:    "testCompositeApp",
			inpCompositeAppVersion: "v1",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					AppKey{App: "testApp", Project: "testProject", CompositeApp: "testCompositeApp", CompositeAppVersion: "v1"}.String(): {
						"appmetadata": []byte(
							"{" +
								"\"metadata\": {" +
								"\"Name\": \"testApp\"," +
								"\"Description\": \"Test App for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}" +
								"}"),
						"appcontent": []byte(
							"{" +
								"\"FileContent\": \"Samplefilecontent\"" +
								"}"),
					},
				},
			},
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
			impl := NewAppClient()
			err := impl.DeleteApp(testCase.inpApp, testCase.inpProject, testCase.inpCompositeAppName, testCase.inpCompositeAppVersion)
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
