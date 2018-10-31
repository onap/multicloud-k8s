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

// VNFDefinition contains the parameters needed for VNFD creation and
type VNFDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UUID        string `json:"uuid,omitempty"`
	ServiceType string `json:"service-type"`
}

// createEntry adds a new entry for the UUID in the database
func (v *VNFDefinition) createEntry() error {
	key := "vnfd/" + v.UUID
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

// CreateVNFDefinition creates an entry for the VNF in the database
func CreateVNFDefinition(vnfd VNFDefinition) (VNFDefinition, error) {
	// If UUID is empty, we will generate one
	if vnfd.UUID == "" {
		vnfd.UUID, _ = uuid.GenerateUUID()
	}

	err := vnfd.createEntry()
	if err != nil {
		return VNFDefinition{}, pkgerrors.Wrap(err, "Creating VNF Definition")
	}

	return vnfd, nil
}

// ListVNFDefinitions lists all vnf entries in the database
func ListVNFDefinitions() ([]VNFDefinition, error) {

	strArray, err := db.DBconn.ReadAll("vnfd/")
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
func GetVNFDefinition(vnfID string) (VNFDefinition, error) {

	value, found, err := db.DBconn.ReadEntry("vnfd/" + vnfID)
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
func DeleteVNFDefinition(vnfID string) error {

	err := db.DBconn.DeleteEntry("vnfd/" + vnfID)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete VNF Definitions")
	}

	return nil
}
