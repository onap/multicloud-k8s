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

package rb

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"k8splugin/db"
	"log"

	uuid "github.com/hashicorp/go-uuid"
	pkgerrors "github.com/pkg/errors"
)

// Definition contains the parameters needed for resource bundle (rb) definitions
// It implements the interface for managing the definitions
type Definition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UUID        string `json:"uuid,omitempty"`
	ServiceType string `json:"service-type"`
}

// DefinitionManager is an interface exposes the resource bundle definition functionality
type DefinitionManager interface {
	Create(def Definition) (Definition, error)
	List() ([]Definition, error)
	Get(resID string) (Definition, error)
	Delete(resID string) error
	Upload(resID string, inp []byte) error
}

// DefinitionClient implements the DefinitionManager
// It will also be used to maintain some localized state
type DefinitionClient struct {
	storeName           string
	tagMeta, tagContent string
}

// NewDefinitionClient returns an instance of the DefinitionClient
// which implements the DefinitionManager
// Uses rbdef collection in underlying db
func NewDefinitionClient() *DefinitionClient {
	return &DefinitionClient{
		storeName:  "rbdef",
		tagMeta:    "metadata",
		tagContent: "content",
	}
}

// Create an entry for the resource in the database
func (v *DefinitionClient) Create(def Definition) (Definition, error) {
	// If UUID is empty, we will generate one
	if def.UUID == "" {
		def.UUID, _ = uuid.GenerateUUID()
	}
	key := def.UUID

	err := db.DBconn.Create(v.storeName, key, v.tagMeta, def)
	if err != nil {
		return Definition{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return def, nil
}

// List all resource entries in the database
func (v *DefinitionClient) List() ([]Definition, error) {
	byteSlice, err := db.DBconn.ReadAll(v.storeName, v.tagMeta)
	if err != nil {
		return []Definition{}, pkgerrors.Wrap(err, "Listing Resource Bundle Definitions")
	}

	var retData []Definition

	for _, key := range byteSlice {
		if key != nil {
			def := Definition{}
			err = json.Unmarshal(key, &def)
			if err != nil {
				log.Printf("Error Unmarshaling data: %s", key)
				continue
			}
			retData = append(retData, def)
		}
	}

	return retData, nil
}

// Get returns the Resource Bundle Definition for corresponding ID
func (v *DefinitionClient) Get(id string) (Definition, error) {
	value, err := db.DBconn.Read(v.storeName, id, v.tagMeta)
	if err != nil {
		return Definition{}, pkgerrors.Wrap(err, "Get Resource Bundle definition")
	}

	if value != nil {
		def := Definition{}
		err = json.Unmarshal(value, &def)
		if err != nil {
			return Definition{}, pkgerrors.Wrap(err, "Deserializing Value")
		}
		return def, nil
	}

	return Definition{}, pkgerrors.New("Error getting Resource Bundle Definition")
}

// Delete the Resource Bundle definition from database
func (v *DefinitionClient) Delete(id string) error {
	err := db.DBconn.Delete(v.storeName, id, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource Bundle Definitions")
	}

	return nil
}

// Upload the contents of resource bundle into database
func (v *DefinitionClient) Upload(id string, inp []byte) error {

	err := isTarGz(bytes.NewBuffer(inp))
	if err != nil {
		return pkgerrors.Errorf("Error in file format %s", err.Error())
	}

	//Encode given byte stream to text for storage
	encodedStr := base64.StdEncoding.EncodeToString(inp)
	err = db.DBconn.Create(v.storeName, id, encodedStr, v.tagContent)
	if err != nil {
		return pkgerrors.Errorf("Error uploading data to db %s", err.Error())
	}

	return nil
}
