// +build unit

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
	"github.com/hashicorp/consul/api"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type MockDB struct {
	Store
	Items api.KVPairs
	Err   error
}

func (m *MockDB) Create(key string, value string) error {
	return m.Err
}

func (m *MockDB) Read(key string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}

	for _, kvpair := range m.Items {
		if kvpair.Key == key {
			return string(kvpair.Value), nil
		}
	}

	return "", nil
}

func (m *MockDB) Delete(key string) error {
	return m.Err
}

func (m *MockDB) ReadAll(prefix string) ([]string, error) {
	if m.Err != nil {
		return []string{}, m.Err
	}

	var res []string

	for _, keypair := range m.Items {
		res = append(res, keypair.Key)
	}

	return res, nil
}
