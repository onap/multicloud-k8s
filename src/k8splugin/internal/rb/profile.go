/*
 * Copyright 2018 Intel Corporation, Inc
 * Copyright © 2021 Samsung Electronics
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
	"log"
	"path/filepath"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"

	pkgerrors "github.com/pkg/errors"
)

// Profile contains the parameters needed for resource bundle (rb) profiles
// It implements the interface for managing the profiles
type Profile struct {
	RBName            string            `json:"rb-name"`
	RBVersion         string            `json:"rb-version"`
	ProfileName       string            `json:"profile-name"`
	ReleaseName       string            `json:"release-name"`
	Namespace         string            `json:"namespace"`
	KubernetesVersion string            `json:"kubernetes-version"`
	Labels            map[string]string `json:"labels"`
}

// ProfileManager is an interface exposes the resource bundle profile functionality
type ProfileManager interface {
	Create(def Profile) (Profile, error)
	Get(rbName, rbVersion, prName string) (Profile, error)
	List(rbName, rbVersion string) ([]Profile, error)
	Delete(rbName, rbVersion, prName string) error
	Upload(rbName, rbVersion, prName string, inp []byte) error
}

type ProfileKey struct {
	RBName      string `json:"rb-name"`
	RBVersion   string `json:"rb-version"`
	ProfileName string `json:"profile-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk ProfileKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}

	return string(out)
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
func NewProfileClient() *ProfileClient {
	return &ProfileClient{
		storeName:    "rbdef",
		tagMeta:      "profilemetadata",
		tagContent:   "profilecontent",
		manifestName: "manifest.yaml",
	}
}

// Create an entry for the resource bundle profile in the database
func (v *ProfileClient) Create(p Profile) (Profile, error) {

	// Name is required
	if p.ProfileName == "" {
		return Profile{}, pkgerrors.New("Name is required for Resource Bundle Profile")
	}

	//Check if profile already exists
	_, err := v.Get(p.RBName, p.RBVersion, p.ProfileName)
	if err == nil {
		return Profile{}, pkgerrors.New("Profile already exists for this Definition")
	}

	//Check if provided resource bundle information is valid
	_, err = NewDefinitionClient().Get(p.RBName, p.RBVersion)
	if err != nil {
		return Profile{}, pkgerrors.Errorf("Invalid Resource Bundle ID provided: %s", err.Error())
	}

	//If release-name is not provided, we store name instead
	if p.ReleaseName == "" {
		p.ReleaseName = p.ProfileName
	}

	key := ProfileKey{
		RBName:      p.RBName,
		RBVersion:   p.RBVersion,
		ProfileName: p.ProfileName,
	}

	err = db.DBconn.Create(v.storeName, key, v.tagMeta, p)
	if err != nil {
		return Profile{}, pkgerrors.Wrap(err, "Creating Profile DB Entry")
	}

	return p, nil
}

// Get returns the Resource Bundle Profile for corresponding ID
func (v *ProfileClient) Get(rbName, rbVersion, prName string) (Profile, error) {
	key := ProfileKey{
		RBName:      rbName,
		RBVersion:   rbVersion,
		ProfileName: prName,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
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

// List returns the Resource Bundle Profile for corresponding ID
func (v *ProfileClient) List(rbName, rbVersion string) ([]Profile, error) {

	//Get all profiles
	dbres, err := db.DBconn.ReadAll(v.storeName, v.tagMeta)
	if err != nil || len(dbres) == 0 {
		return []Profile{}, pkgerrors.Wrap(err, "No Profiles Found")
	}

	var results []Profile
	for key, value := range dbres {
		//value is a byte array
		if value != nil {
			pr := Profile{}
			err = db.DBconn.Unmarshal(value, &pr)
			if err != nil {
				log.Printf("[Profile] Error: %s Unmarshaling value for: %s", err.Error(), key)
				continue
			}
			if pr.RBName == rbName && pr.RBVersion == rbVersion {
				results = append(results, pr)
			}
		}
	}

	if len(results) == 0 {
		return results, pkgerrors.New("No Profiles Found for Definition and Version")
	}

	return results, nil
}

// Delete the Resource Bundle Profile from database
func (v *ProfileClient) Delete(rbName, rbVersion, prName string) error {
	key := ProfileKey{
		RBName:      rbName,
		RBVersion:   rbVersion,
		ProfileName: prName,
	}
	err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource Bundle Profile")
	}

	err = db.DBconn.Delete(v.storeName, key, v.tagContent)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource Bundle Profile Content")
	}

	return nil
}

// Upload the contents of resource bundle into database
func (v *ProfileClient) Upload(rbName, rbVersion, prName string, inp []byte) error {

	//ignore the returned data here.
	_, err := v.Get(rbName, rbVersion, prName)
	if err != nil {
		return pkgerrors.Errorf("Invalid Profile Name provided %s", err.Error())
	}

	err = isTarGz(bytes.NewBuffer(inp))
	if err != nil {
		return pkgerrors.Errorf("Error in file format %s", err.Error())
	}

	key := ProfileKey{
		RBName:      rbName,
		RBVersion:   rbVersion,
		ProfileName: prName,
	}
	//Encode given byte stream to text for storage
	encodedStr := base64.StdEncoding.EncodeToString(inp)
	err = db.DBconn.Create(v.storeName, key, v.tagContent, encodedStr)
	if err != nil {
		return pkgerrors.Errorf("Error uploading data to db %s", err.Error())
	}

	return nil
}

// Download the contents of the resource bundle profile from DB
// Returns a byte array of the contents which is used by the
// ExtractTarBall code to create the folder structure on disk
func (v *ProfileClient) Download(rbName, rbVersion, prName string) ([]byte, error) {

	//ignore the returned data here
	//Check if id is valid
	_, err := v.Get(rbName, rbVersion, prName)
	if err != nil {
		return nil, pkgerrors.Errorf("Invalid Profile Name provided: %s", err.Error())
	}

	key := ProfileKey{
		RBName:      rbName,
		RBVersion:   rbVersion,
		ProfileName: prName,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagContent)
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
//configuration overrides resides and final ReleaseName picked for instantiation
func (v *ProfileClient) Resolve(rbName string, rbVersion string,
	profileName string, values []string, overrideReleaseName string) ([]helm.KubernetesResourceTemplate, []helm.KubernetesResourceTemplate, []*helm.Hook, string, error) {

	var sortedTemplates []helm.KubernetesResourceTemplate
	var crdList []helm.KubernetesResourceTemplate
	var hookList []*helm.Hook
	var finalReleaseName string

	//Download and process the profile first
	//If everything seems okay, then download the definition
	prData, err := v.Download(rbName, rbVersion, profileName)
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Downloading Profile")
	}

	prPath, err := ExtractTarBall(bytes.NewBuffer(prData))
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Extracting Profile Content")
	}

	prYamlClient, err := ProcessProfileYaml(prPath, v.manifestName)
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Processing Profile Manifest")
	}

	definitionClient := NewDefinitionClient()

	definition, err := definitionClient.Get(rbName, rbVersion)
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Getting Definition Metadata")
	}

	defData, err := definitionClient.Download(rbName, rbVersion)
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Downloading Definition")
	}

	chartBasePath, err := ExtractTarBall(bytes.NewBuffer(defData))
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Extracting Definition Charts")
	}

	//Get the definition ID and download its contents
	profile, err := v.Get(rbName, rbVersion, profileName)
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Getting Profile")
	}

	//Copy the profile configresources to the chart locations
	//Corresponds to the following from the profile yaml
	// configresource:
	// - filepath: config.yaml
	//   chartpath: chart/config/resources/config.yaml
	err = prYamlClient.CopyConfigurationOverrides(chartBasePath)
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Copying configresources to chart")
	}

	if overrideReleaseName == "" {
		finalReleaseName = profile.ReleaseName
	} else {
		finalReleaseName = overrideReleaseName
	}

	helmClient := helm.NewTemplateClient(profile.KubernetesVersion,
		profile.Namespace,
		finalReleaseName)

	chartPath := filepath.Join(chartBasePath, definition.ChartName)
	sortedTemplates, crdList, hookList, err = helmClient.GenerateKubernetesArtifacts(chartPath,
		[]string{prYamlClient.GetValues()},
		values)
	if err != nil {
		return sortedTemplates, crdList, hookList, finalReleaseName, pkgerrors.Wrap(err, "Generate final k8s yaml")
	}

	return sortedTemplates, crdList, hookList, finalReleaseName, nil
}

// Returns an empty profile with the following contents
// Contains a manifest.yaml pointing to an override_values.yaml
// The override_values.yaml file is empty.
func (v *ProfileClient) getEmptyProfile() []byte {
	return []byte{
		0x1F, 0x8B, 0x08, 0x08, 0x25, 0x5D, 0xDC, 0x5C, 0x00, 0x03, 0x70,
		0x72, 0x6F, 0x66, 0x69, 0x6C, 0x65, 0x31, 0x2E, 0x74, 0x61, 0x72,
		0x00, 0xED, 0xD4, 0xD1, 0x0A, 0x82, 0x30, 0x14, 0xC6, 0xF1, 0x5D,
		0xEF, 0x29, 0x46, 0xF7, 0xC6, 0x96, 0x3A, 0xC1, 0x97, 0x89, 0x81,
		0x0B, 0xC4, 0xB4, 0x70, 0x26, 0xF8, 0xF6, 0xAD, 0xBC, 0xA8, 0x20,
		0xF0, 0xC6, 0x8A, 0xE0, 0xFF, 0xBB, 0x39, 0xB0, 0x33, 0xD8, 0x81,
		0xED, 0x5B, 0xEB, 0xBA, 0xFA, 0xE0, 0xC3, 0xB0, 0x9D, 0x5C, 0x7B,
		0x14, 0x9F, 0xA1, 0x23, 0x9B, 0x65, 0xB7, 0x6A, 0x8A, 0x5C, 0x3F,
		0xD7, 0x99, 0x2D, 0x84, 0x49, 0x33, 0x5B, 0xE8, 0x3C, 0xDF, 0x99,
		0x9D, 0xD0, 0x26, 0xD7, 0xA9, 0x15, 0x4A, 0x7F, 0x68, 0x9E, 0x17,
		0x97, 0x30, 0xB8, 0x5E, 0x29, 0xE1, 0xAA, 0x7D, 0xD3, 0x34, 0xAE,
		0xAD, 0x3B, 0xFF, 0x76, 0xDF, 0x52, 0xFF, 0x4F, 0x25, 0x49, 0x22,
		0x47, 0xDF, 0x87, 0xFA, 0xD4, 0x95, 0x6A, 0x34, 0x72, 0x98, 0xCE,
		0xBE, 0x94, 0x4A, 0x8D, 0xEE, 0x78, 0xF1, 0xA1, 0x54, 0x9B, 0x53,
		0xEC, 0xF6, 0x75, 0xE5, 0xF7, 0xF3, 0xCA, 0xFD, 0x9D, 0x6C, 0xE4,
		0xAF, 0xC7, 0xC6, 0x4A, 0xDE, 0x5D, 0xEF, 0xDA, 0x67, 0x2C, 0xE6,
		0x5F, 0x9B, 0x47, 0xFE, 0x75, 0x1A, 0xF3, 0x6F, 0xB3, 0xF8, 0x0D,
		0x90, 0xFF, 0x2F, 0x20, 0xC9, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xED, 0x0A, 0xB2, 0xAD,
		0x4D, 0xE5, 0x00, 0x28, 0x00, 0x00,
	}
}
