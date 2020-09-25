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

package module

import (
	pkgerrors "github.com/pkg/errors"
)

// KeyValue contains the parameters needed for a key value
type KeyValue struct {
	MetaData      KVMetaDataList `json:"metadata"`
	Specification KVSpec         `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type KVMetaDataList struct {
	KeyValueName string `json:"name"`
	Description  string `json:"description"`
	UserData1    string `json:"userData1"`
	UserData2    string `json:"userData2"`
}

// Spec contains the parameters needed for spec
type KVSpec struct {
	Kv []map[string]interface{} `json:"kv"`
}

// KeyValueKey is the key structure that is used in the database
type KeyValueKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logical-cloud-name"`
	KeyValueName     string `json:"kvname"`
}

// KeyValueManager is an interface that exposes the connection
// functionality
type KeyValueManager interface {
	CreateKVPair(project, logicalCloud string, c KeyValue) (KeyValue, error)
	GetKVPair(project, logicalCloud, name string) (KeyValue, error)
	GetAllKVPairs(project, logicalCloud string) ([]KeyValue, error)
	DeleteKVPair(project, logicalCloud, name string) error
	UpdateKVPair(project, logicalCloud, name string, c KeyValue) (KeyValue, error)
}

// KeyValueClient implements the KeyValueManager
// It will also be used to maintain some localized state
type KeyValueClient struct {
	storeName string
	tagMeta   string
	util      Utility
}

// KeyValueClient returns an instance of the KeyValueClient
// which implements the KeyValueManager
func NewKeyValueClient() *KeyValueClient {
	service := DBService{}
	return &KeyValueClient{
		storeName: "orchestrator",
		tagMeta:   "keyvalue",
		util:      service,
	}
}

// Create entry for the key value resource in the database
func (v *KeyValueClient) CreateKVPair(project, logicalCloud string, c KeyValue) (KeyValue, error) {

	//Construct key consisting of name
	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     c.MetaData.KeyValueName,
	}

	//Check if project exist
	err := v.util.CheckProject(project)
	if err != nil {
		return KeyValue{}, pkgerrors.New("Unable to find the project")
	}
	//check if logical cloud exists
	err = v.util.CheckLogicalCloud(project, logicalCloud)
	if err != nil {
		return KeyValue{}, pkgerrors.New("Unable to find the logical cloud")
	}
	//Check if this Key Value already exists
	_, err = v.GetKVPair(project, logicalCloud, c.MetaData.KeyValueName)
	if err == nil {
		return KeyValue{}, pkgerrors.New("Key Value already exists")
	}

	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return KeyValue{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns Key Value for correspondin name
func (v *KeyValueClient) GetKVPair(project, logicalCloud, kvPairName string) (KeyValue, error) {

	//Construct the composite key to select the entry
	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     kvPairName,
	}
	value, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return KeyValue{}, pkgerrors.Wrap(err, "Get Key Value")
	}

	//value is a byte array
	if value != nil {
		kv := KeyValue{}
		err = v.util.DBUnmarshal(value[0], &kv)
		if err != nil {
			return KeyValue{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		return kv, nil
	}

	return KeyValue{}, pkgerrors.New("Key Value does not exist")
}

// Get All lists all key value pairs
func (v *KeyValueClient) GetAllKVPairs(project, logicalCloud string) ([]KeyValue, error) {

	//Construct the composite key to select the entry
	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     "",
	}
	var resp []KeyValue
	values, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return []KeyValue{}, pkgerrors.Wrap(err, "Get Key Value")
	}

	for _, value := range values {
		kv := KeyValue{}
		err = v.util.DBUnmarshal(value, &kv)
		if err != nil {
			return []KeyValue{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		resp = append(resp, kv)
	}

	return resp, nil
}

// Delete the Key Value entry from database
func (v *KeyValueClient) DeleteKVPair(project, logicalCloud, kvPairName string) error {

	//Construct the composite key to select the entry
	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     kvPairName,
	}
	err := v.util.DBRemove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Key Value")
	}
	return nil
}

// Update an entry for the Key Value in the database
func (v *KeyValueClient) UpdateKVPair(project, logicalCloud, kvPairName string, c KeyValue) (KeyValue, error) {

	key := KeyValueKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		KeyValueName:     kvPairName,
	}
	//Check if KV pair URl name is the same name in json
	if c.MetaData.KeyValueName != kvPairName {
		return KeyValue{}, pkgerrors.New("Update Error - KV pair name mismatch")
	}
	//Check if this Key Value exists
	_, err := v.GetKVPair(project, logicalCloud, kvPairName)
	if err != nil {
		return KeyValue{}, pkgerrors.New("KV Pair does not exist")
	}
	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return KeyValue{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}
