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

//pkgerrors "github.com/pkg/errors"

/* TODO - db.MockDB needs to be enhanced and then these can be fixed up
func TestCreateCompositeProfile(t *testing.T) {
	testCases := []struct {
		label               string
		compositeProfile    CompositeProfile
		projectName         string
		compositeApp        string
		compositeAppVersion string
		expectedError       string
		mockdb              *db.MockDB
		expected            CompositeProfile
	}{
		{
			label: "Create CompositeProfile",
			compositeProfile: CompositeProfile{
				Metadata: CompositeProfileMetadata{
					Name:        "testCompositeProfile",
					Description: "A sample Composite Profile for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			projectName:         "testProject",
			compositeApp:        "testCompositeApp",
			compositeAppVersion: "v1",
			expected: CompositeProfile{
				Metadata: CompositeProfileMetadata{
					Name:        "testCompositeProfile",
					Description: "A sample Composite Profile for testing",
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
					CompositeAppKey{CompositeAppName: "testCompositeApp", Project: "testProject", Version: "v1"}.String(): {
						"compositeAppmetadata": []byte(
							"{" +
								"\"metadata\" : {" +
								"\"Name\":\"testCompositeApp\"," +
								"\"Description\":\"Test Composite App for unit testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\":\"userData2\"}," +
								"\"spec\": {" +
								"\"Version\": \"v1\"}" +
								"}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			cprofCli := NewCompositeProfileClient()
			got, err := cprofCli.CreateCompositeProfile(testCase.compositeProfile, testCase.projectName, testCase.compositeApp, testCase.compositeAppVersion)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateCompositeProfile returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateCompositeProfile returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateCompositeProfile returned unexpected body: got %v; "+" expected %v", got, testCase.expected)
				}
			}
		})

	}
}

func TestGetCompositeProfile(t *testing.T) {

	testCases := []struct {
		label                string
		expectedError        string
		expected             CompositeProfile
		mockdb               *db.MockDB
		compositeProfileName string
		projectName          string
		compositeAppName     string
		compositeAppVersion  string
	}{
		{
			label:                "Get CompositeProfile",
			compositeProfileName: "testCompositeProfile",
			projectName:          "testProject",
			compositeAppName:     "testCompositeApp",
			compositeAppVersion:  "v1",
			expected: CompositeProfile{
				Metadata: CompositeProfileMetadata{
					Name:        "testCompositeProfile",
					Description: "A sample CompositeProfile for testing",
					UserData1:   "userData1",
					UserData2:   "userData2",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					CompositeProfileKey{
						Name:         "testCompositeProfile",
						Project:      "testProject",
						CompositeApp: "testCompositeApp",
						Version:      "v1",
					}.String(): {
						"compositeprofile": []byte(
							"{\"metadata\":{\"Name\":\"testCompositeProfile\"," +
								"\"Description\":\"A sample CompositeProfile for testing\"," +
								"\"UserData1\": \"userData1\"," +
								"\"UserData2\": \"userData2\"}}"),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			cprofCli := NewCompositeProfileClient()
			got, err := cprofCli.GetCompositeProfile(testCase.compositeProfileName, testCase.projectName, testCase.compositeAppName, testCase.compositeAppVersion)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetCompositeProfile returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetCompositeProfile returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetCompositeProfile returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}

		})
	}

}
*/
