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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"k8splugin/internal/db"

	uuid "github.com/hashicorp/go-uuid"
	pkgerrors "github.com/pkg/errors"
)

// Definition contains the parameters needed for resource bundle (rb) definitions
// It implements the interface for managing the definitions
type Definition struct {
	UUID        string `json:"uuid,omitempty"`
	Name        string `json:"name"`
	ChartName   string `json:"chart-name"`
	Description string `json:"description"`
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
	res, err := db.DBconn.ReadAll(v.storeName, v.tagMeta)
	if err != nil || len(res) == 0 {
		return []Definition{}, pkgerrors.Wrap(err, "Listing Resource Bundle Definitions")
	}

	var results []Definition
	for key, value := range res {
		//value is a byte array
		if len(value) > 0 {
			def := Definition{}
			err = db.DBconn.Unmarshal(value, &def)
			if err != nil {
				log.Printf("[Definition] Error Unmarshaling value for: %s", key)
				continue
			}
			results = append(results, def)
		}
	}

	return results, nil
}

// Get returns the Resource Bundle Definition for corresponding ID
func (v *DefinitionClient) Get(id string) (Definition, error) {
	value, err := db.DBconn.Read(v.storeName, id, v.tagMeta)
	if err != nil {
		return Definition{}, pkgerrors.Wrap(err, "Get Resource Bundle definition")
	}

	//value is a byte array
	if value != nil {
		def := Definition{}
		err = db.DBconn.Unmarshal(value, &def)
		if err != nil {
			return Definition{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return def, nil
	}

	return Definition{}, pkgerrors.New("Error getting Resource Bundle Definition")
}

// Delete the Resource Bundle definition from database
func (v *DefinitionClient) Delete(id string) error {
	err := db.DBconn.Delete(v.storeName, id, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource Bundle Definition")
	}

	//Delete the content when the delete operation happens
	err = db.DBconn.Delete(v.storeName, id, v.tagContent)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource Bundle Definition Content")
	}

	return nil
}

// Upload the contents of resource bundle into database
func (v *DefinitionClient) Upload(id string, inp []byte) error {

	//Check if definition metadata exists
	def, err := v.Get(id)
	if err != nil {
		return pkgerrors.Errorf("Invalid Definition ID provided: %s", err.Error())
	}

	err = isTarGz(bytes.NewBuffer(inp))
	if err != nil {
		return pkgerrors.Errorf("Error in file format: %s", err.Error())
	}

	//Detect chart name from data if it was not provided originally
	if def.ChartName == "" {
		path, err := ExtractTarBall(bytes.NewBuffer(inp))
		if err != nil {
			return pkgerrors.Wrap(err, "Detecting chart name")
		}

		finfo, err := ioutil.ReadDir(path)
		if err != nil {
			return pkgerrors.Wrap(err, "Detecting chart name")
		}

		//Store the first directory with Chart.yaml found as the chart name
		for _, f := range finfo {
			if f.IsDir() {
				//Check if Chart.yaml exists
				if _, err = os.Stat(filepath.Join(path, f.Name(), "Chart.yaml")); err == nil {
					def.ChartName = f.Name()
					break
				}
			}
		}

		if def.ChartName == "" {
			return pkgerrors.New("Unable to detect chart name")
		}

		_, err = v.Create(def)
		if err != nil {
			return pkgerrors.Wrap(err, "Storing updated chart metadata")
		}
	}

	//Encode given byte stream to text for storage
	encodedStr := base64.StdEncoding.EncodeToString(inp)
	err = db.DBconn.Create(v.storeName, id, v.tagContent, encodedStr)
	if err != nil {
		return pkgerrors.Errorf("Error uploading data to db: %s", err.Error())
	}

	return nil
}

// Download the contents of the resource bundle definition from DB
// Returns a byte array of the contents which is used by the
// ExtractTarBall code to create the folder structure on disk
func (v *DefinitionClient) Download(id string) ([]byte, error) {

	//ignore the returned data here
	//Check if id is valid
	_, err := v.Get(id)
	if err != nil {
		return nil, pkgerrors.Errorf("Invalid Definition ID provided: %s", err.Error())
	}

	value, err := db.DBconn.Read(v.storeName, id, v.tagContent)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Resource Bundle definition content")
	}

	if value != nil {
		//Decode the string from base64
		out, err := base64.StdEncoding.DecodeString(string(value))
		if err != nil {
			return nil, pkgerrors.Wrap(err, "Decode base64 string")
		}

		if out != nil && len(out) != 0 {
			return out, nil
		}
	}
	return nil, pkgerrors.New("Error downloading Definition content")
}
