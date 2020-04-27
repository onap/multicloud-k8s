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

package app

import (
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"reflect"
	"strings"
	"testing"
	//	pkgerrors "github.com/pkg/errors"
)

func TestCreateConfig(t *testing.T) {
	testCases := []struct {
		label         string
		rbName        string
		rbVersion     string
		profileName   string
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
			inp: Config{
				ConfigName:   "testconfig1",
				TemplateName: "testtemplate1",
				Values: map[string]interface{}{
					"values": "{\"namespace\": \"kafka\", \"topic\": {\"name\":\"orders\", \"cluster\":\"my-cluster\", \"partitions\": 10,\"replicas\":   2, }}"},
			},
			expected: ConfigResult{
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
			resolve = func(rbName, rbVersion, profileName string, p Config) (configResourceList, error) {
				return configResourceList{}, nil
			}
			impl := NewConfigClient()
			got, err := impl.Create(testCase.rbName, testCase.rbVersion, testCase.profileName, testCase.inp)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Create Resource Bundle returned unexpected body: got %v;"+
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
				ConfigName:        "testconfig1",
				TemplateName:      "testtemplate1",
				ConfigVersion:     1,
			},
			expected2: ConfigResult{
				DefinitionName:    "testdef1",
				DefinitionVersion: "v1",
				ProfileName:       "testprofile1",
				ConfigName:        "testconfig1",
				TemplateName:      "testtemplate1",
				ConfigVersion:     2,
			},
			expected3: ConfigResult{
				DefinitionName:    "testdef1",
				DefinitionVersion: "v1",
				ProfileName:       "testprofile1",
				ConfigName:        "testconfig1",
				TemplateName:      "testtemplate1",
				ConfigVersion:     3,
			},
			expected4: ConfigResult{
				DefinitionName:    "testdef1",
				DefinitionVersion: "v1",
				ProfileName:       "testprofile1",
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
			resolve = func(rbName, rbVersion, profileName string, p Config) (configResourceList, error) {
				return configResourceList{}, nil
			}
			impl := NewConfigClient()
			got, err := impl.Create(testCase.rbName, testCase.rbVersion, testCase.profileName, testCase.inp)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected1, got) == false {
					t.Errorf("Create Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected1)
				}
			}
			got, err = impl.Update(testCase.rbName, testCase.rbVersion, testCase.profileName, testCase.inp.ConfigName, testCase.inpUpdate1)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected2, got) == false {
					t.Errorf("Create Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected2)
				}
			}
			got, err = impl.Update(testCase.rbName, testCase.rbVersion, testCase.profileName, testCase.inp.ConfigName, testCase.inpUpdate2)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected3, got) == false {
					t.Errorf("Create Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected3)
				}
			}
			got, err = impl.Delete(testCase.rbName, testCase.rbVersion, testCase.profileName, testCase.inp.ConfigName)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected4, got) == false {
					t.Errorf("Create Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected4)
				}
			}
			testCase.rollbackConfig.AnyOf.ConfigVersion = "2"
			err = impl.Rollback(testCase.rbName, testCase.rbVersion, testCase.profileName, testCase.rollbackConfig)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			}
			rollbackConfig, err := impl.Get(testCase.rbName, testCase.rbVersion, testCase.profileName, testCase.inp.ConfigName)
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
