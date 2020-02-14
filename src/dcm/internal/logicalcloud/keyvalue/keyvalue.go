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

package keyvalue

import (
        "encoding/json"

        "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

        pkgerrors "github.com/pkg/errors"
)

// KeyValue contains the parameters needed for a key value
type KeyValue struct {
        MetaData            MetaDataList        `json:"metadata"`
        Specification       Spec                `json:"spec"`
}


// MetaData contains the parameters needed for metadata
type MetaDataList struct {
        KeyValueName        string  `json:"name"`
        Description         string  `json:"description"`
        UserData1           string  `json:"userData1"`
        UserData2           string  `json:"userData2"`
}

// Spec contains the parameters needed for spec
type Spec struct {
        KV              KVData      `json:"kv"`
}

// UserData contains the parameters needed for user
type KVData struct {
        Key1            string      `json:"key1"`
        Key2            string      `json:"key2"`
}

// KeyValueKey is the key structure that is used in the database
type KeyValueKey struct {
        KeyValueName    string      `json:"name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure
func (dk KeyValueKey) String() string {
        out, err := json.Marshal(dk)
        if err != nil {
                return ""
        }

        return string(out)
}

// KeyValueManager is an interface that exposes the connection
// functionality
type KeyValueManager interface {
        Create(c KeyValue) (KeyValue, error)
        Get(name string) (KeyValue, error)
        Delete(name string) error
        Update(name string, c KeyValue) (KeyValue, error)
}

// KeyValueClient implements the KeyValueManager
// It will also be used to maintain some localized state
type KeyValueClient struct {
        storeName   string
        tagMeta     string
}

// KeyValueClient returns an instance of the KeyValueClient
// which implements the KeyValueManager
func NewKeyValueClient() *KeyValueClient {
        return &KeyValueClient{
               storeName:   "keyvalue",
               tagMeta:     "metadata",
        }
}

// Create entry for the key value resource in the database
func (v *KeyValueClient) Create(c KeyValue) (KeyValue, error) {

        //Construct key consisting of name
        key := KeyValueKey{KeyValueName: c.MetaData.KeyValueName}

        //Check if this Key Value already exists
        _, err := v.Get(c.MetaData.KeyValueName)
        if err == nil {
                return KeyValue{}, pkgerrors.New("Key Value already exists")
        }

        err = db.DBconn.Create(v.storeName, key, v.tagMeta, c)
        if err != nil {
                return KeyValue{}, pkgerrors.Wrap(err, "Creating DB Entry")
        }

        return c, nil
}

// Get returns Key Value for correspondin name
func (v *KeyValueClient) Get(name string) (KeyValue, error) {

        //Construct the composite key to select the entry
        key := KeyValueKey{KeyValueName: name}
        value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
        if err != nil {
                return KeyValue{}, pkgerrors.Wrap(err, "Get Key Value")
        }

        //value is a byte array
        if value != nil {
                kv := KeyValue{}
                err = db.DBconn.Unmarshal(value, &kv)
                if err != nil {
                        return KeyValue{}, pkgerrors.Wrap(err, "Unmarshaling value")
                }
                return kv, nil
        }

        return KeyValue{}, pkgerrors.New("Error getting Key Value")
}

// Delete the Key Value entry from database
func (v *KeyValueClient) Delete(name string) error {
        //Construct the composite key to select the entry
        key := KeyValueKey{KeyValueName: name}
        err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
        if err != nil {
                return pkgerrors.Wrap(err, "Delete Key Value")
        }
        return nil
}

// Update an entry for the Key Value in the database
func (v *KeyValueClient) Update(name string, c KeyValue) (KeyValue, error) {

    key := KeyValueKey{
        KeyValueName: name,
    }
    //Check if this Key Value exists
    _, err := v.Get(name)
    if err != nil {
        return KeyValue{}, pkgerrors.New("Update Error - Key Value doesn't exist")
    }
    err = db.DBconn.Update(v.storeName, key, v.tagMeta, c)
    if err != nil {
        return KeyValue{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}
