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
var DBconn DatabaseConnection

// DatabaseConnection is an interface for accessing a database
type DatabaseConnection interface {
	InitializeDatabase() error
	CheckDatabase() error
	CreateEntry(string, string) error
	ReadEntry(string) (string, bool, error)
	DeleteEntry(string) error
	ReadAll(string) ([]string, error)
}

// CreateDBClient creates the DB client
var CreateDBClient = func(dbType string) error {
	switch dbType {
	case "consul":
		DBconn = &ConsulDB{}
		return nil
	default:
		return pkgerrors.New(dbType + "DB not supported")
	}
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
