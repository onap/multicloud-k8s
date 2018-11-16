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

package resource

import (
	"k8splugin/db"
	"log"

	uuid "github.com/hashicorp/go-uuid"
	pkgerrors "github.com/pkg/errors"
)

// BundleDefinition contains the parameters needed for bundle definitions
// It implements the interface for managing the definitions
type BundleDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UUID        string `json:"uuid,omitempty"`
	ServiceType string `json:"service-type"`
}

// BundleDefInterface is an interface exposes the Definition functionality
type BundleDefInterface interface {
	Create(def BundleDefinition) (BundleDefinition, error)
	List() ([]BundleDefinition, error)
	Get(resID string) (BundleDefinition, error)
	Delete(resID string) error
}

// BundleDefClient implements the BundleDefInterface
// It will also be used to maintain some localized state
type BundleDefClient struct {
	keyPrefix string
}

// GetBundleDefClient Returns an instance of the BundleDefClient
// which implements the DefinitionInterface interface
// Uses resource/def prefix
func GetBundleDefClient() *BundleDefClient {
	return &BundleDefClient{
		keyPrefix: "resource/def/"}
}

// Create creates an entry for the resource in the database
func (v *BundleDefClient) Create(def BundleDefinition) (BundleDefinition, error) {
	// If UUID is empty, we will generate one
	if def.UUID == "" {
		def.UUID, _ = uuid.GenerateUUID()
	}
	key := v.keyPrefix + def.UUID

	serData, err := db.Serialize(v)
	if err != nil {
		return BundleDefinition{}, pkgerrors.Wrap(err, "Serialize Resource Bundle Definition")
	}

	err = db.DBconn.Create(key, serData)
	if err != nil {
		return BundleDefinition{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return def, nil
}

// List lists all resource entries in the database
func (v *BundleDefClient) List() ([]BundleDefinition, error) {
	strArray, err := db.DBconn.ReadAll(v.keyPrefix)
	if err != nil {
		return []BundleDefinition{}, pkgerrors.Wrap(err, "Listing Resource Bundle Definitions")
	}

	var retData []BundleDefinition

	for _, key := range strArray {
		value, err := db.DBconn.Read(key)
		if err != nil {
			log.Printf("Error Reading Key: %s", key)
			continue
		}
		if value != "" {
			def := BundleDefinition{}
			err = db.DeSerialize(value, &def)
			if err != nil {
				log.Printf("Error Deserializing Value: %s", value)
				continue
			}
			retData = append(retData, def)
		}
	}

	return retData, nil
}

// Get returns the Bundle Definition for corresponding ID
func (v *BundleDefClient) Get(id string) (BundleDefinition, error) {
	value, err := db.DBconn.Read(v.keyPrefix + id)
	if err != nil {
		return BundleDefinition{}, pkgerrors.Wrap(err, "Get Resource Bundle Definitions")
	}

	if value != "" {
		def := BundleDefinition{}
		err = db.DeSerialize(value, &def)
		if err != nil {
			return BundleDefinition{}, pkgerrors.Wrap(err, "Deserializing Value")
		}
		return def, nil
	}

	return BundleDefinition{}, pkgerrors.New("Error getting Resource Bundle Definition")
}

// Delete deletes the Bundle Definition from database
func (v *BundleDefClient) Delete(id string) error {
	err := db.DBconn.Delete(v.keyPrefix + id)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource Bundle Definitions")
	}

	return nil
}
