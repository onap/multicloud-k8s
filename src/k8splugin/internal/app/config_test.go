/*
 * Copyright 2018 Intel Corporation, Inc
 * Copyright © 2021 Samsung Electronics
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

package app

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	//	pkgerrors "github.com/pkg/errors"
)

func provideMockModelData(instanceID, rbName, rbVersion, profileName string) *db.MockDB {
	return &db.MockDB{
		Items: map[string]map[string][]byte{
			InstanceKey{ID: instanceID}.String(): {
				"instance": []byte(fmt.Sprintf(
					`{
          "id": "%s",
          "request": {
            "rb-name": "%s",
            "rb-version": "%s",
            "profile-name": "%s"
          }
        }`, instanceID, rbName, rbVersion, profileName)),
			},
		},
	}
}

func TestCreateConfig(t *testing.T) {
	testCases := []struct {
		label         string
		rbName        string
		rbVersion     string
		profileName   string
		instanceID    string
		inp           Config
		expectedError string
		mockdb        *db.MockEtcdClient
		expected      ConfigResult
	}{
		{
			label:       "Create Config",
			rbName:      "testdef1",
			rbVersion:   "v1",
			profileName: "testprofile1",
			instanceID:  "testinstance1",
			inp: Config{
				ConfigName:   "testconfig1",
				TemplateName: "testtemplate1",
				Values: map[string]interface{}{
					"values": "{\"namespace\": \"kafka\", \"topic\": {\"name\":\"orders\", \"cluster\":\"my-cluster\", \"partitions\": 10,\"replicas\":   2, }}"},
			},
			expected: ConfigResult{
				InstanceName:      "testinstance1",
				DefinitionName:    "testdef1",
				DefinitionVersion: "v1",
				ProfileName:       "testprofile1",
				ConfigName:        "testconfig1",
				TemplateName:      "testtemplate1",
				ConfigVersion:     1,
			},
			expectedError: "",
			mockdb: &db.MockEtcdClient{
				Items: nil,
				Err:   nil,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.Etcd = testCase.mockdb
			db.DBconn = provideMockModelData(testCase.instanceID, testCase.rbName,
				testCase.rbVersion, testCase.profileName)
			resolve = func(rbName, rbVersion, profileName string, p Config, releaseName string) (configResourceList, error) {
				return configResourceList{}, nil
			}
			impl := NewConfigClient()
			got, err := impl.Create(testCase.instanceID, testCase.inp)
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

func TestRollbackConfig(t *testing.T) {
	testCases := []struct {
		label          string
		rbName         string
		rbVersion      string
		profileName    string
		instanceID     string
		inp            Config
		inpUpdate1     Config
		inpUpdate2     Config
		expectedError  string
		mockdb         *db.MockEtcdClient
		expected1      ConfigResult
		expected2      ConfigResult
		expected3      ConfigResult
		expected4      ConfigResult
		rollbackConfig ConfigRollback
	}{
		{
			label:       "Rollback Config",
			rbName:      "testdef1",
			rbVersion:   "v1",
			profileName: "testprofile1",
			instanceID:  "testinstance1",
			inp: Config{
				ConfigName:   "testconfig1",
				TemplateName: "testtemplate1",
				Values: map[string]interface{}{
					"values": "{\"namespace\": \"kafka\", \"topic\": {\"name\":\"orders\", \"cluster\":\"my-cluster\", \"partitions\": 10,\"replicas\":   2, }}"},
			},
			inpUpdate1: Config{
				ConfigName:   "testconfig1",
				TemplateName: "testtemplate1",
				Values: map[string]interface{}{
					"values": "{\"namespace\": \"kafka\", \"topic\": {\"name\":\"orders\", \"cluster\":\"my-cluster\", \"partitions\": 20,\"replicas\":   2, }}"},
			},
			inpUpdate2: Config{
				ConfigName:   "testconfig1",
				TemplateName: "testtemplate1",
				Values: map[string]interface{}{
					"values": "{\"namespace\": \"kafka\", \"topic\": {\"name\":\"orders\", \"cluster\":\"my-cluster\", \"partitions\": 30,\"replicas\":   2, }}"},
			},
			expected1: ConfigResult{
				DefinitionName:    "testdef1",
				DefinitionVersion: "v1",
				ProfileName:       "testprofile1",
				InstanceName:      "testinstance1",
				ConfigName:        "testconfig1",
				TemplateName:      "testtemplate1",
				ConfigVersion:     1,
			},
			expected2: ConfigResult{
				DefinitionName:    "testdef1",
				DefinitionVersion: "v1",
				ProfileName:       "testprofile1",
				InstanceName:      "testinstance1",
				ConfigName:        "testconfig1",
				TemplateName:      "testtemplate1",
				ConfigVersion:     2,
			},
			expected3: ConfigResult{
				DefinitionName:    "testdef1",
				DefinitionVersion: "v1",
				ProfileName:       "testprofile1",
				InstanceName:      "testinstance1",
				ConfigName:        "testconfig1",
				TemplateName:      "testtemplate1",
				ConfigVersion:     3,
			},
			expected4: ConfigResult{
				DefinitionName:    "testdef1",
				DefinitionVersion: "v1",
				ProfileName:       "testprofile1",
				InstanceName:      "testinstance1",
				ConfigName:        "testconfig1",
				TemplateName:      "testtemplate1",
				ConfigVersion:     4,
			},
			expectedError: "",
			mockdb: &db.MockEtcdClient{
				Items: nil,
				Err:   nil,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.Etcd = testCase.mockdb
			db.DBconn = provideMockModelData(testCase.instanceID, testCase.rbName,
				testCase.rbVersion, testCase.profileName)
			resolve = func(rbName, rbVersion, profileName string, p Config, releaseName string) (configResourceList, error) {
				return configResourceList{}, nil
			}
			impl := NewConfigClient()
			got, err := impl.Create(testCase.instanceID, testCase.inp)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected1, got) == false {
					t.Errorf("Create returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected1)
				}
			}
			get, err := impl.Get(testCase.instanceID, testCase.inp.ConfigName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Get returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.inp, get) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", get, testCase.inp)
				}
			}
			getList, err := impl.List(testCase.instanceID)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("List returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("List returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual([]Config{testCase.inp}, getList) == false {
					t.Errorf("List returned unexpected body: got %v;"+
						" expected %v", getList, []Config{testCase.inp})
				}
			}
			got, err = impl.Update(testCase.instanceID, testCase.inp.ConfigName, testCase.inpUpdate1)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected2, got) == false {
					t.Errorf("Create returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected2)
				}
			}
			got, err = impl.Update(testCase.instanceID, testCase.inp.ConfigName, testCase.inpUpdate2)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected3, got) == false {
					t.Errorf("Create returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected3)
				}
			}
			got, err = impl.Delete(testCase.instanceID, testCase.inp.ConfigName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected4, got) == false {
					t.Errorf("Create returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected4)
				}
			}
			testCase.rollbackConfig.AnyOf.ConfigVersion = "2"
			err = impl.Rollback(testCase.instanceID, testCase.inp.ConfigName, testCase.rollbackConfig)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			}
			rollbackConfig, err := impl.Get(testCase.instanceID, testCase.inp.ConfigName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.inpUpdate1, rollbackConfig) == false {
					t.Errorf("Rollback config failed: got %v;"+
						" expected %v", rollbackConfig, testCase.inpUpdate1)
				}
			}
		})
	}
}

func main() {
}
