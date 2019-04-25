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
	"encoding/json"
	pkgerrors "github.com/pkg/errors"
)

type MockKey struct {
	Key string
}

func (m MockKey) String() string {
	return m.Key
}

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type MockDB struct {
	Store
	Items map[string]map[string][]byte
	Err   error
}

func (m *MockDB) HealthCheck() error {
	return m.Err
}

func (m *MockDB) Create(table string, key Key, tag string, data interface{}) error {
	return m.Err
}

func (m *MockDB) Update(table string, key Key, tag string, data interface{}) error {
	return m.Err
}

// MockDB uses simple JSON and not BSON
func (m *MockDB) Unmarshal(inp []byte, out interface{}) error {
	err := json.Unmarshal(inp, out)
	if err != nil {
		return pkgerrors.Wrap(err, "Unmarshaling json")
	}
	return nil
}

func (m *MockDB) Read(table string, key Key, tag string) ([]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	for k, v := range m.Items {
		if k == key.String() {
			return v[tag], nil
		}
	}

	return nil, m.Err
}

func (m *MockDB) Delete(table string, key Key, tag string) error {
	return m.Err
}

func (m *MockDB) ReadAll(table string, tag string) (map[string][]byte, error) {
	if m.Err != nil {
		return nil, m.Err
	}

	ret := make(map[string][]byte)

	for k, v := range m.Items {
		for k1, v1 := range v {
			if k1 == tag {
				ret[k] = v1
			}
		}
	}

	return ret, nil
}
