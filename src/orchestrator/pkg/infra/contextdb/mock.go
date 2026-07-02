/*
Copyright 2020 Intel Corporation.
Copyright 2026 Deutsche Telekom AG
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
	"encoding/json"
	"strings"

	pkgerrors "github.com/pkg/errors"
)

// MockEtcd is an in-memory implementation of the ContextDb interface for use
// in unit tests. It mirrors the JSON marshal/unmarshal and prefix semantics of
// the real etcd-backed client (see etcd.go) so that code exercising an
// AppContext behaves the same against the mock as it would against etcd.
//
// Set Err to force every operation to return that error, which is convenient
// for exercising error-handling paths.
type MockEtcd struct {
	Items map[string]string
	Err   error
}

func (c *MockEtcd) put(key, value string) {
	if c.Items == nil {
		c.Items = make(map[string]string)
	}
	c.Items[key] = value
}

// Put marshals value to JSON and stores it under key.
func (c *MockEtcd) Put(key string, value interface{}) error {
	if c.Err != nil {
		return c.Err
	}
	if key == "" {
		return pkgerrors.Errorf("Key is null")
	}
	if value == nil {
		return pkgerrors.Errorf("Value is nil")
	}
	v, err := json.Marshal(value)
	if err != nil {
		return pkgerrors.Errorf("Json Marshal error: %s", err.Error())
	}
	c.put(key, string(v))
	return nil
}

// Get retrieves the value stored under key and unmarshals it into value, which
// must be a non-nil pointer. It returns an error if the key does not exist.
func (c *MockEtcd) Get(key string, value interface{}) error {
	if c.Err != nil {
		return c.Err
	}
	if key == "" {
		return pkgerrors.Errorf("Key is null")
	}
	if value == nil {
		return pkgerrors.Errorf("Value is nil")
	}
	v, ok := c.Items[key]
	if !ok {
		return pkgerrors.Errorf("Key doesn't exist")
	}
	return json.Unmarshal([]byte(v), value)
}

// Delete removes the exact key.
func (c *MockEtcd) Delete(key string) error {
	if c.Err != nil {
		return c.Err
	}
	delete(c.Items, key)
	return nil
}

// DeleteAll removes every key sharing the given prefix.
func (c *MockEtcd) DeleteAll(key string) error {
	if c.Err != nil {
		return c.Err
	}
	for k := range c.Items {
		if strings.HasPrefix(k, key) {
			delete(c.Items, k)
		}
	}
	return nil
}

// GetAllKeys returns every key sharing the given prefix. Like the real client,
// it returns an error when no key matches.
func (c *MockEtcd) GetAllKeys(path string) ([]string, error) {
	if c.Err != nil {
		return nil, c.Err
	}
	var keys []string
	for k := range c.Items {
		if strings.HasPrefix(k, path) {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return nil, pkgerrors.Errorf("Key doesn't exist")
	}
	return keys, nil
}

func (c *MockEtcd) HealthCheck() error {
	return nil
}
