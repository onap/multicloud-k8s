/*
Copyright 2018 Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package db

import (
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/consul/api"
	pkgerrors "github.com/pkg/errors"
)

type mockConsulKVStore struct {
	Items api.KVPairs
	Err   error
}

func (c *mockConsulKVStore) Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error) {
	return nil, c.Err
}

func (c *mockConsulKVStore) Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	if c.Err != nil {
		return nil, nil, c.Err
	}
	for _, kvpair := range c.Items {
		if kvpair.Key == key {
			return kvpair, nil, nil
		}
	}
	return nil, nil, nil
}

func (c *mockConsulKVStore) Delete(key string, w *api.WriteOptions) (*api.WriteMeta, error) {
	return nil, c.Err
}

func (c *mockConsulKVStore) List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error) {
	if c.Err != nil {
		return nil, nil, c.Err
	}
	return c.Items, nil, nil
}

func TestConsulHealthCheck(t *testing.T) {
	testCases := []struct {
		label         string
		mock          *mockConsulKVStore
		expectedError string
	}{
		{
			label: "Sucessful health check Consul Database",
			mock: &mockConsulKVStore{
				Items: api.KVPairs{
					&api.KVPair{
						Key:   "test-key",
						Value: nil,
					},
				},
			},
		},
		{
			label: "Fail connectivity to Consul Database",
			mock: &mockConsulKVStore{
				Err: pkgerrors.New("Timeout"),
			},
			expectedError: "Cannot talk to Datastore. Check if it is running/reachable.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			client, _ := NewConsulStore(testCase.mock)
			err := client.HealthCheck()
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("HealthCheck method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("HealthCheck method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestConsulCreate(t *testing.T) {
	testCases := []struct {
		label         string
		input         map[string]string
		key           Key
		mock          *mockConsulKVStore
		expectedError string
	}{
		{
			label: "Sucessful register a record to Consul Database",
			key:   MockKey{Key: "test-key"},
			input: map[string]string{"root": "rbinst", "tag": "data", "value": "test-value"},
			mock:  &mockConsulKVStore{},
		},
		{
			label: "Fail to create a new record in Consul Database",
			key:   MockKey{Key: "test-key"},
			input: map[string]string{"root": "rbinst", "tag": "data", "value": "test-value"},
			mock: &mockConsulKVStore{
				Err: pkgerrors.New("DB error"),
			},
			expectedError: "DB error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			client, _ := NewConsulStore(testCase.mock)
			err := client.Create(testCase.input["root"], testCase.key,
				testCase.input["tag"], testCase.input["value"])
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Create method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestConsulRead(t *testing.T) {
	testCases := []struct {
		label          string
		input          map[string]string
		key            Key
		mock           *mockConsulKVStore
		expectedError  string
		expectedResult string
	}{
		{
			label: "Sucessful retrieve a record from Consul Database",
			key:   MockKey{Key: "test"},
			input: map[string]string{"root": "rbinst", "tag": "data"},
			mock: &mockConsulKVStore{
				Items: api.KVPairs{
					&api.KVPair{
						Key:   "rbinst/test/data",
						Value: []byte("test-value"),
					},
				},
			},
			expectedResult: "test-value",
		},
		{
			label: "Fail retrieve a non-existing record from Consul Database",
			key:   MockKey{Key: "test-key"},
			input: map[string]string{"root": "rbinst", "tag": "data"},
			mock:  &mockConsulKVStore{},
		},
		{
			label: "Fail retrieve a record from Consul Database",
			key:   MockKey{Key: "test-key"},
			input: map[string]string{"root": "rbinst", "tag": "data"},
			mock: &mockConsulKVStore{
				Err: pkgerrors.New("DB error"),
			},
			expectedError: "DB error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			client, _ := NewConsulStore(testCase.mock)
			result, err := client.Read(testCase.input["root"], testCase.key,
				testCase.input["tag"])
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Read method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Read method returned an error (%s)", err)
				}
			} else {
				if testCase.expectedError != "" && testCase.expectedResult == "" {
					t.Fatalf("Read method was expecting \"%s\" error message", testCase.expectedError)
				}
				if !reflect.DeepEqual(testCase.expectedResult, string(result)) {

					t.Fatalf("Read method returned: \n%v\n while expected value was: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}

func TestConsulDelete(t *testing.T) {
	testCases := []struct {
		label         string
		input         map[string]string
		key           Key
		mock          *mockConsulKVStore
		expectedError string
	}{
		{
			label: "Sucessful delete a record to Consul Database",
			key:   MockKey{Key: "test-key"},
			input: map[string]string{"root": "rbinst", "tag": "data"},
			mock:  &mockConsulKVStore{},
		},
		{
			label: "Fail to delete a record in Consul Database",
			key:   MockKey{Key: "test-key"},
			mock: &mockConsulKVStore{
				Err: pkgerrors.New("DB error"),
			},
			expectedError: "DB error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			client, _ := NewConsulStore(testCase.mock)
			err := client.Delete(testCase.input["root"], testCase.key,
				testCase.input["tag"])
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Delete method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Delete method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestConsulReadAll(t *testing.T) {
	testCases := []struct {
		label          string
		input          map[string]string
		mock           *mockConsulKVStore
		expectedError  string
		expectedResult map[string][]byte
	}{
		{
			label: "Sucessful retrieve a list from Consul Database",
			input: map[string]string{"root": "rbinst", "key": "test-key",
				"tag": "data"},
			mock: &mockConsulKVStore{
				Items: api.KVPairs{
					&api.KVPair{
						Key:   "test",
						Value: []byte("test-value"),
					},
					&api.KVPair{
						Key:   "test2",
						Value: []byte("test-value2"),
					},
				},
			},
			expectedResult: map[string][]byte{"test": []byte("test-value"),
				"test2": []byte("test-value2")},
		},
		{
			label: "Sucessful retrieve an empty list from Consul Database",
			input: map[string]string{"root": "rbinst", "key": "test-key",
				"tag": "data"},
			mock:           &mockConsulKVStore{},
			expectedResult: map[string][]byte{},
		},
		{
			label: "Fail retrieve a record from Consul Database",
			input: map[string]string{"root": "rbinst", "key": "test-key",
				"tag": "data"},
			mock: &mockConsulKVStore{
				Err: pkgerrors.New("DB error"),
			},
			expectedError: "DB error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			client, _ := NewConsulStore(testCase.mock)
			result, err := client.ReadAll(testCase.input["root"],
				testCase.input["tag"])
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("ReadAll method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("ReadAll method returned an error (%s)", err)
				}
			} else {
				if testCase.expectedError != "" && testCase.expectedResult == nil {
					t.Fatalf("ReadAll method was expecting \"%s\" error message", testCase.expectedError)
				}
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("ReadAll method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}
