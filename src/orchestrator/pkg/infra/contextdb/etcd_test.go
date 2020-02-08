/*
Copyright 2020 Intel Corporation.
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

package contextdb

import (
	"context"
	mvccpb "github.com/coreos/etcd/mvcc/mvccpb"
	pkgerrors "github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"strings"
	"testing"
)

type kv struct {
	Key   []byte
	Value []byte
}

// MockEtcdClient for mocking etcd
type MockEtcdClient struct {
	Kvs   []*mvccpb.KeyValue
	Count int64
	Err   error
}

// Mocking only Single Value
// Put function
func (e *MockEtcdClient) Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	var m mvccpb.KeyValue
	m.Key = []byte(key)
	m.Value = []byte(val)
	e.Count = e.Count + 1
	e.Kvs = append(e.Kvs, &m)
	return &clientv3.PutResponse{}, e.Err
}

// Get function
func (e *MockEtcdClient) Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	var g clientv3.GetResponse
	g.Kvs = e.Kvs
	g.Count = e.Count
	return &g, e.Err
}

// Delete function
func (e *MockEtcdClient) Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	return &clientv3.DeleteResponse{}, e.Err
}

type testStruct struct {
	Name string `json:"name"`
	Num  int    `json:"num"`
}

// TestPut test Put
func TestPut(t *testing.T) {
	testCases := []struct {
		label         string
		mockEtcd      *MockEtcdClient
		expectedError string
		key           string
		value         *testStruct
	}{
		{
			label:    "Success Case",
			mockEtcd: &MockEtcdClient{},
			key:      "test1",
			value:    &testStruct{Name: "test", Num: 5},
		},
		{
			label:         "Key is null",
			mockEtcd:      &MockEtcdClient{},
			key:           "",
			expectedError: "Key is null",
		},
		{
			label:         "Value is nil",
			mockEtcd:      &MockEtcdClient{},
			key:           "test1",
			value:         nil,
			expectedError: "Value is nil",
		},
		{
			label:         "Error creating etcd entry",
			mockEtcd:      &MockEtcdClient{Err: pkgerrors.New("DB Error")},
			key:           "test1",
			value:         &testStruct{Name: "test", Num: 5},
			expectedError: "Error creating etcd entry: DB Error",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			cli, _ := NewEtcdClient(&clientv3.Client{}, EtcdConfig{})
			getEtcd = func(e *EtcdClient) Etcd {
				return testCase.mockEtcd
			}
			err := cli.Put(testCase.key, testCase.value)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}

		})
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		label         string
		mockEtcd      *MockEtcdClient
		expectedError string
		key           string
		value         *testStruct
	}{
		{
			label:         "Key is null",
			mockEtcd:      &MockEtcdClient{},
			key:           "",
			value:         nil,
			expectedError: "Key is null",
		},
		{
			label:         "Key doesn't exist",
			mockEtcd:      &MockEtcdClient{},
			key:           "test1",
			value:         &testStruct{},
			expectedError: "Key doesn't exist",
		},
		{
			label:         "Error getting etcd entry",
			mockEtcd:      &MockEtcdClient{Err: pkgerrors.New("DB Error")},
			key:           "test1",
			value:         &testStruct{},
			expectedError: "Error getting etcd entry: DB Error",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			cli, _ := NewEtcdClient(&clientv3.Client{}, EtcdConfig{})
			getEtcd = func(e *EtcdClient) Etcd {
				return testCase.mockEtcd
			}
			err := cli.Get(testCase.key, testCase.value)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}

		})
	}
}

func TestGetString(t *testing.T) {
	testCases := []struct {
		label         string
		mockEtcd      *MockEtcdClient
		expectedError string
		value         string
	}{
		{
			label:    "Success Case",
			mockEtcd: &MockEtcdClient{},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			cli, _ := NewEtcdClient(&clientv3.Client{}, EtcdConfig{})
			getEtcd = func(e *EtcdClient) Etcd {
				return testCase.mockEtcd
			}
			err := cli.Put("test", "test1")
			if err != nil {
				t.Error("Test failed", err)
			}
			var s string
			err = cli.Get("test", &s)
			if err != nil {
				t.Error("Test failed", err)
			}
			if "test1" != s {
				t.Error("Get Failed")
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		label         string
		mockEtcd      *MockEtcdClient
		expectedError string
	}{
		{
			label:    "Success Case",
			mockEtcd: &MockEtcdClient{},
		},
		{
			label:         "Delete failed etcd entry",
			mockEtcd:      &MockEtcdClient{Err: pkgerrors.New("DB Error")},
			expectedError: "Delete failed etcd entry: DB Error",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			cli, _ := NewEtcdClient(&clientv3.Client{}, EtcdConfig{})
			getEtcd = func(e *EtcdClient) Etcd {
				return testCase.mockEtcd
			}
			err := cli.Delete("test")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}

		})
	}
}

func TestGetAll(t *testing.T) {
	testCases := []struct {
		label         string
		mockEtcd      *MockEtcdClient
		expectedError string
	}{
		{
			label:         "Key doesn't exist",
			mockEtcd:      &MockEtcdClient{},
			expectedError: "Key doesn't exist",
		},
		{
			label:         "Error getting etcd entry",
			mockEtcd:      &MockEtcdClient{Err: pkgerrors.New("DB Error")},
			expectedError: "Error getting etcd entry: DB Error",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			cli, _ := NewEtcdClient(&clientv3.Client{}, EtcdConfig{})
			getEtcd = func(e *EtcdClient) Etcd {
				return testCase.mockEtcd
			}
			_, err := cli.GetAllKeys("test")
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Method returned an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}
