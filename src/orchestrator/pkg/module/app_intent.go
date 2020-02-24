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
	"reflect"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// AppIntent has two components - metadata, spec
type AppIntent struct {
	MetaData MetaData `json:"metadata"`
	Spec     SpecData `json:"spec"`
}

// MetaData has - name, description, userdata1, userdata2
type MetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// AllOf consists of AnyOfArray and ClusterNames array
type AllOf struct {
	ClusterName string  `json:"cluster-name,omitempty"`
	ClusterLabelName string `json:"cluster-label-name,omitempty"`
	AnyOfArray  []AnyOf `json:"anyOf"`
}

// AnyOf consists of Array of ClusterLabelNames
type AnyOf struct {
	ClusterName string  `json:"cluster-name,omitempty"`
	ClusterLabelName string `json:"cluster-label-name,omitempty"`
}

// IntentStruc consists of AllOfArray and AnyOfArray
type IntentStruc struct {
	AllOfArray []AllOf `json:"allOf"`
	AnyOfArray []AnyOf `json:"anyOf"`
}

// SpecData consists of appName and intent
type SpecData struct {
	AppName string      `json:"app-name"`
	Intent  IntentStruc `json:"intent"`
}

// AppIntentManager is an interface which exposes the
// AppIntentManager functionalities
type AppIntentManager interface {
	CreateAppIntent(a AppIntent, p string, ca string, v string, i string) (AppIntent, error)
	GetAppIntent(ai string, p string, ca string, v string, i string) (AppIntent, error)
	DeleteAppIntent(ai string, p string, ca string, v string, i string) error
}

// AppIntentKey is used as primary key
type AppIntentKey struct {
	Name         string `json:"name"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeapp"`
	Version      string `json:"version"`
	Intent       string `json:"intent-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ak AppIntentKey) String() string {
	out, err := json.Marshal(ak)
	if err != nil {
		return ""
	}
	return string(out)
}

// AppIntentClient implements the AppIntentManager interface
type AppIntentClient struct {
	storeName   string
	tagMetaData string
}

// NewAppIntentClient returns an instance of AppIntentClient
func NewAppIntentClient() *AppIntentClient {
	return &AppIntentClient{
		storeName:   "orchestrator",
		tagMetaData: "appintent",
	}
}

// CreateAppIntent creates an entry for AppIntent in the db. Other input parameters for it - projectName, compositeAppName, version, intentName.
func (c *AppIntentClient) CreateAppIntent(a AppIntent, p string, ca string, v string, i string) (AppIntent, error) {

	//Check for the AppIntent already exists here.
	res, err := c.GetAppIntent(a.MetaData.Name, p, ca, v, i)
	if !reflect.DeepEqual(res, AppIntent{}) {
		return AppIntent{}, pkgerrors.New("AppIntent already exists")
	}

	//Check if project exists
	_, err = NewProjectClient().GetProject(p)
	if err != nil {
		return AppIntent{}, pkgerrors.New("Unable to find the project")
	}

	// check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return AppIntent{}, pkgerrors.New("Unable to find the composite-app")
	}

	// check if Intent exists
	_, err = NewGenericPlacementIntentClient().GetGenericPlacementIntent(i, p, ca, v)
	if err != nil {
		return AppIntent{}, pkgerrors.New("Unable to find the intent")
	}

	akey := AppIntentKey{
		Name:         a.MetaData.Name,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		Intent:       i,
	}

	err = db.DBconn.Create(c.storeName, akey, c.tagMetaData, a)
	if err != nil {
		return AppIntent{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return a, nil
}

// GetAppIntent shall take arguments - name of the app intent, name of the project, name of the composite app, version of the composite app and intent name. It shall return the AppIntent
func (c *AppIntentClient) GetAppIntent(ai string, p string, ca string, v string, i string) (AppIntent, error) {

	k := AppIntentKey{
		Name:         ai,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		Intent:       i,
	}

	result, err := db.DBconn.Read(c.storeName, k, c.tagMetaData)
	if err != nil {
		return AppIntent{}, pkgerrors.Wrap(err, "Get AppIntent error")
	}

	if result != nil {
		a := AppIntent{}
		err = db.DBconn.Unmarshal(result, &a)
		if err != nil {
			return AppIntent{}, pkgerrors.Wrap(err, "Unmarshalling  AppIntent")
		}
		return a, nil

	}
	return AppIntent{}, pkgerrors.New("Error getting AppIntent")
}

// DeleteAppIntent delete an AppIntent
func (c *AppIntentClient) DeleteAppIntent(ai string, p string, ca string, v string, i string) error {
	k := AppIntentKey{
		Name:         ai,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		Intent:       i,
	}

	err := db.DBconn.Delete(c.storeName, k, c.tagMetaData)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Project entry;")
	}
	return nil

}
