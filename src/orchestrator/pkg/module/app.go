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

// App contains metadata for Apps
type App struct {
	Metadata AppMetaData `json:"metadata"`
}

//AppMetaData contains the parameters needed for Apps
type AppMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
	File        string `json:"file"`
}

// AppKey is the key structure that is used in the database
type AppKey struct {
	AppName             string `json:"appname"`
	Project             string `json:"project"`
	CompositeAppName    string `json:"compositeappname"`
	CompositeAppVersion string `json:"compositeappversion"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (aK AppKey) String() string {
	out, err := json.Marshal(aK)
	if err != nil {
		return ""
	}
	return string(out)
}

// AppManager is an interface exposes the App functionality
type AppManager interface {
	CreateApp(c App, p string, cN string, cV string) (App, error)
	GetApp(name string, p string, cN string, cV string) (App, error)
	DeleteApp(name string, p string, cN string, cV string) error
}

// AppClient implements the AppManager
// It will also be used to maintain some localized state
type AppClient struct {
	storeName           string
	tagMeta, tagContent string
}

// NewAppClient returns an instance of the AppClient
// which implements the AppManager
func NewAppClient() *AppClient {
	return &AppClient{
		storeName:  "orchestrator",
		tagMeta:    "app",
		tagContent: "file",
	}
}

// CreateApp creates a new collection based on the App
func (v *AppClient) CreateApp(a App, p string, cN string, cV string) (App, error) {

	//Construct the composite key to select the entry
	key := AppKey{
		AppName:             a.Metadata.Name,
		Project:             p,
		CompositeAppName:    cN,
		CompositeAppVersion: cV,
	}

	//Check if this App already exists
	_, err := v.GetApp(a.Metadata.Name, p, cN, cV)
	if err == nil {
		return App{}, pkgerrors.New("App already exists")
	}

	//Check if Project exists
	_, err = NewProjectClient().GetProject(p)
	if err != nil {
		return App{}, pkgerrors.New("Unable to find the project")
	}

	//check if CompositeApp with version exists
	_, err = NewCompositeAppClient().GetCompositeApp(cN, cV, p)
	if err != nil {
		return App{}, pkgerrors.New("Unable to find the composite app with version")
	}

	err = db.DBconn.Create(v.storeName, key, v.tagMeta, a)
	if err != nil {
		return App{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return a, nil
}

// GetApp returns the App for corresponding name
func (v *AppClient) GetApp(name string, p string, cN string, cV string) (App, error) {

	//Construct the composite key to select the entry
	key := AppKey{
		AppName:             name,
		Project:             p,
		CompositeAppName:    cN,
		CompositeAppVersion: cV,
	}
	value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
	if err != nil {
		return App{}, pkgerrors.Wrap(err, "Get app")
	}

	//value is a byte array
	if value != nil {
		app := App{}
		err = db.DBconn.Unmarshal(value, &app)
		if err != nil {
			return App{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return app, nil
	}

	return App{}, pkgerrors.New("Error getting app")
}

// DeleteApp deletes the  App from database
func (v *AppClient) DeleteApp(name string, p string, cN string, cV string) error {

	//Construct the composite key to select the entry
	key := AppKey{
		AppName:             name,
		Project:             p,
		CompositeAppName:    cN,
		CompositeAppVersion: cV,
	}
	err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete App Entry;")
	}

	return nil
}
