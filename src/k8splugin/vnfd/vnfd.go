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

package vnfd

import (
	uuid "github.com/hashicorp/go-uuid"
	pkgerrors "github.com/pkg/errors"
	"k8splugin/db"
	"log"
)

// VNFDefinitionInterface is an interface exposes the VNFDefinition functionality
type VNFDefinitionInterface interface {
	CreateVNFDefinition(vnfd VNFDefinition) (VNFDefinition, error)
	ListVNFDefinitions() ([]VNFDefinition, error)
	GetVNFDefinition(vnfID string) (VNFDefinition, error)
	DeleteVNFDefinition(vnfID string) error
}

// VNFDefinition contains the parameters needed for VNF Definitions
// It implements the interface for managing the definitions
type VNFDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UUID        string `json:"uuid,omitempty"`
	ServiceType string `json:"service-type"`
}

// createEntry adds a new entry for the UUID in the database
func (v *VNFDefinition) createEntry(prefix string) error {
	key := prefix + v.UUID
	serData, err := db.Serialize(v)
	if err != nil {
		return pkgerrors.Wrap(err, "Serialize VNF Definition")
	}

	err = db.DBconn.CreateEntry(key, serData)
	if err == nil {
		return pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return nil
}

// VNFDefinitionImpl implements the VNFDefinitionInterface
// It will also be used to maintain some localized state
type VNFDefinitionImpl struct {
	keyPrefix string
}

// GetVNFDImplementation Returns an implementation of the VNFDefinitionImpl
// which implements the VNFDefinitionInterface interface
func GetVNFDImplementation() *VNFDefinitionImpl {
	return &VNFDefinitionImpl{keyPrefix: "vnfd/"}
}

// CreateVNFDefinition creates an entry for the VNF in the database
func (v *VNFDefinitionImpl) CreateVNFDefinition(vnfd VNFDefinition) (VNFDefinition, error) {
	// If UUID is empty, we will generate one
	if vnfd.UUID == "" {
		vnfd.UUID, _ = uuid.GenerateUUID()
	}

	err := vnfd.createEntry(v.keyPrefix)
	if err != nil {
		return VNFDefinition{}, pkgerrors.Wrap(err, "Creating VNF Definition")
	}

	return vnfd, nil
}

// ListVNFDefinitions lists all vnf entries in the database
func (v *VNFDefinitionImpl) ListVNFDefinitions() ([]VNFDefinition, error) {

	strArray, err := db.DBconn.ReadAll(v.keyPrefix)
	if err != nil {
		return []VNFDefinition{}, pkgerrors.Wrap(err, "Listing VNF Definitions")
	}

	var retData []VNFDefinition

	for _, key := range strArray {
		value, found, err := db.DBconn.ReadEntry(key)
		if err != nil {
			log.Printf("Error Reading Key: %s", key)
			continue
		}
		if found == true {
			vnfd := VNFDefinition{}
			err = db.DeSerialize(value, &vnfd)
			if err != nil {
				log.Printf("Error Deserializing Value: %s", value)
				continue
			}
			retData = append(retData, vnfd)
		}
	}

	return retData, nil
}

// GetVNFDefinition returns the VNF Definition for corresponding ID
func (v *VNFDefinitionImpl) GetVNFDefinition(vnfID string) (VNFDefinition, error) {

	value, found, err := db.DBconn.ReadEntry(v.keyPrefix + vnfID)
	if err != nil {
		return VNFDefinition{}, pkgerrors.Wrap(err, "Get VNF Definitions")
	}

	if found == true {
		vnfd := VNFDefinition{}
		err = db.DeSerialize(value, &vnfd)
		if err != nil {
			return VNFDefinition{}, pkgerrors.Wrap(err, "Deserializing Value")
		}
		return vnfd, nil

	}

	return VNFDefinition{}, pkgerrors.New("Error getting VNF Definition")
}

// DeleteVNFDefinition deletes the VNF Definition from database
func (v *VNFDefinitionImpl) DeleteVNFDefinition(vnfID string) error {

	err := db.DBconn.DeleteEntry(v.keyPrefix + vnfID)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete VNF Definitions")
	}

	return nil
}
