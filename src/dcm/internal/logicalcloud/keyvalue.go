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

package logicalcloud

import (
    "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
    "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

    pkgerrors "github.com/pkg/errors"
)

// KeyValue contains the parameters needed for a key value
type KeyValue struct {
    MetaData            KVMetaDataList        `json:"metadata"`
    Specification       KVSpec                `json:"spec"`
}


// MetaData contains the parameters needed for metadata
type KVMetaDataList struct {
    KeyValueName        string  `json:"kv-pair-name"`
    Description         string  `json:"description"`
    UserData1           string  `json:"userData1"`
    UserData2           string  `json:"userData2"`
}

// Spec contains the parameters needed for spec
type KVSpec struct {
    KV              KVData      `json:"kv"`
}

// UserData contains the parameters needed for user
type KVData struct {
    Key1            string      `json:"key1"`
    Key2            string      `json:"key2"`
}

// KeyValueKey is the key structure that is used in the database
type KeyValueKey struct {
    Project             string      `json:"project"`
    LogicalCloudName    string      `json:"logical-cloud-name"`
    KeyValueName        string      `json:"kv-pair-name"`
}

// KeyValueManager is an interface that exposes the connection
// functionality
type KeyValueManager interface {
    Create(project, logicalCloud string, c KeyValue) (KeyValue, error)
    Get(project, logicalCloud, name string) (KeyValue, error)
    GetAll(project, logicalCloud string) ([]KeyValue, error)
    Delete(project, logicalCloud, name string) error
    Update(project, logicalCloud, name string, c KeyValue) (KeyValue, error)
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
func (v *KeyValueClient) Create(project, logicalCloud string, c KeyValue) (KeyValue, error) {

    //Construct key consisting of name
    key := KeyValueKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        KeyValueName:       c.MetaData.KeyValueName,
    }

    //Check if project exists
    _, err := module.NewProjectClient().GetProject(project)
    if err != nil {
        return KeyValue{}, pkgerrors.New("Unable to find the project")
    }
    //check if logical cloud exists
    _, err = NewLogicalCloudClient().Get(project, logicalCloud)
    if err != nil {
        return KeyValue{}, pkgerrors.New("Unable to find the logical cloud")
    }
    //Check if this Key Value already exists
    _, err = v.Get(project, logicalCloud, c.MetaData.KeyValueName)
    if err == nil {
        return KeyValue{}, pkgerrors.New("Key Value already exists")
    }

    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
    if err != nil {
        return KeyValue{}, pkgerrors.Wrap(err, "Creating DB Entry")
    }

    return c, nil
}

// Get returns Key Value for correspondin name
func (v *KeyValueClient) Get(project, logicalCloud, kvPairName string) (KeyValue, error) {

    //Construct the composite key to select the entry
    key := KeyValueKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        KeyValueName:       kvPairName,
    }
    value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
    if err != nil {
        return KeyValue{}, pkgerrors.Wrap(err, "Get Key Value")
    }

    //value is a byte array
    if value != nil {
        kv := KeyValue{}
        err = db.DBconn.Unmarshal(value[0], &kv)
        if err != nil {
            return KeyValue{}, pkgerrors.Wrap(err, "Unmarshaling value")
        }
        return kv, nil
    }

    return KeyValue{}, pkgerrors.New("Error getting Key Value")
}

// Get All lists all key value pairs
func (v *KeyValueClient) GetAll(project, logicalCloud string) ([]KeyValue, error) {

    //Construct the composite key to select the entry
    key := KeyValueKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        KeyValueName:       "",
    }
    var resp  []KeyValue
    values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
    if err != nil {
        return []KeyValue{}, pkgerrors.Wrap(err, "Get Key Value")
    }

    for _, value := range values {
        kv := KeyValue{}
        err = db.DBconn.Unmarshal(value, &kv)
        if err != nil {
            return []KeyValue{}, pkgerrors.Wrap(err, "Unmarshaling value")
        }
        resp = append(resp, kv)
    }

    return resp, nil
}

// Delete the Key Value entry from database
func (v *KeyValueClient) Delete(project, logicalCloud, kvPairName string) error {

    //Construct the composite key to select the entry
    key := KeyValueKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        KeyValueName:       kvPairName,
    }
        err := db.DBconn.Remove(v.storeName, key)
        if err != nil {
                return pkgerrors.Wrap(err, "Delete Key Value")
        }
        return nil
}

// Update an entry for the Key Value in the database
func (v *KeyValueClient) Update(project, logicalCloud, kvPairName string, c KeyValue) (KeyValue, error) {

    key := KeyValueKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        KeyValueName:       kvPairName,
    }
    //Check if KV pair URl name is the same name in json
    if c.MetaData.KeyValueName != kvPairName {
        return KeyValue{}, pkgerrors.New("Update Error - KV pair name mismatch")
    }
    //Check if this Key Value exists
    _, err := v.Get(project, logicalCloud, kvPairName)
    if err != nil {
        return KeyValue{}, pkgerrors.New("Update Error - Key Value Pair doesn't exist")
    }
    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
    if err != nil {
        return KeyValue{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}
