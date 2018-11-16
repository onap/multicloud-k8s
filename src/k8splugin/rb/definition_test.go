// +build unit

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

package rb

import (
	"k8splugin/db"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
	pkgerrors "github.com/pkg/errors"
)

func TestCreate(t *testing.T) {
	testCases := []struct {
		label         string
		inp           Definition
		expectedError string
		mockdb        *db.MockDB
		expected      Definition
	}{
		{
			label: "Create Resource Bundle Definition",
			inp: Definition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testresourcebundle",
				Description: "testresourcebundle",
				ServiceType: "firewall",
			},
			expected: Definition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testresourcebundle",
				Description: "testresourcebundle",
				ServiceType: "firewall",
			},
			expectedError: "",
			mockdb:        &db.MockDB{},
		},
		{
			label:         "Failed Create Resource Bundle Definition",
			expectedError: "Error Creating Definition",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Definition"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewDefinitionClient()
			got, err := impl.Create(testCase.inp)
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

func TestList(t *testing.T) {

	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      []Definition
	}{
		{
			label: "List Resource Bundle Definition",
			expected: []Definition{
				{
					UUID:        "123e4567-e89b-12d3-a456-426655440000",
					Name:        "testresourcebundle",
					Description: "testresourcebundle",
					ServiceType: "firewall",
				},
				{
					UUID:        "123e4567-e89b-12d3-a456-426655441111",
					Name:        "testresourcebundle2",
					Description: "testresourcebundle2",
					ServiceType: "dns",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key: "rb/def/123e4567-e89b-12d3-a456-426655440000",
						Value: []byte("{\"name\":\"testresourcebundle\"," +
							"\"description\":\"testresourcebundle\"," +
							"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
							"\"service-type\":\"firewall\"}"),
					},
					&api.KVPair{
						Key: "rb/def/123e4567-e89b-12d3-a456-426655441111",
						Value: []byte("{\"name\":\"testresourcebundle2\"," +
							"\"description\":\"testresourcebundle2\"," +
							"\"uuid\":\"123e4567-e89b-12d3-a456-426655441111\"," +
							"\"service-type\":\"dns\"}"),
					},
				},
			},
		},
		{
			label:         "List Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewDefinitionClient()
			got, err := impl.List()
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("List returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("List returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("List Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {

	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      Definition
	}{
		{
			label: "Get Resource Bundle Definition",
			inp:   "123e4567-e89b-12d3-a456-426655440000",
			expected: Definition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testresourcebundle",
				Description: "testresourcebundle",
				ServiceType: "firewall",
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key: "rb/def/123e4567-e89b-12d3-a456-426655440000",
						Value: []byte("{\"name\":\"testresourcebundle\"," +
							"\"description\":\"testresourcebundle\"," +
							"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
							"\"service-type\":\"firewall\"}"),
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
			impl := NewDefinitionClient()
			got, err := impl.Get(testCase.inp)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Get returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {

	testCases := []struct {
		label         string
		inp           string
		expectedError string
		mockdb        *db.MockDB
		expected      []Definition
	}{
		{
			label:  "Delete Resource Bundle Definition",
			inp:    "123e4567-e89b-12d3-a456-426655440000",
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
			impl := NewDefinitionClient()
			err := impl.Delete(testCase.inp)
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
