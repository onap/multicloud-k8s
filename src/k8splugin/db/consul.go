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
	"os"

	"github.com/hashicorp/consul/api"
	pkgerrors "github.com/pkg/errors"
)

// ConsulKVStore defines the a subset of Consul DB operations
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
		config.Address = os.Getenv("DATABASE_IP") + ":8500"

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
	_, err := c.Read("test")
	if err != nil {
		return pkgerrors.New("[ERROR] Cannot talk to Datastore. Check if it is running/reachable.")
	}
	return nil
}

// Create is used to create a DB entry
func (c *ConsulStore) Create(key, value string) error {
	p := &api.KVPair{
		Key:   key,
		Value: []byte(value),
	}
	_, err := c.client.Put(p, nil)
	return err
}

// Read method returns the internalID for a particular externalID
func (c *ConsulStore) Read(key string) (string, error) {
	pair, _, err := c.client.Get(key, nil)
	if err != nil {
		return "", err
	}
	if pair == nil {
		return "", pkgerrors.New("No value found for ID: " + key)
	}
	return string(pair.Value), nil
}

// Delete method removes an internalID from the Database
func (c *ConsulStore) Delete(key string) error {
	_, err := c.client.Delete(key, nil)
	return err
}

// ReadAll is used to get all ExternalIDs in a namespace
func (c *ConsulStore) ReadAll(prefix string) ([]string, error) {
	pairs, _, err := c.client.List(prefix, nil)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, keypair := range pairs {
		result = append(result, keypair.Key)
	}

	return result, nil
}
