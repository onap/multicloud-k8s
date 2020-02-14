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
	"encoding/json"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// Microservice contains the parameters needed for Microservices
// It implements the interface for managing the Microservices
type Microservice struct {
	Name string `json:"name"`

	IpAddress string `json:"ip-address"`

	Port int64 `json:"port"`
}

// MicroserviceKey is the key structure that is used in the database
type MicroserviceKey struct {
	MicroserviceName string `json:"microservice-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (mk MicroserviceKey) String() string {
	out, err := json.Marshal(mk)
	if err != nil {
		return ""
	}

	return string(out)
}

// MicroserviceManager is an interface exposes the Microservice functionality
type MicroserviceManager interface {
	CreateMicroservice(ms Microservice) (Microservice, error)
	GetMicroservice(name string) (Microservice, error)
	DeleteMicroservice(name string) error
}

// MicroserviceClient implements the Manager
// It will also be used to maintain some localized state
type MicroserviceClient struct {
	storeName           string
	tagMeta, tagContent string
}

// NewMicroserviceClient returns an instance of the MicroserviceClient
// which implements the Manager
func NewMicroserviceClient() *MicroserviceClient {
	return &MicroserviceClient{
		tagMeta: "microservicemetadata",
	}
}

// CreateMicroservice a new collection based on the Microservice
func (v *MicroserviceClient) CreateMicroservice(m Microservice) (Microservice, error) {

	//Construct the composite key to select the entry
	key := MicroserviceKey{
		MicroserviceName: m.Name,
	}

	//Check if this Microservice already exists
	_, err := v.GetMicroservice(m.Name)
	if err == nil {
		return Microservice{}, pkgerrors.New("Microservice already exists")
	}

	err = db.DBconn.Create(m.Name, key, v.tagMeta, m)
	if err != nil {
		return Microservice{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return m, nil
}

// GetMicroservice returns the Microservice for corresponding name
func (v *MicroserviceClient) GetMicroservice(name string) (Microservice, error) {

	//Construct the composite key to select the entry
	key := MicroserviceKey{
		MicroserviceName: name,
	}
	value, err := db.DBconn.Read(name, key, v.tagMeta)
	if err != nil {
		return Microservice{}, pkgerrors.Wrap(err, "Get Microservice")
	}

	//value is a byte array
	if value != nil {
		microserv := Microservice{}
		err = db.DBconn.Unmarshal(value, &microserv)
		if err != nil {
			return Microservice{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return microserv, nil
	}

	return Microservice{}, pkgerrors.New("Error getting Microservice")
}

// DeleteMicroservice the  Microservice from database
func (v *MicroserviceClient) DeleteMicroservice(name string) error {

	//Construct the composite key to select the entry
	key := MicroserviceKey{
		MicroserviceName: name,
	}
	err := db.DBconn.Delete(name, key, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Microservice Entry;")
	}
	return nil
}
