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
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
	pkgerrors "github.com/pkg/errors"
)

// Db interface used to talk a concrete Database connection
var Db ContextDb

// ContextDb is an interface for accessing the context database
type ContextDb interface {
	// Returns nil if db health is good
	HealthCheck() error
	// Puts Json Struct in db with key
	Put(key string, value interface{}) error
	// Delete k,v
	Delete(key string) error
	// Delete all keys in heirarchy
	DeleteAll(key string) error
	// Gets Json Struct from db
	Get(key string, value interface{}) error
	// Returns all keys with a prefix
	GetAllKeys(path string) ([]string, error)
}

// createContextDBClient creates the DB client
func createContextDBClient(dbType string) error {
	var err error
	switch dbType {
	case "etcd":
		c := EtcdConfig{
			Endpoint: config.GetConfiguration().EtcdIP,
			CertFile: config.GetConfiguration().EtcdCert,
			KeyFile:  config.GetConfiguration().EtcdKey,
			CAFile:   config.GetConfiguration().EtcdCAFile,
		}
		Db, err = NewEtcdClient(nil, c)
		if err != nil {
			pkgerrors.Wrap(err, "Etcd Client Initialization failed with error")
		}
	default:
		return pkgerrors.New(dbType + "DB not supported")
	}
	return err
}

// InitializeContextDatabase sets up the connection to the
// configured database to allow the application to talk to it.
func InitializeContextDatabase() error {
	// Only support Etcd for now
	err := createContextDBClient("etcd")
	if err != nil {
		return pkgerrors.Cause(err)
	}
	err = Db.HealthCheck()
	if err != nil {
		return pkgerrors.Cause(err)
	}
	return nil
}
