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
	pkgerrors "github.com/pkg/errors"
)

type MockEtcd struct {
	Items map[string]interface{}
	Err   error
}

func (c *MockEtcd) Put(key string, value interface{}) error {
	if c.Items == nil {
		c.Items = make(map[string]interface{})
	}
	c.Items[key] = value
	return c.Err
}

func (c *MockEtcd) Get(key string, value interface{}) error {
	for kvKey, kvValue := range c.Items {
		if kvKey == key {
			value = kvValue
			return nil
		}
	}
	return pkgerrors.Errorf("Key doesn't exist")
}

func (c *MockEtcd) Delete(key string) error {
	delete(c.Items, key)
	return c.Err
}

func (c *MockEtcd) GetAllKeys(path string) ([]string, error) {
	var keys []string
	for k := range c.Items {
		keys = append(keys, string(k))
	}
	return keys, nil
}

func (e *MockEtcd) HealthCheck() error {
	return nil
}
