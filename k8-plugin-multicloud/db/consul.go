package db

import (
	consulapi "github.com/hashicorp/consul/api"
	pkgerrors "github.com/pkg/errors"
	"os"
)

// ConsulDB is an implementation of the DatabaseConnection interface
type ConsulDB struct {
	consulClient *consulapi.Client
}

// InitializeDatabase initialized the initial steps
func (c *ConsulDB) InitializeDatabase() error {
	if os.Getenv("DATABASE_IP") == "" {
		return pkgerrors.New("DATABASE_IP environment variable not set.")
	}
	config := consulapi.DefaultConfig()
	config.Address = os.Getenv("DATABASE_IP") + ":8500"

	client, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}
	c.consulClient = client
	return nil
}

// CheckDatabase checks if the database is running
func (c *ConsulDB) CheckDatabase() error {
	kv := c.consulClient.KV()
	_, _, err := kv.Get("test", nil)
	if err != nil {
		return pkgerrors.New("[ERROR] Cannot talk to Datastore. Check if it is running/reachable.")
	}
	return nil
}

// CreateEntry is used to create a DB entry
func (c *ConsulDB) CreateEntry(key string, value string) error {
	kv := c.consulClient.KV()

	p := &consulapi.KVPair{Key: key, Value: []byte(value)}

	_, err := kv.Put(p, nil)

	if err != nil {
		return err
	}

	return nil
}

// ReadEntry returns the internalID for a particular externalID is present in a namespace
func (c *ConsulDB) ReadEntry(key string) (string, bool, error) {

	kv := c.consulClient.KV()

	pair, _, err := kv.Get(key, nil)

	if pair == nil {
		return string("No value found for ID: " + key), false, err
	}
	return string(pair.Value), true, err
}

// DeleteEntry is used to delete an ID
func (c *ConsulDB) DeleteEntry(key string) error {

	kv := c.consulClient.KV()

	_, err := kv.Delete(key, nil)

	if err != nil {
		return err
	}

	return nil
}

// ReadAll is used to get all ExternalIDs in a namespace
func (c *ConsulDB) ReadAll(prefix string) ([]string, error) {
	kv := c.consulClient.KV()

	pairs, _, err := kv.List(prefix, nil)

	if len(pairs) == 0 {
		return []string{""}, err
	}

	var res []string

	for _, keypair := range pairs {
		res = append(res, keypair.Key)
	}

	return res, err
}
