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
	pkgerrors "github.com/pkg/errors"
)

type MockEtcdClient struct {
	Items map[string]string
	Err   error
}

func (c *MockEtcdClient) Put(key, value string) error {
	if c.Items == nil {
		c.Items = make(map[string]string)
	}
	c.Items[key] = value
	return c.Err
}

func (c *MockEtcdClient) Get(key string) ([]byte, error) {
	for kvKey, kvValue := range c.Items {
		if kvKey == key {
			return []byte(kvValue), nil
		}
	}
	return nil, pkgerrors.Errorf("Key doesn't exist")
}

func (c *MockEtcdClient) Delete(key string) error {
	delete(c.Items, key)
	return c.Err
}
