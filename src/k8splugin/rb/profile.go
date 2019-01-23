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
	"k8splugin/db"
	"log"

	uuid "github.com/hashicorp/go-uuid"
	pkgerrors "github.com/pkg/errors"
)

// Profile contains the parameters needed for resource bundle (rb) profiles
// It implements the interface for managing the profiles
type Profile struct {
	UUID              string `json:"uuid,omitempty"`
	RBDID             string `json:"rbdid"`
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	KubernetesVersion string `json:"kubernetesversion"`
}

// ProfileManager is an interface exposes the resource bundle profile functionality
type ProfileManager interface {
	Create(def Profile) (Profile, error)
	List() ([]Profile, error)
	Get(resID string) (Profile, error)
	Help() map[string]string
	Delete(resID string) error
	Upload(resID string, inp []byte) error
}

// ProfileClient implements the ProfileManager
// It will also be used to maintain some localized state
type ProfileClient struct {
	storeName           string
	tagMeta, tagContent string
	manifestName        string
}

// NewProfileClient returns an instance of the ProfileClient
// which implements the ProfileManager
// Uses rb/def prefix
func NewProfileClient() *ProfileClient {
	return &ProfileClient{
		storeName:    "rbprofile",
		tagMeta:      "metadata",
		tagContent:   "content",
		manifestName: "manifest.yaml",
	}
}

// Help returns some information on how to create the content
// for the profile in the form of html formatted page
func (v *ProfileClient) Help() map[string]string {
	ret := make(map[string]string)

	return ret
}

// Create an entry for the resource bundle profile in the database
func (v *ProfileClient) Create(p Profile) (Profile, error) {

	//Check if provided RBID is a valid resource bundle
	_, err := NewDefinitionClient().Get(p.RBDID)
	if err != nil {
		return Profile{}, pkgerrors.Errorf("Invalid Resource Bundle ID provided: %s", err.Error())
	}

	// Name is required
	if p.Name == "" {
		return Profile{}, pkgerrors.New("Name is required for Resource Bundle Profile")
	}

	// If UUID is empty, we will generate one
	if p.UUID == "" {
		p.UUID, _ = uuid.GenerateUUID()
	}
	key := p.UUID

	err = db.DBconn.Create(v.storeName, key, v.tagMeta, p)
	if err != nil {
		return Profile{}, pkgerrors.Wrap(err, "Creating Profile DB Entry")
	}

	return p, nil
}

// List all resource entries in the database
func (v *ProfileClient) List() ([]Profile, error) {
	res, err := db.DBconn.ReadAll(v.storeName, v.tagMeta)
	if err != nil || len(res) == 0 {
		return []Profile{}, pkgerrors.Wrap(err, "Listing Resource Bundle Profiles")
	}

	var retData []Profile

	for key, value := range res {
		//value is a byte array
		if len(value) > 0 {
			pr := Profile{}
			err = db.DBconn.Unmarshal(value, &pr)
			if err != nil {
				log.Printf("[Profile] Error Unmarshaling value for: %s", key)
				continue
			}
			retData = append(retData, pr)
		}
	}

	return retData, nil
}

// Get returns the Resource Bundle Profile for corresponding ID
func (v *ProfileClient) Get(id string) (Profile, error) {
	value, err := db.DBconn.Read(v.storeName, id, v.tagMeta)
	if err != nil {
		return Profile{}, pkgerrors.Wrap(err, "Get Resource Bundle Profile")
	}

	//value is a byte array
	if value != nil {
		pr := Profile{}
		err = db.DBconn.Unmarshal(value, &pr)
		if err != nil {
			return Profile{}, pkgerrors.Wrap(err, "Unmarshaling Profile Value")
		}
		return pr, nil
	}

	return Profile{}, pkgerrors.New("Error getting Resource Bundle Profile")
}

// Delete the Resource Bundle Profile from database
func (v *ProfileClient) Delete(id string) error {
	err := db.DBconn.Delete(v.storeName, id, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource Bundle Profile")
	}

	err = db.DBconn.Delete(v.storeName, id, v.tagContent)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource Bundle Profile Content")
	}

	return nil
}

// Upload the contents of resource bundle into database
func (v *ProfileClient) Upload(id string, inp []byte) error {

	//ignore the returned data here.
	_, err := v.Get(id)
	if err != nil {
		return pkgerrors.Errorf("Invalid Profile ID provided %s", err.Error())
	}

	err = isTarGz(bytes.NewBuffer(inp))
	if err != nil {
		return pkgerrors.Errorf("Error in file format %s", err.Error())
	}

	//Encode given byte stream to text for storage
	encodedStr := base64.StdEncoding.EncodeToString(inp)
	err = db.DBconn.Create(v.storeName, id, v.tagContent, encodedStr)
	if err != nil {
		return pkgerrors.Errorf("Error uploading data to db %s", err.Error())
	}

	return nil
}
