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

package connectivity

import (
	"encoding/json"
	"k8splugin/internal/db"

	pkgerrors "github.com/pkg/errors"
)

// Connectivity contains the parameters needed for connectivity information for a Cloud region
type Connectivity struct {
	Name          string                 `json:"name"`
	CloudOwner    string                 `json:"cloud-owner"`
	CloudRegionID string                 `json:"cloud-region-id"`
	Kubeconfig    map[string]interface{} `json:"kubeconfig"`
	OtherConnectivityList        map[string]interface{} `json:"other-connectivity-list"`
}

//  ConnectivityKey is the key structure that is used in the database
type ConnectivityKey struct {
	ConnectivityName          string `json:"connectivity-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk ConnectivityKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}

	return string(out)
}

//  ConnectivityManager is an interface exposes the connectivity functionality
type ConnectivityManager interface {
	Create(c  Connectivity) (Connectivity, error)
	Get(name string) (Connectivity, error)
	Delete(name string) error
}

//  ConnectivityClient implements the  ConnectivityManager
// It will also be used to maintain some localized state
type  ConnectivityClient struct {
	storeName           string
	tagMeta             string
}

// New ConnectivityClient returns an instance of the  ConnectivityClient
// which implements the  ConnectivityManager
func NewConnectivityClient() * ConnectivityClient {
	return & ConnectivityClient{
		storeName:  "connectivity",
		tagMeta:    "metadata",
	}
}

// Create an entry for the connectivity resource in the database`
func (v * ConnectivityClient) Create(c Connectivity) (Connectivity, error) {

	//Construct composite key consisting of name
	key :=  ConnectivityKey{ConnectivityName: c.Name}

	//Check if this connectivity already exists
	_, err := v.Get( c.Name)
	if err == nil {
		return  Connectivity{}, pkgerrors.New(" Connectivity already exists")
	}

	err = db.DBconn.Create(v.storeName, key, v.tagMeta, c)
	if err != nil {
		return  Connectivity{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return  c, nil
}

// Get returns Connectivity for corresponding to name
func (v * ConnectivityClient) Get(name string) ( Connectivity, error) {

	//Construct the composite key to select the entry
	key := ConnectivityKey{ConnectivityName: name}
	value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
	if err != nil {
		return  Connectivity{}, pkgerrors.Wrap(err, "Get  connectivity")
	}

	//value is a byte array
	if value != nil {
		 c :=  Connectivity{}
		err = db.DBconn.Unmarshal(value, & c)
		if err != nil {
			return  Connectivity{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return c, nil
	}

	return  Connectivity{}, pkgerrors.New("Error getting   Connectivity")
}

// Delete the  connectivity from database
func (v * ConnectivityClient) Delete(name string) error {

	//Construct the composite key to select the entry
	key :=  ConnectivityKey{ConnectivityName: name}
	err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete   Connectivity")
	}
	return nil
}


