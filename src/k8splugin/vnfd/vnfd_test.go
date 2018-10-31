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

package vnfd

import (
	"k8splugin/db"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockDB struct {
	db.DatabaseConnection
	Items api.KVPairs
	Err   error
}

func (m *mockDB) CreateEntry(key string, value string) error {
	return m.Err
}

func (m *mockDB) ReadEntry(key string) (string, bool, error) {
	if m.Err != nil {
		return "", false, m.Err
	}

	for _, kvpair := range m.Items {
		if kvpair.Key == key {
			return string(kvpair.Value), true, nil
		}
	}

	return "", false, nil
}

func (m *mockDB) DeleteEntry(key string) error {
	return m.Err
}

func (m *mockDB) ReadAll(prefix string) ([]string, error) {
	if m.Err != nil {
		return []string{}, m.Err
	}

	var res []string

	for _, keypair := range m.Items {
		res = append(res, keypair.Key)
	}

	return res, nil
}

func TestCreate(t *testing.T) {
	testCases := []struct {
		label         string
		inp           VNFDefinition
		expectedError string
		mockdb        *mockDB
		expected      VNFDefinition
	}{
		{
			label: "Create VNF Definition",
			inp: VNFDefinition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testvnf",
				Description: "testvnf",
				ServiceType: "firewall",
			},
			expected: VNFDefinition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testvnf",
				Description: "testvnf",
				ServiceType: "firewall",
			},
			expectedError: "",
			mockdb:        &mockDB{},
		},
		{
			label:         "Failed Create VNF Definition",
			expectedError: "Error Creating Definition",
			mockdb: &mockDB{
				Err: pkgerrors.New("Error Creating Definition"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			vimpl := GetVNFDClient()
			got, err := vimpl.Create(testCase.inp)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Create VNF returned unexpected body: got %v;"+
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
		mockdb        *mockDB
		expected      []VNFDefinition
	}{
		{
			label: "List VNF Definition",
			expected: []VNFDefinition{
				{
					UUID:        "123e4567-e89b-12d3-a456-426655440000",
					Name:        "testvnf",
					Description: "testvnf",
					ServiceType: "firewall",
				},
				{
					UUID:        "123e4567-e89b-12d3-a456-426655441111",
					Name:        "testvnf2",
					Description: "testvnf2",
					ServiceType: "dns",
				},
			},
			expectedError: "",
			mockdb: &mockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key: "vnfd/123e4567-e89b-12d3-a456-426655440000",
						Value: []byte("{\"name\":\"testvnf\"," +
							"\"description\":\"testvnf\"," +
							"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
							"\"service-type\":\"firewall\"}"),
					},
					&api.KVPair{
						Key: "vnfd/123e4567-e89b-12d3-a456-426655441111",
						Value: []byte("{\"name\":\"testvnf2\"," +
							"\"description\":\"testvnf2\"," +
							"\"uuid\":\"123e4567-e89b-12d3-a456-426655441111\"," +
							"\"service-type\":\"dns\"}"),
					},
				},
			},
		},
		{
			label:         "List Error",
			expectedError: "DB Error",
			mockdb: &mockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			vimpl := GetVNFDClient()
			got, err := vimpl.List()
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("List returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("List returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("List VNF returned unexpected body: got %v;"+
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
		mockdb        *mockDB
		inp           string
		expected      VNFDefinition
	}{
		{
			label: "Get VNF Definition",
			inp:   "123e4567-e89b-12d3-a456-426655440000",
			expected: VNFDefinition{
				UUID:        "123e4567-e89b-12d3-a456-426655440000",
				Name:        "testvnf",
				Description: "testvnf",
				ServiceType: "firewall",
			},
			expectedError: "",
			mockdb: &mockDB{
				Items: api.KVPairs{
					&api.KVPair{
						Key: "vnfd/123e4567-e89b-12d3-a456-426655440000",
						Value: []byte("{\"name\":\"testvnf\"," +
							"\"description\":\"testvnf\"," +
							"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
							"\"service-type\":\"firewall\"}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb: &mockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			vimpl := GetVNFDClient()
			got, err := vimpl.Get(testCase.inp)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Get returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get VNF returned unexpected body: got %v;"+
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
		mockdb        *mockDB
		expected      []VNFDefinition
	}{
		{
			label:  "Delete VNF Definition",
			inp:    "123e4567-e89b-12d3-a456-426655440000",
			mockdb: &mockDB{},
		},
		{
			label:         "Delete Error",
			expectedError: "DB Error",
			mockdb: &mockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			vimpl := GetVNFDClient()
			err := vimpl.Delete(testCase.inp)
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
