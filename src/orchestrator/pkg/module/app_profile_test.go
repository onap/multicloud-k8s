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
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// appProfileKeyStr and appProfileByAppKeyStr build the MockDB map keys the same
// way db.MockDB.Find does (fmt.Sprintf("%v", key)), since AppProfileKey and
// AppProfileFindByAppKey do not implement fmt.Stringer.
func appProfileKeyStr(profile string) string {
	return fmt.Sprintf("%v", AppProfileKey{
		Project:             apTestProject,
		CompositeApp:        apTestCompositeApp,
		CompositeAppVersion: apTestVersion,
		CompositeProfile:    apTestCompProfile,
		Profile:             profile,
	})
}

func appProfileByAppKeyStr(appName string) string {
	return fmt.Sprintf("%v", AppProfileFindByAppKey{
		Project:             apTestProject,
		CompositeApp:        apTestCompositeApp,
		CompositeAppVersion: apTestVersion,
		CompositeProfile:    apTestCompProfile,
		AppName:             appName,
	})
}

const (
	apTestProject      = "testProject"
	apTestCompositeApp = "testCompositeApp"
	apTestVersion      = "v1"
	apTestCompProfile  = "testCompositeProfile"
)

// appProfileParentItems seeds MockDB with the composite profile record that
// CreateAppProfile validates against.
func appProfileParentItems() map[string]map[string][]byte {
	return map[string]map[string][]byte{
		CompositeProfileKey{Name: apTestCompProfile, Project: apTestProject, CompositeApp: apTestCompositeApp, Version: apTestVersion}.String(): {
			"compositeprofilemetadata": []byte(
				"{\"metadata\":{\"name\":\"testCompositeProfile\"}}"),
		},
	}
}

func TestNewAppProfileClient(t *testing.T) {
	c := NewAppProfileClient()
	if c.storeName != "orchestrator" || c.tagMeta != "profilemetadata" || c.tagContent != "profilecontent" {
		t.Fatalf("NewAppProfileClient returned unexpected client: %+v", c)
	}
}

func TestCreateAppProfile(t *testing.T) {
	testCases := []struct {
		label         string
		inpProfile    AppProfile
		inpContent    AppProfileContent
		expectedError string
		mockdb        *db.MockDB
		expected      AppProfile
	}{
		{
			label: "Create App Profile",
			inpProfile: AppProfile{
				Metadata: AppProfileMetadata{Name: "testAppProfile"},
				Spec:     AppProfileSpec{AppName: "app1"},
			},
			inpContent: AppProfileContent{Profile: "profileContent"},
			expected: AppProfile{
				Metadata: AppProfileMetadata{Name: "testAppProfile"},
				Spec:     AppProfileSpec{AppName: "app1"},
			},
			mockdb: &db.MockDB{Items: appProfileParentItems()},
		},
		{
			label: "Create Existing App Profile",
			inpProfile: AppProfile{
				Metadata: AppProfileMetadata{Name: "existingAppProfile"},
				Spec:     AppProfileSpec{AppName: "app1"},
			},
			expectedError: "AppProfile already exists",
			mockdb: func() *db.MockDB {
				items := appProfileParentItems()
				items[appProfileKeyStr("existingAppProfile")] = map[string][]byte{
					"profilemetadata": []byte("{\"metadata\":{\"name\":\"existingAppProfile\"},\"spec\":{\"app-name\":\"app1\"}}"),
				}
				return &db.MockDB{Items: items}
			}(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewAppProfileClient()
			got, err := impl.CreateAppProfile(apTestProject, apTestCompositeApp, apTestVersion, apTestCompProfile, testCase.inpProfile, testCase.inpContent)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateAppProfile returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("CreateAppProfile returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("CreateAppProfile returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAppProfile(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      AppProfile
	}{
		{
			label: "Get App Profile",
			expected: AppProfile{
				Metadata: AppProfileMetadata{Name: "testAppProfile"},
				Spec:     AppProfileSpec{AppName: "app1"},
			},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					appProfileKeyStr("testAppProfile"): {
						"profilemetadata": []byte("{\"metadata\":{\"name\":\"testAppProfile\"},\"spec\":{\"app-name\":\"app1\"}}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb:        &db.MockDB{Err: pkgerrors.New("DB Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewAppProfileClient()
			got, err := impl.GetAppProfile(apTestProject, apTestCompositeApp, apTestVersion, apTestCompProfile, "testAppProfile")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAppProfile returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAppProfile returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAppProfile returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAppProfiles(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      []AppProfile
	}{
		{
			label: "Get All App Profiles",
			expected: []AppProfile{
				{
					Metadata: AppProfileMetadata{Name: "testAppProfile"},
					Spec:     AppProfileSpec{AppName: "app1"},
				},
			},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					appProfileKeyStr(""): {
						"profilemetadata": []byte("{\"metadata\":{\"name\":\"testAppProfile\"},\"spec\":{\"app-name\":\"app1\"}}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb:        &db.MockDB{Err: pkgerrors.New("DB Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewAppProfileClient()
			got, err := impl.GetAppProfiles(apTestProject, apTestCompositeApp, apTestVersion, apTestCompProfile)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAppProfiles returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAppProfiles returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAppProfiles returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAppProfileByApp(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      AppProfile
	}{
		{
			label: "Get App Profile By App",
			expected: AppProfile{
				Metadata: AppProfileMetadata{Name: "testAppProfile"},
				Spec:     AppProfileSpec{AppName: "app1"},
			},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					appProfileByAppKeyStr("app1"): {
						"profilemetadata": []byte("{\"metadata\":{\"name\":\"testAppProfile\"},\"spec\":{\"app-name\":\"app1\"}}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb:        &db.MockDB{Err: pkgerrors.New("DB Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewAppProfileClient()
			got, err := impl.GetAppProfileByApp(apTestProject, apTestCompositeApp, apTestVersion, apTestCompProfile, "app1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAppProfileByApp returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAppProfileByApp returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAppProfileByApp returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

// appProfileFindByAppKeyString is a compile-time helper to build the string key
// for AppProfileFindByAppKey since it has no String() method of its own; the
// db.MockDB uses fmt.Sprintf("%v", key) which relies on the struct's default
// formatting, matched below via the same fmt path.

func TestGetAppProfileContent(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      AppProfileContent
	}{
		{
			label:    "Get App Profile Content",
			expected: AppProfileContent{Profile: "profileContent"},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					appProfileKeyStr("testAppProfile"): {
						"profilecontent": []byte("{\"profile\":\"profileContent\"}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb:        &db.MockDB{Err: pkgerrors.New("DB Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewAppProfileClient()
			got, err := impl.GetAppProfileContent(apTestProject, apTestCompositeApp, apTestVersion, apTestCompProfile, "testAppProfile")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAppProfileContent returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAppProfileContent returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAppProfileContent returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAppProfileContentByApp(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      AppProfileContent
	}{
		{
			label:    "Get App Profile Content By App",
			expected: AppProfileContent{Profile: "profileContent"},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					appProfileByAppKeyStr("app1"): {
						"profilecontent": []byte("{\"profile\":\"profileContent\"}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb:        &db.MockDB{Err: pkgerrors.New("DB Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewAppProfileClient()
			got, err := impl.GetAppProfileContentByApp(apTestProject, apTestCompositeApp, apTestVersion, apTestCompProfile, "app1")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAppProfileContentByApp returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAppProfileContentByApp returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAppProfileContentByApp returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteAppProfile(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:  "Delete App Profile",
			mockdb: &db.MockDB{},
		},
		{
			label:         "Delete Error",
			expectedError: "DB Error",
			mockdb:        &db.MockDB{Err: pkgerrors.New("DB Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewAppProfileClient()
			err := impl.DeleteAppProfile(apTestProject, apTestCompositeApp, apTestVersion, apTestCompProfile, "testAppProfile")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("DeleteAppProfile returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("DeleteAppProfile returned an unexpected error %s", err)
				}
			}
		})
	}
}
