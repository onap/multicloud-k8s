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
	k8sconfig "github.com/onap/multicloud-k8s/src/k8splugin/internal/config"

	"github.com/hashicorp/consul/api"
	pkgerrors "github.com/pkg/errors"
)

// ConsulKVStore defines the a subset of Consul DB operations
// Note: This interface is defined mainly for allowing mock testing
type ConsulKVStore interface {
	List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
	Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
	Put(p *api.KVPair, q *api.WriteOptions) (*api.WriteMeta, error)
	Delete(key string, w *api.WriteOptions) (*api.WriteMeta, error)
}

// ConsulStore is an implementation of the ConsulKVStore interface
type ConsulStore struct {
	client ConsulKVStore
}

// NewConsulStore initializes a Consul Store instance using the default values
func NewConsulStore(store ConsulKVStore) (Store, error) {
	if store == nil {
		config := api.DefaultConfig()
		config.Address = k8sconfig.GetConfiguration().DatabaseAddress + ":8500"

		consulClient, err := api.NewClient(config)
		if err != nil {
			return nil, err
		}
		store = consulClient.KV()
	}

	return &ConsulStore{
		client: store,
	}, nil
}

// HealthCheck verifies if the database is up and running
func (c *ConsulStore) HealthCheck() error {
	_, _, err := c.client.Get("test", nil)
	if err != nil {
		return pkgerrors.New("[ERROR] Cannot talk to Datastore. Check if it is running/reachable.")
	}
	return nil
}

// Unmarshal implements any unmarshaling that is needed when using consul
func (c *ConsulStore) Unmarshal(inp []byte, out interface{}) error {
	return nil
}

// Create is used to create a DB entry
func (c *ConsulStore) Create(root string, key Key, tag string, data interface{}) error {

	//Convert to string as Consul only supports string based keys
	k := key.String()
	if k == "" {
		return pkgerrors.New("Key.String() returned an empty string")
	}

	value, err := Serialize(data)
	if err != nil {
		return pkgerrors.Wrap(err, "Serializing input data")
	}

	p := &api.KVPair{
		Key:   k,
		Value: []byte(value),
	}
	_, err = c.client.Put(p, nil)
	return err
}

// Update is used to update a DB entry
func (c *ConsulStore) Update(root string, key Key, tag string, data interface{}) error {
	return c.Create(root, key, tag, data)
}

// Read method returns the internalID for a particular externalID
func (c *ConsulStore) Read(root string, key Key, tag string) ([]byte, error) {

	//Convert to string as Consul only supports string based keys
	k := key.String()
	if k == "" {
		return nil, pkgerrors.New("Key.String() returned an empty string")
	}

	k = root + "/" + k + "/" + tag
	pair, _, err := c.client.Get(k, nil)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, nil
	}
	return pair.Value, nil
}

// Delete method removes an internalID from the Database
func (c *ConsulStore) Delete(root string, key Key, tag string) error {

	//Convert to string as Consul only supports string based keys
	k := key.String()
	if k == "" {
		return pkgerrors.New("Key.String() returned an empty string")
	}
	_, err := c.client.Delete(k, nil)
	return err
}

// ReadAll is used to get all ExternalIDs in a namespace
func (c *ConsulStore) ReadAll(root string, tag string) (map[string][]byte, error) {
	pairs, _, err := c.client.List(root, nil)
	if err != nil {
		return nil, err
	}

	//TODO: Filter results by tag and return it
	result := make(map[string][]byte)
	for _, keypair := range pairs {
		result[keypair.Key] = keypair.Value
	}

	return result, nil
}
