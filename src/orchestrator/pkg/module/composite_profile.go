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

// CompositeProfile contains the parameters needed for CompositeProfiles
// It implements the interface for managing the CompositeProfiles
type CompositeProfile struct {
	Metadata CompositeProfileMetadata `json:"metadata"`
}

// CompositeProfileMetadata contains the metadata for CompositeProfiles
type CompositeProfileMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// CompositeProfileKey is the key structure that is used in the database
type CompositeProfileKey struct {
	Name         string `json:"compositeprofile"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeapp"`
	Version      string `json:"compositeappversion"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (cpk CompositeProfileKey) String() string {
	out, err := json.Marshal(cpk)
	if err != nil {
		return ""
	}

	return string(out)
}

// CompositeProfileManager exposes the CompositeProfile functionality
type CompositeProfileManager interface {
	CreateCompositeProfile(cpf CompositeProfile, p string, ca string,
		v string) (CompositeProfile, error)
	GetCompositeProfile(compositeProfileName string, projectName string,
		compositeAppName string, version string) (CompositeProfile, error)
	GetCompositeProfiles(projectName string, compositeAppName string,
		version string) ([]CompositeProfile, error)
	DeleteCompositeProfile(compositeProfileName string, projectName string,
		compositeAppName string, version string) error
}

// CompositeProfileClient implements the Manager
// It will also be used to maintain some localized state
type CompositeProfileClient struct {
	storeName string
	tagMeta   string
}

// NewCompositeProfileClient returns an instance of the CompositeProfileClient
// which implements the Manager
func NewCompositeProfileClient() *CompositeProfileClient {
	return &CompositeProfileClient{
		storeName: "orchestrator",
		tagMeta:   "compositeprofilemetadata",
	}
}

// CreateCompositeProfile creates an entry for CompositeProfile in the database. Other Input parameters for it - projectName, compositeAppName, version
func (c *CompositeProfileClient) CreateCompositeProfile(cpf CompositeProfile, p string, ca string,
	v string) (CompositeProfile, error) {

	res, err := c.GetCompositeProfile(cpf.Metadata.Name, p, ca, v)
	if res != (CompositeProfile{}) {
		return CompositeProfile{}, pkgerrors.New("CompositeProfile already exists")
	}

	//Check if project exists
	_, err = NewProjectClient().GetProject(p)
	if err != nil {
		return CompositeProfile{}, pkgerrors.New("Unable to find the project")
	}

	// check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return CompositeProfile{}, pkgerrors.New("Unable to find the composite-app")
	}

	cProfkey := CompositeProfileKey{
		Name:         cpf.Metadata.Name,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	err = db.DBconn.Insert(c.storeName, cProfkey, nil, c.tagMeta, cpf)
	if err != nil {
		return CompositeProfile{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return cpf, nil
}

// GetCompositeProfile shall take arguments - name of the composite profile, name of the project, name of the composite app and version of the composite app. It shall return the CompositeProfile if its present.
func (c *CompositeProfileClient) GetCompositeProfile(cpf string, p string, ca string, v string) (CompositeProfile, error) {
	key := CompositeProfileKey{
		Name:         cpf,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	result, err := db.DBconn.Find(c.storeName, key, c.tagMeta)
	if err != nil {
		return CompositeProfile{}, pkgerrors.Wrap(err, "Get Composite Profile error")
	}

	if result != nil {
		cProf := CompositeProfile{}
		err = db.DBconn.Unmarshal(result[0], &cProf)
		if err != nil {
			return CompositeProfile{}, pkgerrors.Wrap(err, "Unmarshalling CompositeProfile")
		}
		return cProf, nil
	}

	return CompositeProfile{}, pkgerrors.New("Error getting CompositeProfile")
}

// GetCompositeProfiles shall take arguments - name of the project, name of the composite profile and version of the composite app. It shall return an array of CompositeProfile.
func (c *CompositeProfileClient) GetCompositeProfiles(p string, ca string, v string) ([]CompositeProfile, error) {
	key := CompositeProfileKey{
		Name:         "",
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	values, err := db.DBconn.Find(c.storeName, key, c.tagMeta)
	if err != nil {
		return []CompositeProfile{}, pkgerrors.Wrap(err, "Get Composite Profiles error")
	}

	var resp []CompositeProfile

	for _, value := range values {
		cp := CompositeProfile{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []CompositeProfile{}, pkgerrors.Wrap(err, "Get Composite Profiles unmarshalling error")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteCompositeProfile deletes the compsiteApp profile from the database
func (c *CompositeProfileClient) DeleteCompositeProfile(cpf string, p string, ca string, v string) error {
	key := CompositeProfileKey{
		Name:         cpf,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	err := db.DBconn.Remove(c.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete CompositeProfile entry;")
	}
	return nil
}
