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

// intentTestParentItems returns MockDB Items seeded with the parent
// project/compositeApp/deploymentIntentGroup records that AddIntent validates
// against before inserting.
func intentTestParentItems() map[string]map[string][]byte {
	return map[string]map[string][]byte{
		ProjectKey{ProjectName: "testProject"}.String(): {
			"projectmetadata": []byte(
				"{\"metadata\":{\"Name\":\"testProject\"}}"),
		},
		CompositeAppKey{CompositeAppName: "testCompositeApp", Version: "v1", Project: "testProject"}.String(): {
			"compositeappmetadata": []byte(
				"{\"metadata\":{\"Name\":\"testCompositeApp\"},\"spec\":{\"Version\":\"v1\"}}"),
		},
		DeploymentIntentGroupKey{Name: "testDig", Project: "testProject", CompositeApp: "testCompositeApp", Version: "v1"}.String(): {
			"deploymentintentgroupmetadata": []byte(
				"{\"metadata\":{\"name\":\"testDig\"},\"spec\":{\"profile\":\"prof\",\"version\":\"v1\",\"logical-cloud\":\"lc\"}}"),
		},
	}
}

func TestIntentKeyString(t *testing.T) {
	k := IntentKey{
		Name:                  "i1",
		Project:               "p1",
		CompositeApp:          "ca1",
		Version:               "v1",
		DeploymentIntentGroup: "dig1",
	}
	got := k.String()
	if !strings.Contains(got, "\"intentname\":\"i1\"") ||
		!strings.Contains(got, "\"deploymentintentgroup\":\"dig1\"") {
		t.Fatalf("IntentKey.String returned unexpected value: %s", got)
	}
}

func TestNewIntentClient(t *testing.T) {
	c := NewIntentClient()
	if c.storeName != "orchestrator" || c.tagMetaData != "addintent" {
		t.Fatalf("NewIntentClient returned unexpected client: %+v", c)
	}
}

func TestAddIntent(t *testing.T) {
	testCases := []struct {
		label         string
		inp           Intent
		expectedError string
		mockdb        *db.MockDB
		expected      Intent
	}{
		{
			label: "Add Intent",
			inp: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample intent",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{"gpi": "gpiName"},
				},
			},
			expected: Intent{
				MetaData: IntentMetaData{
					Name:        "testIntent",
					Description: "A sample intent",
				},
				Spec: IntentSpecData{
					Intent: map[string]string{"gpi": "gpiName"},
				},
			},
			mockdb: &db.MockDB{Items: intentTestParentItems()},
		},
		{
			label: "Add Existing Intent",
			inp: Intent{
				MetaData: IntentMetaData{Name: "existingIntent"},
			},
			expectedError: "Intent already exists",
			mockdb: func() *db.MockDB {
				items := intentTestParentItems()
				items[IntentKey{Name: "existingIntent", Project: "testProject", CompositeApp: "testCompositeApp", Version: "v1", DeploymentIntentGroup: "testDig"}.String()] = map[string][]byte{
					"addintent": []byte("{\"metadata\":{\"name\":\"existingIntent\"}}"),
				}
				return &db.MockDB{Items: items}
			}(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewIntentClient()
			got, err := impl.AddIntent(testCase.inp, "testProject", "testCompositeApp", "v1", "testDig")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("AddIntent returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("AddIntent returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("AddIntent returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetIntent(t *testing.T) {
	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		expected      Intent
	}{
		{
			label: "Get Intent",
			name:  "testIntent",
			expected: Intent{
				MetaData: IntentMetaData{Name: "testIntent", Description: "sample"},
			},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					IntentKey{Name: "testIntent", Project: "testProject", CompositeApp: "testCompositeApp", Version: "v1", DeploymentIntentGroup: "testDig"}.String(): {
						"addintent": []byte("{\"metadata\":{\"name\":\"testIntent\",\"description\":\"sample\"}}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			name:          "testIntent",
			expectedError: "DB Error",
			mockdb:        &db.MockDB{Err: pkgerrors.New("DB Error")},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewIntentClient()
			got, err := impl.GetIntent(testCase.name, "testProject", "testCompositeApp", "v1", "testDig")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetIntent returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetIntent returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetIntent returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetIntentByName(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      IntentSpecData
	}{
		{
			label: "Get Intent By Name",
			expected: IntentSpecData{
				Intent: map[string]string{"gpi": "gpiName"},
			},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					IntentKey{Name: "testIntent", Project: "testProject", CompositeApp: "testCompositeApp", Version: "v1", DeploymentIntentGroup: "testDig"}.String(): {
						"addintent": []byte("{\"metadata\":{\"name\":\"testIntent\"},\"spec\":{\"intent\":{\"gpi\":\"gpiName\"}}}"),
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
			impl := NewIntentClient()
			got, err := impl.GetIntentByName("testIntent", "testProject", "testCompositeApp", "v1", "testDig")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetIntentByName returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetIntentByName returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetIntentByName returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetAllIntents(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      ListOfIntents
	}{
		{
			label: "Get All Intents",
			expected: ListOfIntents{
				ListOfIntents: []map[string]string{{"gpi": "gpiName"}},
			},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					IntentKey{Name: "", Project: "testProject", CompositeApp: "testCompositeApp", Version: "v1", DeploymentIntentGroup: "testDig"}.String(): {
						"addintent": []byte("{\"metadata\":{\"name\":\"testIntent\"},\"spec\":{\"intent\":{\"gpi\":\"gpiName\"}}}"),
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
			impl := NewIntentClient()
			got, err := impl.GetAllIntents("testProject", "testCompositeApp", "v1", "testDig")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("GetAllIntents returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("GetAllIntents returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("GetAllIntents returned unexpected body: got %v; expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteIntent(t *testing.T) {
	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:  "Delete Intent",
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
			impl := NewIntentClient()
			err := impl.DeleteIntent("testIntent", "testProject", "testCompositeApp", "v1", "testDig")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("DeleteIntent returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("DeleteIntent returned an unexpected error %s", err)
				}
			}
		})
	}
}
