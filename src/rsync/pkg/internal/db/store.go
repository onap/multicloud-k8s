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
	"reflect"

	pkgerrors "github.com/pkg/errors"
)

// DBconn interface used to talk a concrete Database connection
var DBconn Store

// Key is an interface that will be implemented by anypackage
// that wants to use the Store interface. This allows various
// db backends and key types.
type Key interface {
	String() string
}

// Store is an interface for accessing a database
type Store interface {
	// Returns nil if db health is good
	HealthCheck() error

	// Unmarshal implements any unmarshaling needed for the database
	Unmarshal(inp []byte, out interface{}) error

	// Creates a new master table with key and links data with tag and
	// creates a pointer to the newly added data in the master table
	Create(table string, key Key, tag string, data interface{}) error

	// Reads data for a particular key with specific tag.
	Read(table string, key Key, tag string) ([]byte, error)

	// Update data for particular key with specific tag
	Update(table string, key Key, tag string, data interface{}) error

	// Deletes a specific tag data for key.
	// TODO: If tag is empty, it will delete all tags under key.
	Delete(table string, key Key, tag string) error

	// Reads all master tables and data from the specified tag in table
	ReadAll(table string, tag string) (map[string][]byte, error)
}

// CreateDBClient creates the DB client
func CreateDBClient(dbType string) error {
	var err error
	switch dbType {
	case "mongo":
		// create a mongodb database with k8splugin as the name
		DBconn, err = NewMongoStore("k8splugin", nil)
	case "consul":
		// create a consul kv store
		DBconn, err = NewConsulStore(nil)
	default:
		return pkgerrors.New(dbType + "DB not supported")
	}
	return err
}

// Serialize converts given data into a JSON string
func Serialize(v interface{}) (string, error) {
	out, err := json.Marshal(v)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error serializing "+reflect.TypeOf(v).String())
	}
	return string(out), nil
}

// DeSerialize converts string to a json object specified by type
func DeSerialize(str string, v interface{}) error {
	err := json.Unmarshal([]byte(str), &v)
	if err != nil {
		return pkgerrors.Wrap(err, "Error deSerializing "+str)
	}
	return nil
}
