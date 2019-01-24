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
	"k8splugin/internal/db"
	"log"
	"path/filepath"

	uuid "github.com/hashicorp/go-uuid"
	pkgerrors "github.com/pkg/errors"

	"k8splugin/internal/helm"
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

// Download the contents of the resource bundle profile from DB
// Returns a byte array of the contents which is used by the
// ExtractTarBall code to create the folder structure on disk
func (v *ProfileClient) Download(id string) ([]byte, error) {

	//ignore the returned data here
	//Check if id is valid
	_, err := v.Get(id)
	if err != nil {
		return nil, pkgerrors.Errorf("Invalid Profile ID provided: %s", err.Error())
	}

	value, err := db.DBconn.Read(v.storeName, id, v.tagContent)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Get Resource Bundle Profile content")
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
	return nil, pkgerrors.New("Error downloading Profile content")
}

//Resolve returns the path where the helm chart merged with
//configuration overrides resides.
func (v *ProfileClient) Resolve(id string, values []string) (map[string][]string, error) {

	var retMap map[string][]string

	//Download and process the profile first
	//If everything seems okay, then download the definition
	prData, err := v.Download(id)
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Downloading Profile")
	}

	prPath, err := ExtractTarBall(bytes.NewBuffer(prData))
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Extracting Profile Content")
	}

	prYamlClient, err := ProcessProfileYaml(prPath, v.manifestName)
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Processing Profile Manifest")
	}

	//Get the definition ID and download its contents
	profile, err := v.Get(id)
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Getting Profile")
	}

	definitionClient := NewDefinitionClient()

	definition, err := definitionClient.Get(profile.RBDID)
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Getting Definition Metadata")
	}

	defData, err := definitionClient.Download(profile.RBDID)
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Downloading Definition")
	}

	chartBasePath, err := ExtractTarBall(bytes.NewBuffer(defData))
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Extracting Definition Charts")
	}

	//Copy the profile configresources to the chart locations
	//Corresponds to the following from the profile yaml
	// configresource:
	// - filepath: config.yaml
	//   chartpath: chart/config/resources/config.yaml
	err = prYamlClient.CopyConfigurationOverrides(chartBasePath)
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Copying configresources to chart")
	}

	helmClient := helm.NewTemplateClient(profile.KubernetesVersion,
		profile.Namespace,
		profile.Name)

	chartPath := filepath.Join(chartBasePath, definition.ChartName)
	retMap, err = helmClient.GenerateKubernetesArtifacts(chartPath,
		[]string{prYamlClient.GetValues()},
		values)
	if err != nil {
		return retMap, pkgerrors.Wrap(err, "Generate final k8s yaml")
	}

	return retMap, nil
}
