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

/*
This file deals with the backend implementation of
Adding/Querying AppIntents for each application in the composite-app
*/

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
	ProviderName     string  `json:"provider-name,omitempty"`
	ClusterName      string  `json:"cluster-name,omitempty"`
	ClusterLabelName string  `json:"cluster-label-name,omitempty"`
	AnyOfArray       []AnyOf `json:"anyOf,omitempty"`
}

// AnyOf consists of Array of ProviderName & ClusterLabelNames
type AnyOf struct {
	ProviderName     string `json:"provider-name,omitempty"`
	ClusterName      string `json:"cluster-name,omitempty"`
	ClusterLabelName string `json:"cluster-label-name,omitempty"`
}

// IntentStruc consists of AllOfArray and AnyOfArray
type IntentStruc struct {
	AllOfArray []AllOf `json:"allOf,omitempty"`
	AnyOfArray []AnyOf `json:"anyOf,omitempty"`
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
	GetAllIntentsByApp(aN, p, ca, v, i string) (SpecData, error)
	GetAllAppIntents(p, ca, v, i string) (ApplicationsAndClusterInfo, error)
	DeleteAppIntent(ai string, p string, ca string, v string, i string) error
}

//AppIntentQueryKey required for query
type AppIntentQueryKey struct {
	AppName string `json:"app-name"`
}

// AppIntentKey is used as primary key
type AppIntentKey struct {
	Name         string `json:"appintent"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeapp"`
	Version      string `json:"compositeappversion"`
	Intent       string `json:"genericplacement"`
}

// AppIntentFindByAppKey required for query
type AppIntentFindByAppKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	Intent              string `json:"genericplacement"`
	AppName             string `json:"app-name"`
}

// ApplicationsAndClusterInfo type represents the list of
type ApplicationsAndClusterInfo struct {
	ArrayOfAppClusterInfo []AppClusterInfo `json:"applications"`
}

// AppClusterInfo is a type linking the app and the clusters
// on which they need to be installed.
type AppClusterInfo struct {
	Name       string  `json:"name"`
	AllOfArray []AllOf `json:"allOf,omitempty"`
	AnyOfArray []AnyOf `json:"anyOf,omitempty"`
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
		tagMetaData: "appintentmetadata",
	}
}

// CreateAppIntent creates an entry for AppIntent in the db.
// Other input parameters for it - projectName, compositeAppName, version, intentName.
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

	qkey := AppIntentQueryKey{
		AppName: a.Spec.AppName,
	}

	err = db.DBconn.Insert(c.storeName, akey, qkey, c.tagMetaData, a)
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

	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return AppIntent{}, pkgerrors.Wrap(err, "Get AppIntent error")
	}

	if result != nil {
		a := AppIntent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			return AppIntent{}, pkgerrors.Wrap(err, "Unmarshalling  AppIntent")
		}
		return a, nil

	}
	return AppIntent{}, pkgerrors.New("Error getting AppIntent")
}

/*
GetAllIntentsByApp takes in parameters AppName, CompositeAppName, CompositeNameVersion
and GenericPlacementIntentName. Returns SpecData which contains
all the intents for the app.
*/
func (c *AppIntentClient) GetAllIntentsByApp(aN, p, ca, v, i string) (SpecData, error) {
	k := AppIntentFindByAppKey{
		Project:             p,
		CompositeApp:        ca,
		CompositeAppVersion: v,
		Intent:              i,
		AppName:             aN,
	}
	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return SpecData{}, pkgerrors.Wrap(err, "Get AppIntent error")
	}
	var a AppIntent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		return SpecData{}, pkgerrors.Wrap(err, "Unmarshalling  AppIntent")
	}
	return a.Spec, nil

}

/*
GetAllAppIntents takes in paramaters ProjectName, CompositeAppName, CompositeNameVersion
and GenericPlacementIntentName. Returns the ApplicationsAndClusterInfo Object - an array of AppClusterInfo
*/
func (c *AppIntentClient) GetAllAppIntents(p, ca, v, i string) (ApplicationsAndClusterInfo, error) {
	k := AppIntentKey{
		Name:         "",
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		Intent:       i,
	}
	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return ApplicationsAndClusterInfo{}, pkgerrors.Wrap(err, "Get AppClusterInfo error")
	}

	var a AppIntent
	var appClusterInfoArray []AppClusterInfo

	if len(result) != 0 {
		for i := range result {
			a = AppIntent{}
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				return ApplicationsAndClusterInfo{}, pkgerrors.Wrap(err, "Unmarshalling  AppIntent")
			}
			appName := a.Spec.AppName
			allOfArray := a.Spec.Intent.AllOfArray
			anyOfArray := a.Spec.Intent.AnyOfArray
			appClusterInfo := AppClusterInfo{appName, allOfArray,
				anyOfArray}
			appClusterInfoArray = append(appClusterInfoArray, appClusterInfo)
		}
	}
	applicationsAndClusterInfo := ApplicationsAndClusterInfo{appClusterInfoArray}
	return applicationsAndClusterInfo, err
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

	err := db.DBconn.Remove(c.storeName, k)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Project entry;")
	}
	return nil

}
