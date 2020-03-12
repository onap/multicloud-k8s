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
}

//AppContent contains fileContent
type AppContent struct {
	FileContent string
}

// AppKey is the key structure that is used in the database
type AppKey struct {
	App                 string `json:"app"`
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
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
	CreateApp(a App, ac AppContent, p string, cN string, cV string) (App, error)
	GetApp(name string, p string, cN string, cV string) (App, error)
	GetAppContent(name string, p string, cN string, cV string) (AppContent, error)
	GetApps(p string, cN string, cV string) ([]App, error)
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
		tagMeta:    "appmetadata",
		tagContent: "appcontent",
	}
}

// CreateApp creates a new collection based on the App
func (v *AppClient) CreateApp(a App, ac AppContent, p string, cN string, cV string) (App, error) {

	//Construct the composite key to select the entry
	key := AppKey{
		App:                 a.Metadata.Name,
		Project:             p,
		CompositeApp:        cN,
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

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, a)
	if err != nil {
		return App{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagContent, ac)
	if err != nil {
		return App{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return a, nil
}

// GetApp returns the App for corresponding name
func (v *AppClient) GetApp(name string, p string, cN string, cV string) (App, error) {

	//Construct the composite key to select the entry
	key := AppKey{
		App:                 name,
		Project:             p,
		CompositeApp:        cN,
		CompositeAppVersion: cV,
	}
	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return App{}, pkgerrors.Wrap(err, "Get app")
	}

	//value is a byte array
	if value != nil {
		app := App{}
		err = db.DBconn.Unmarshal(value[0], &app)
		if err != nil {
			return App{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return app, nil
	}

	return App{}, pkgerrors.New("Error getting app")
}

// GetAppContent returns content for corresponding app
func (v *AppClient) GetAppContent(name string, p string, cN string, cV string) (AppContent, error) {

	//Construct the composite key to select the entry
	key := AppKey{
		App:                 name,
		Project:             p,
		CompositeApp:        cN,
		CompositeAppVersion: cV,
	}
	value, err := db.DBconn.Find(v.storeName, key, v.tagContent)
	if err != nil {
		return AppContent{}, pkgerrors.Wrap(err, "Get app content")
	}

	//value is a byte array
	if value != nil {
		ac := AppContent{}
		err = db.DBconn.Unmarshal(value[0], &ac)
		if err != nil {
			return AppContent{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return ac, nil
	}

	return AppContent{}, pkgerrors.New("Error getting app content")
}

// GetApps returns all Apps for given composite App
func (v *AppClient) GetApps(project, compositeApp, compositeAppVersion string) ([]App, error) {

	key := AppKey{
		App:                 "",
		Project:             project,
		CompositeApp:        compositeApp,
		CompositeAppVersion: compositeAppVersion,
	}

	var resp []App
	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []App{}, pkgerrors.Wrap(err, "Get Apps")
	}

	for _, value := range values {
		a := App{}
		err = db.DBconn.Unmarshal(value, &a)
		if err != nil {
			return []App{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		resp = append(resp, a)
	}

	return resp, nil
}

// DeleteApp deletes the  App from database
func (v *AppClient) DeleteApp(name string, p string, cN string, cV string) error {

	//Construct the composite key to select the entry
	key := AppKey{
		App:                 name,
		Project:             p,
		CompositeApp:        cN,
		CompositeAppVersion: cV,
	}
	err := db.DBconn.Remove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete App Entry;")
	}

	return nil
}
