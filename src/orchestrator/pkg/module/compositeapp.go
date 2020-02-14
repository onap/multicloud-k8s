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
 * See the License for the specific language governinog permissions and
 * limitations under the License.
 */

package module

import (
	"encoding/json"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// CompositeApp contains metadata and spec for CompositeApps
// It implements the interface for managing the composite apps
type CompositeApp struct {
	Metadata CompositeAppMeta `json:"metadata"`
	Spec     CompositeAppSpec `json:"spec"`
}

//CompositeAppMeta contains the parameters needed for CompositeApps
type CompositeAppMeta struct {
	CompositeAppName string `json:"name"`
	Description      string `json:"description"`
}

//CompositeAppSpec contains the Version of the CompositeApp
type CompositeAppSpec struct {
	Version string `json:"version"`
}

// CompositeAppKey is the key structure that is used in the database
type CompositeAppKey struct {
	CompositeAppName string `json:"name"`
	Version          string `json:"version"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (cK CompositeAppKey) String() string {
	out, err := json.Marshal(cK)
	if err != nil {
		return ""
	}
	return string(out)
}

// CompositeAppManager is an interface exposes the CompositeApp functionality
type CompositeAppManager interface {
	CreateCompositeApp(c CompositeApp) (CompositeApp, error)
	GetCompositeApp(name string, version string) (CompositeApp, error)
	DeleteCompositeApp(name string, version string) error
}

// CompositeAppClient implements the CompositeAppManager
// It will also be used to maintain some localized state
type CompositeAppClient struct {
	storeName           string
	tagMeta, tagContent string
}

// NewCompositeAppClient returns an instance of the CompositeAppClient
// which implements the CompositeAppManager
func NewCompositeAppClient() *CompositeAppClient {
	return &CompositeAppClient{
		storeName: "orchestrator",
		tagMeta:   "compositeAppmetadata",
	}
}

// CreateCompositeApp creates a new collection based on the CompositeApp
func (v *CompositeAppClient) CreateCompositeApp(c CompositeApp) (CompositeApp, error) {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: c.Metadata.CompositeAppName,
		Version:          c.Spec.Version,
	}

	//Check if this CompositeApp already exists
	_, err := v.GetCompositeApp(c.Metadata.CompositeAppName, c.Spec.Version)
	if err == nil {
		return CompositeApp{}, pkgerrors.New("CompositeApp already exists")
	}

	err = db.DBconn.Create(v.storeName, key, v.tagMeta, c)
	if err != nil {
		return CompositeApp{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// GetCompositeApp returns the CompositeApp for corresponding name
func (v *CompositeAppClient) GetCompositeApp(name string, version string) (CompositeApp, error) {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: name,
		Version:          version,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
	if err != nil {
		return CompositeApp{}, pkgerrors.Wrap(err, "Get composite application")
	}

	//value is a byte array
	if value != nil {
		compApp := CompositeApp{}
		err = db.DBconn.Unmarshal(value, &compApp)
		if err != nil {
			return CompositeApp{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return compApp, nil
	}

	return CompositeApp{}, pkgerrors.New("Error getting composite application")
}

// DeleteCompositeApp deletes the  CompositeApp from database
func (v *CompositeAppClient) DeleteCompositeApp(name string, version string) error {

	//Construct the composite key to select the entry
	key := CompositeAppKey{
		CompositeAppName: name,
		Version:          version,
	}
	err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete CompositeApp Entry;")
	}

	return nil
}
