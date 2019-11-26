/*
 * Copyright 2019 Intel Corporation, Inc
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

package app

import (
	"encoding/json"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/namegenerator"

	pkgerrors "github.com/pkg/errors"
)

// RegistryRequest contains the parameters needed for instantiation
// of profiles
type RegistryRequest struct {
	CloudOwner      string            `json:"cloud-owner"`
	CloudRegion     string            `json:"cloud-region"`
}

// RegistryResponse contains the response from instantiation
type RegistryResponse struct {
	ID        string                    `json:"id"`
	Request   RegistryRequest           `json:"request"`
}

// InstanceKey is used as the primary key in the db
type RegistryKey struct {
	ID string `json:"id"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk RegistryKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}

	return string(out)
}

// RegistryManager is an interface exposes the instantiation functionality
type RegistryManager interface {
	Create(i RegistryRequest) (RegistryResponse, error)
	Get(id string) (RegistryResponse, error)
	Delete(id string) error
}

// RegistryClient implements the RegistryManager interface
// It will also be used to maintain some localized state
type RegistryClient struct {
	storeName     string
	tagReg        string
}

// NewRegistryClient returns an instance of the RegistryClient
// which implements the RegistryManager
func NewRegistryClient() *RegistryClient {
	return &RegistryClient{
		storeName:     "registry",
		tagReg:  "vim",
	}
}

// Create an entry for the registry in the database
func (v *RegistryClient) Create(i RegistryRequest) (RegistryResponse, error) {

	// Name is required
	if i.CloudOwner == "" || i.CloudRegion == "" {
		return RegistryResponse{},
			pkgerrors.New("CloudOwner, CloudRegion are required to create a new registry")
	}

	id  := namegenerator.Generate()

	//Compose the return response
	resp := RegistryResponse{
		ID:        id,
		Request:   i,
	}

	key := RegistryKey{
		ID: id,
	}

	err := db.DBconn.Create(v.storeName, key, v.tagReg, resp)
	if err != nil {
		return RegistryResponse{}, pkgerrors.Wrap(err, "Creating Registry DB Entry")
	}

	return resp, nil
}

// Get returns the registry for corresponding ID
func (v *RegistryClient) Get(id string) (RegistryResponse, error) {
	key := RegistryKey{
		ID: id,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagReg)
	if err != nil {
		return RegistryResponse{}, pkgerrors.Wrap(err, "Get Registry")
	}

	//value is a byte array
	if value != nil {
		resp := RegistryResponse{}
		err = db.DBconn.Unmarshal(value, &resp)
		if err != nil {
			return RegistryResponse{}, pkgerrors.Wrap(err, "Unmarshaling Registry Value")
		}
		return resp, nil
	}

	return RegistryResponse{}, pkgerrors.New("Error getting Registry INFO")
}

// Delete the Registry from database
func (v *RegistryClient) Delete(id string) error {
	key := RegistryKey{
		ID: id,
	}
	err := db.DBconn.Delete(v.storeName, key, v.tagReg)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Registry")
	}

	return nil
}
