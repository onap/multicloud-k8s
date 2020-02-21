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
)

func TestCreateGenericPlacementIntent(t *testing.T) {
	testCases := []struct {
		label                    string
		inputIntent              GenericPlacementIntent
		inputProject             string
		inputCompositeApp        string
		inputCompositeAppVersion string
		expectedError            string
		mockdb                   *db.MockDB
		expected                 GenericPlacementIntent
	}{
		{
			label: "Create GenericPlacementIntent",
			inputIntent: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: " A sample intent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: GenIntentSpecData{
					LogicalCloud: "logicalCloud1",
				},
			},
			inputProject:             "testProject",
			inputCompositeApp:        "testCompositeApp",
			inputCompositeAppVersion: "testCompositeAppVersion",
			expected: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testGenericPlacement",
					Description: " A sample intent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: GenIntentSpecData{
					LogicalCloud: "logicalCloud1",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProjectKey{ProjectName: "testProject"}.String(): {
						"projectmetadata": []byte(
							"{\"project-name\":\"testProject\"," +
								"\"description\":\"Test project for unit testing\"}"),
					},
					CompositeAppKey{CompositeAppName: "testCompositeApp",
						Version: "testCompositeAppVersion", Project: "testProject"}.String(): {
						"compositeapp": []byte(
							"{\"metadata\":{" +
								"\"name\":\"testCompositeApp\"," +
								"\"description\":\"description\"," +
								"\"userData1\":\"user data\"," +
								"\"userData2\":\"user data\"" +
								"}," +
								"\"spec\":{" +
								"\"version\":\"version of the composite app\"}}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			intentCli := NewGenericPlacementIntentClient()
			got, err := intentCli.CreateGenericPlacementIntent(testCase.inputIntent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateGenericPlacementIntent returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateGenericPlacementIntent returned unexpected body: got %v; "+" expected %v", got, testCase.expected)
				}
			}
		})

	}
}

func TestGetGenericPlacementIntent(t *testing.T) {

	testCases := []struct {
		label               string
		expectedError       string
		expected            GenericPlacementIntent
		mockdb              *db.MockDB
		intentName          string
		projectName         string
		compositeAppName    string
		compositeAppVersion string
	}{
		{
			label:               "Get Intent",
			intentName:          "testIntent",
			projectName:         "testProject",
			compositeAppName:    "testCompositeApp",
			compositeAppVersion: "testVersion",
			expected: GenericPlacementIntent{
				MetaData: GenIntentMetaData{
					Name:        "testIntent",
					Description: "A sample intent for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: GenIntentSpecData{
					LogicalCloud: "logicalCloud1",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					GenericPlacementIntentKey{
						Name:         "testIntent",
						Project:      "testProject",
						CompositeApp: "testCompositeApp",
						Version:      "testVersion",
					}.String(): {
						"genericplacementintent": []byte(
							"{\"metadata\":{\"Name\":\"testIntent\"," +
								"\"Description\":\"A sample intent for testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}," +
								"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			intentCli := NewGenericPlacementIntentClient()
			got, err := intentCli.GetGenericPlacementIntent(testCase.intentName, testCase.projectName, testCase.compositeAppName, testCase.compositeAppVersion)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetGenericPlacementIntent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetGenericPlacementIntent returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetGenericPlacementIntent returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}

		})
	}

}
