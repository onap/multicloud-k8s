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

package project

import (
	"encoding/json"

	"github.com/onap/multicloud-k8s/src/orchestrator/internal/db"

	pkgerrors "github.com/pkg/errors"
)

// Project contains the parameters needed for Projects
// It implements the interface for managing the Projects
type Project struct {
	ProjectName string `json:"project-name"`
	Description string `json:"description"`
}

// ProjectKey is the key structure that is used in the database
type ProjectKey struct {
	ProjectName string `json:"rb-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (pk ProjectKey) String() string {
	out, err := json.Marshal(pk)
	if err != nil {
		return ""
	}

	return string(out)
}

// ProjectManager is an interface exposes the Project functionality
type ProjectManager interface {
	Create(pr Project) (Project, error)
	Get(name string) (Project, error)
	Delete(name string) error
}

// ProjectClient implements the ProjectManager
// It will also be used to maintain some localized state
type ProjectClient struct {
	storeName           string
	tagMeta, tagContent string
}

// NewProjectClient returns an instance of the ProjectClient
// which implements the ProjectManager
func NewProjectClient() *ProjectClient {
	return &ProjectClient{
		tagMeta: "projectmetadata",
	}
}

// Create a new collection based on the project
func (v *ProjectClient) Create(p Project) (Project, error) {

	//Construct the composite key to select the entry
	key := ProjectKey{
		ProjectName: p.ProjectName,
	}

	//Check if this Project already exists
	_, err := v.Get(p.ProjectName)
	if err == nil {
		return Project{}, pkgerrors.New("Project already exists")
	}

	err = db.DBconn.Create(p.ProjectName, key, v.tagMeta, p)
	if err != nil {
		return Project{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// Get returns the Project for corresponding name
func (v *ProjectClient) Get(name string) (Project, error) {

	//Construct the composite key to select the entry
	key := ProjectKey{
		ProjectName: name,
	}
	value, err := db.DBconn.Read(name, key, v.tagMeta)
	if err != nil {
		return Project{}, pkgerrors.Wrap(err, "Get Project")
	}

	//value is a byte array
	if value != nil {
		proj := Project{}
		err = db.DBconn.Unmarshal(value, &proj)
		if err != nil {
			return Project{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return proj, nil
	}

	return Project{}, pkgerrors.New("Error getting Project")
}

// Delete the  Project from database
func (v *ProjectClient) Delete(name string) error {

	//Construct the composite key to select the entry
	key := ProjectKey{
		ProjectName: name,
	}
	err := db.DBconn.Delete(name, key, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Project Entry;")
	}

	//TODO: Delete the collection when the project is deleted
	return nil
}
