/*
 * Copyright 2018 Intel Corporation, Inc
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

package db

import (
	"context"
	"time"

	pkgerrors "github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/pkg/transport"
)

// EtcdConfig Configuration values needed for Etcd Client
type EtcdConfig struct {
	Endpoint string
	CertFile string
	KeyFile  string
	CAFile   string
}

// EtcdStore Interface needed for mocking
type EtcdStore interface {
	Get(key string) ([]byte, error)
	GetAll(key string) ([][]byte, error)
	Put(key, value string) error
	Delete(key string) error
	DeletePrefix(keyPrefix string) error
}

// EtcdClient for Etcd
type EtcdClient struct {
	cli *clientv3.Client
}

// Etcd handle for interface
var Etcd EtcdStore

// NewEtcdClient function initializes Etcd client
func NewEtcdClient(store *clientv3.Client, c EtcdConfig) error {
	var err error
	Etcd, err = newClient(store, c)
	return err
}

func newClient(store *clientv3.Client, c EtcdConfig) (EtcdClient, error) {
	if store == nil {
		tlsInfo := transport.TLSInfo{
			CertFile: c.CertFile,
			KeyFile:  c.KeyFile,
			CAFile:   c.CAFile,
		}
		tlsConfig, err := tlsInfo.ClientConfig()
		if err != nil {
			return EtcdClient{}, pkgerrors.Errorf("Error creating etcd TLSInfo: %s", err.Error())
		}
		// NOTE: Client relies on nil tlsConfig
		// for non-secure connections, update the implicit variable
		if len(c.CertFile) == 0 && len(c.KeyFile) == 0 && len(c.CAFile) == 0 {
			tlsConfig = nil
		}
		endpoint := ""
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
			return EtcdClient{}, pkgerrors.Errorf("Error creating etcd client: %s", err.Error())
		}
	}

	return EtcdClient{
		cli: store,
	}, nil
}

// Put values in Etcd DB
func (e EtcdClient) Put(key, value string) error {

	if e.cli == nil {
		return pkgerrors.Errorf("Etcd Client not initialized")
	}
	_, err := e.cli.Put(context.Background(), key, value)
	if err != nil {
		return pkgerrors.Errorf("Error creating etcd entry: %s", err.Error())
	}
	return nil
}

// Get values from Etcd DB
func (e EtcdClient) Get(key string) ([]byte, error) {

	if e.cli == nil {
		return nil, pkgerrors.Errorf("Etcd Client not initialized")
	}
	getResp, err := e.cli.Get(context.Background(), key)
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting etcd entry: %s", err.Error())
	}
	if getResp.Count == 0 {
		return nil, pkgerrors.Errorf("Key doesn't exist")
	}
	return getResp.Kvs[0].Value, nil
}

// GetAll sub values from Etcd DB
func (e EtcdClient) GetAll(key string) ([][]byte, error) {
	if e.cli == nil {
		return nil, pkgerrors.Errorf("Etcd Client not initialized")
	}
	getResp, err := e.cli.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return nil, pkgerrors.Errorf("Error getting etcd entry: %s", err.Error())
	}
	result := make([][]byte, 0)
	for _, v := range getResp.Kvs {
		result = append(result, v.Value)
	}
	return result, nil
}

// Delete values from Etcd DB
func (e EtcdClient) Delete(key string) error {

	if e.cli == nil {
		return pkgerrors.Errorf("Etcd Client not initialized")
	}
	_, err := e.cli.Delete(context.Background(), key)
	if err != nil {
		return pkgerrors.Errorf("Delete failed etcd entry:%s", err.Error())
	}
	return nil
}

// Delete values by prefix from Etcd DB
func (e EtcdClient) DeletePrefix(keyPrefix string) error {

	if e.cli == nil {
		return pkgerrors.Errorf("Etcd Client not initialized")
	}
	_, err := e.cli.Delete(context.Background(), keyPrefix, clientv3.WithPrefix())
	if err != nil {
		return pkgerrors.Errorf("Delete prefix failed etcd entry:%s", err.Error())
	}
	return nil
}
