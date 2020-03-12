/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package contextdb

import (
	"context"
	"encoding/json"
	pkgerrors "github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
	"time"
)

// EtcdConfig Configuration values needed for Etcd Client
type EtcdConfig struct {
	Endpoint string
	CertFile string
	KeyFile  string
	CAFile   string
}

// EtcdClient for Etcd
type EtcdClient struct {
	cli      *clientv3.Client
	endpoint string
}

// Etcd For Mocking purposes
type Etcd interface {
	Put(ctx context.Context, key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Delete(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
}

var getEtcd = func(e *EtcdClient) Etcd {
	return e.cli
}

// NewEtcdClient function initializes Etcd client
func NewEtcdClient(store *clientv3.Client, c EtcdConfig) (ContextDb, error) {
	var endpoint string
	if store == nil {
		tlsInfo := transport.TLSInfo{
			CertFile: c.CertFile,
			KeyFile:  c.KeyFile,
			CAFile:   c.CAFile,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return nil, pkgerrors.Errorf("Error creating etcd TLSInfo: %s", err.Error())
		}
		// NOTE: Client relies on nil tlsConfig
		// for non-secure connections, update the implicit variable
		if len(c.CertFile) == 0 && len(c.KeyFile) == 0 && len(c.CAFile) == 0 {
			tlsConfig = nil
		}
		endpoint = ""
		if tlsConfig == nil {
			endpoint = "http://" + c.Endpoint + ":2379"
		} else {
			endpoint = "https://" + c.Endpoint + ":2379"
		}

		store, err = clientv3.New(clientv3.Config{
			Endpoints:   []string{endpoint},
			DialTimeout: 5 * time.Second,
			TLS:         tlsConfig,
		})
		if err != nil {
			return nil, pkgerrors.Errorf("Error creating etcd client: %s", err.Error())
		}
	}

	return &EtcdClient{
		cli:      store,
		endpoint: endpoint,
	}, nil
}

// Put values in Etcd DB
func (e *EtcdClient) Put(key string, value interface{}) error {
	cli := getEtcd(e)
	if cli == nil {
		return pkgerrors.Errorf("Etcd Client not initialized")
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
	_, err = cli.Put(context.Background(), key, string(v))
	if err != nil {
		return pkgerrors.Errorf("Error creating etcd entry: %s", err.Error())
	}
	return nil
}

// Get values from Etcd DB and decodes from json
func (e *EtcdClient) Get(key string, value interface{}) error {
	cli := getEtcd(e)
	if cli == nil {
		return pkgerrors.Errorf("Etcd Client not initialized")
	}
	if key == "" {
		return pkgerrors.Errorf("Key is null")
	}
	if value == nil {
		return pkgerrors.Errorf("Value is nil")
	}
	getResp, err := cli.Get(context.Background(), key)
	if err != nil {
		return pkgerrors.Errorf("Error getting etcd entry: %s", err.Error())
	}
	if getResp.Count == 0 {
		return pkgerrors.Errorf("Key doesn't exist")
	}
	return json.Unmarshal(getResp.Kvs[0].Value, value)
}

// GetAllKeys values from Etcd DB
func (e *EtcdClient) GetAllKeys(key string) ([]string, error) {
	cli := getEtcd(e)
	if cli == nil {
		return nil, pkgerrors.Errorf("Etcd Client not initialized")
	}
	getResp, err := cli.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting etcd entry: %s", err.Error())
	}
	if getResp.Count == 0 {
		return nil, pkgerrors.Errorf("Key doesn't exist")
	}
	var keys []string
	for _, ev := range getResp.Kvs {
		keys = append(keys, string(ev.Key))
	}
	return keys, nil
}

// DeleteAll keys from Etcd DB
func (e *EtcdClient) DeleteAll(key string) error {
	cli := getEtcd(e)
	if cli == nil {
		return pkgerrors.Errorf("Etcd Client not initialized")
	}
	_, err := cli.Delete(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return pkgerrors.Errorf("Delete failed etcd entry: %s", err.Error())
	}
	return nil
}

// Delete values from Etcd DB
func (e *EtcdClient) Delete(key string) error {
	cli := getEtcd(e)
	if cli == nil {
		return pkgerrors.Errorf("Etcd Client not initialized")
	}
	_, err := cli.Delete(context.Background(), key)
	if err != nil {
		return pkgerrors.Errorf("Delete failed etcd entry: %s", err.Error())
	}
	return nil
}

// HealthCheck for checking health of the etcd cluster
func (e *EtcdClient) HealthCheck() error {
	return nil
}
