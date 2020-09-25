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

	gpic "github.com/onap/multicloud-k8s/src/orchestrator/pkg/gpic"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
)

func TestCreateAppIntent(t *testing.T) {
	testCases := []struct {
		label                       string
		inputAppIntent              AppIntent
		inputProject                string
		inputCompositeApp           string
		inputCompositeAppVersion    string
		inputGenericPlacementIntent string
		inputDeploymentIntentGrpName string
		expectedError               string
		mockdb                      *db.MockDB
		expected                    AppIntent
	}{
		{
			label: "Create AppIntent",
			inputAppIntent: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
								//ClusterLabelName: "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
								//ClusterLabelName: "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{ProviderName: "aws",
										ClusterLabelName: "east-us1"},
									{ProviderName: "aws",
										ClusterLabelName: "east-us2"},
									//{ClusterName: "east-us1"},
									//{ClusterName: "east-us2"},
								},
							},
						},

						AnyOfArray: []gpic.AnyOf{},
					},
				},
			},

			inputProject:                "testProject",
			inputCompositeApp:           "testCompositeApp",
			inputCompositeAppVersion:    "testCompositeAppVersion",
			inputGenericPlacementIntent: "testIntent",
			inputDeploymentIntentGrpName: "testDeploymentIntentGroup",
			expected: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "A sample AppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
								//ClusterLabelName: "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
								//ClusterLabelName: "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{ProviderName: "aws",
										ClusterLabelName: "east-us1"},
									{ProviderName: "aws",
										ClusterLabelName: "east-us2"},
									//{ClusterName: "east-us1"},
									//{ClusterName: "east-us2"},
								},
							},
						},
						AnyOfArray: []gpic.AnyOf{},
					},
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
						"compositeappmetadata": []byte(
							"{\"metadata\":{" +
								"\"name\":\"testCompositeApp\"," +
								"\"description\":\"description\"," +
								"\"userData1\":\"user data\"," +
								"\"userData2\":\"user data\"" +
								"}," +
								"\"spec\":{" +
								"\"version\":\"version of the composite app\"}}"),
					},
					GenericPlacementIntentKey{
						Name:         "testIntent",
						Project:      "testProject",
						CompositeApp: "testCompositeApp",
						Version:      "testCompositeAppVersion",
						DigName: "testDeploymentIntentGroup",
					}.String(): {
						"genericplacementintentmetadata": []byte(
							"{\"metadata\":{\"Name\":\"testIntent\"," +
								"\"Description\":\"A sample intent for testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}," +
								"\"spec\":{\"Logical-Cloud\": \"logicalCloud1\"}}"),
					},
					DeploymentIntentGroupKey{
						Name:         "testDeploymentIntentGroup",
						Project:      "testProject",
						CompositeApp: "testCompositeApp",
						Version:      "testCompositeAppVersion",
					}.String(): {
						"deploymentintentgroupmetadata": []byte(
							"{\"metadata\":{\"name\":\"testDeploymentIntentGroup\"," +
								"\"description\":\"DescriptionTestDeploymentIntentGroup\"," +
								"\"userData1\": \"userData1\"," +
								"\"userData2\": \"userData2\"}," +
								"\"spec\":{\"profile\": \"Testprofile\"," +
								"\"version\": \"version of deployment\"," +
								"\"override-values\":[" +
								"{" +
								"\"app-name\": \"TestAppName\"," +
								"\"values\": " +
								"{" +
								"\"imageRepository\":\"registry.hub.docker.com\"" +
								"}" +
								"}," +
								"{" +
								"\"app-name\": \"TestAppName\"," +
								"\"values\": " +
								"{" +
								"\"imageRepository\":\"registry.hub.docker.com\"" +
								"}" +
								"}" +
								"]," +
								"\"logical-cloud\": \"cloud1\"" +
								"}"+
								"}"),
					},
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			appIntentCli := NewAppIntentClient()
			got, err := appIntentCli.CreateAppIntent(testCase.inputAppIntent, testCase.inputProject, testCase.inputCompositeApp, testCase.inputCompositeAppVersion, testCase.inputGenericPlacementIntent, testCase.inputDeploymentIntentGrpName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateAppIntent returned an unexpected error %s, ", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateAppIntent returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateAppIntent returned unexpected body: got %v; "+" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAppIntent(t *testing.T) {
	testCases := []struct {
		label                  string
		expectedError          string
		expected               AppIntent
		mockdb                 *db.MockDB
		appIntentName          string
		projectName            string
		compositeAppName       string
		compositeAppVersion    string
		genericPlacementIntent string
		deploymentIntentgrpName string
	}{
		{
			label:                  "Get Intent",
			appIntentName:          "testAppIntent",
			projectName:            "testProject",
			compositeAppName:       "testCompositeApp",
			compositeAppVersion:    "testCompositeAppVersion",
			genericPlacementIntent: "testIntent",
			deploymentIntentgrpName: "testDeploymentIntentGroup",
			expected: AppIntent{
				MetaData: MetaData{
					Name:        "testAppIntent",
					Description: "testAppIntent",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
				Spec: SpecData{
					AppName: "SampleApp",
					Intent: gpic.IntentStruc{
						AllOfArray: []gpic.AllOf{
							{
								ProviderName: "aws",
								ClusterName:  "edge1",
							},
							{
								ProviderName: "aws",
								ClusterName:  "edge2",
							},
							{
								AnyOfArray: []gpic.AnyOf{
									{ProviderName: "aws",
										ClusterLabelName: "east-us1"},
									{ProviderName: "aws",
										ClusterLabelName: "east-us2"},
								},
							},
						},
					},
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					AppIntentKey{
						Name:         "testAppIntent",
						Project:      "testProject",
						CompositeApp: "testCompositeApp",
						Version:      "testCompositeAppVersion",
						Intent:       "testIntent",
						DeploymentIntentGroupName: "testDeploymentIntentGroup",
					}.String(): {
						"appintentmetadata": []byte(
							"{\"metadata\":{\"Name\":\"testAppIntent\"," +
								"\"Description\":\"testAppIntent\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}," +
								"\"spec\":{\"app-name\": \"SampleApp\"," +
								"\"intent\": {" +
								"\"allOf\":[" +
								"{" +
								"\"provider-name\":\"aws\"," +
								"\"cluster-name\":\"edge1\"}," +
								"{" +
								"\"provider-name\":\"aws\"," +
								"\"cluster-name\":\"edge2\"}," +
								"{" +
								"\"anyOf\":[" +
								"{" +
								"\"provider-name\":\"aws\"," +
								"\"cluster-label-name\":\"east-us1\"}," +
								"{" +
								"\"provider-name\":\"aws\"," +
								"\"cluster-label-name\":\"east-us2\"}" +
								"]}]" +
								"}}}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			appIntentCli := NewAppIntentClient()
			got, err := appIntentCli.GetAppIntent(testCase.appIntentName, testCase.projectName, testCase.compositeAppName, testCase.compositeAppVersion,
				testCase.genericPlacementIntent, testCase.deploymentIntentgrpName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAppIntent returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAppIntent returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAppIntent returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}

		})
	}
}
