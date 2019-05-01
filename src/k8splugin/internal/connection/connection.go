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

package connection

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"k8splugin/internal/db"

	pkgerrors "github.com/pkg/errors"
)

// Connection contains the parameters needed for Connection information for a Cloud region
type Connection struct {
	CloudRegion           string                 `json:"cloud-region"`
	CloudOwner            string                 `json:"cloud-owner"`
	Kubeconfig            string                 `json:"kubeconfig"`
	OtherConnectivityList map[string]interface{} `json:"other-connectivity-list"`
}

// ConnectionKey is the key structure that is used in the database
type ConnectionKey struct {
	CloudRegion string `json:"cloud-region"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk ConnectionKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}

	return string(out)
}

// ConnectionManager is an interface exposes the Connection functionality
type ConnectionManager interface {
	Create(c Connection) (Connection, error)
	Get(name string) (Connection, error)
	Delete(name string) error
}

// ConnectionClient implements the  ConnectionManager
// It will also be used to maintain some localized state
type ConnectionClient struct {
	storeName string
	tagMeta   string
}

// New ConnectionClient returns an instance of the  ConnectionClient
// which implements the  ConnectionManager
func NewConnectionClient() *ConnectionClient {
	return &ConnectionClient{
		storeName: "connection",
		tagMeta:   "metadata",
	}
}

// Create an entry for the Connection resource in the database`
func (v *ConnectionClient) Create(c Connection) (Connection, error) {

	//Construct composite key consisting of name
	key := ConnectionKey{CloudRegion: c.CloudRegion}

	//Check if this Connection already exists
	_, err := v.Get(c.CloudRegion)
	if err == nil {
		return Connection{}, pkgerrors.New("Connection already exists")
	}

	err = db.DBconn.Create(v.storeName, key, v.tagMeta, c)
	if err != nil {
		return Connection{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns Connection for corresponding to name
func (v *ConnectionClient) Get(name string) (Connection, error) {

	//Construct the composite key to select the entry
	key := ConnectionKey{CloudRegion: name}
	value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
	if err != nil {
		return Connection{}, pkgerrors.Wrap(err, "Get Connection")
	}

	//value is a byte array
	if value != nil {
		c := Connection{}
		err = db.DBconn.Unmarshal(value, &c)
		if err != nil {
			return Connection{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return c, nil
	}

	return Connection{}, pkgerrors.New("Error getting Connection")
}

// Delete the Connection from database
func (v *ConnectionClient) Delete(name string) error {

	//Construct the composite key to select the entry
	key := ConnectionKey{CloudRegion: name}
	err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Connection")
	}
	return nil
}

// Download the connection information onto a kubeconfig file
// The file is named after the name of the connection and will
// be placed in the provided parent directory
func (v *ConnectionClient) Download(name string, parentdir string) (string, error) {

	conn, err := v.Get(name)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Getting Connection info")
	}

	//Decode the kubeconfig from base64 to string
	kubeContent, err := base64.StdEncoding.DecodeString(conn.Kubeconfig)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Converting from base64")
	}

	target := filepath.Join(parentdir, conn.CloudRegion)
	err = ioutil.WriteFile(target, kubeContent, 0644)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Writing kubeconfig to file")
	}

	return target, nil
}
