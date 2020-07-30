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

	appcontext "github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// DeploymentIntentGroup shall have 2 fields - MetaData and Spec
type DeploymentIntentGroup struct {
	MetaData DepMetaData `json:"metadata"`
	Spec     DepSpecData `json:"spec"`
}

// DepMetaData has Name, description, userdata1, userdata2
type DepMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// DepSpecData has profile, version, OverrideValuesObj
type DepSpecData struct {
	Profile           string           `json:"profile"`
	Version           string           `json:"version"`
	OverrideValuesObj []OverrideValues `json:"override-values"`
}

// OverrideValues has appName and ValuesObj
type OverrideValues struct {
	AppName   string            `json:"app-name"`
	ValuesObj map[string]string `json:"values"`
}

// Values has ImageRepository
// type Values struct {
// 	ImageRepository string `json:"imageRepository"`
// }

// DeploymentIntentGroupManager is an interface which exposes the DeploymentIntentGroupManager functionality
type DeploymentIntentGroupManager interface {
	CreateDeploymentIntentGroup(d DeploymentIntentGroup, p string, ca string, v string) (DeploymentIntentGroup, error)
	GetDeploymentIntentGroup(di string, p string, ca string, v string) (DeploymentIntentGroup, error)
	GetDeploymentIntentGroupContext(di string, p string, ca string, v string) (appcontext.AppContext, string, error)
	DeleteDeploymentIntentGroup(di string, p string, ca string, v string) error
	GetAllDeploymentIntentGroups(p string, ca string, v string) ([]DeploymentIntentGroup, error)
}

// DeploymentIntentGroupKey consists of Name of the deployment group, project name, CompositeApp name, CompositeApp version
type DeploymentIntentGroupKey struct {
	Name         string `json:"deploymentintentgroup"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeapp"`
	Version      string `json:"compositeappversion"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (dk DeploymentIntentGroupKey) String() string {
	out, err := json.Marshal(dk)
	if err != nil {
		return ""
	}
	return string(out)
}

// DeploymentIntentGroupClient implements the DeploymentIntentGroupManager interface
type DeploymentIntentGroupClient struct {
	storeName   string
	tagMetaData string
	tagContext  string
}

// NewDeploymentIntentGroupClient return an instance of DeploymentIntentGroupClient which implements DeploymentIntentGroupManager
func NewDeploymentIntentGroupClient() *DeploymentIntentGroupClient {
	return &DeploymentIntentGroupClient{
		storeName:   "orchestrator",
		tagMetaData: "deploymentintentgroupmetadata",
		tagContext:  "contextid",
	}
}

// CreateDeploymentIntentGroup creates an entry for a given  DeploymentIntentGroup in the database. Other Input parameters for it - projectName, compositeAppName, version
func (c *DeploymentIntentGroupClient) CreateDeploymentIntentGroup(d DeploymentIntentGroup, p string, ca string,
	v string) (DeploymentIntentGroup, error) {

	res, err := c.GetDeploymentIntentGroup(d.MetaData.Name, p, ca, v)
	if !reflect.DeepEqual(res, DeploymentIntentGroup{}) {
		return DeploymentIntentGroup{}, pkgerrors.New("DeploymentIntent already exists")
	}

	//Check if project exists
	_, err = NewProjectClient().GetProject(p)
	if err != nil {
		return DeploymentIntentGroup{}, pkgerrors.New("Unable to find the project")
	}

	//check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return DeploymentIntentGroup{}, pkgerrors.New("Unable to find the composite-app")
	}

	gkey := DeploymentIntentGroupKey{
		Name:         d.MetaData.Name,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	err = db.DBconn.Insert(c.storeName, gkey, nil, c.tagMetaData, d)
	if err != nil {
		return DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return d, nil
}

// GetDeploymentIntentGroup returns the DeploymentIntentGroup with a given name, project, compositeApp and version of compositeApp
func (c *DeploymentIntentGroupClient) GetDeploymentIntentGroup(di string, p string, ca string, v string) (DeploymentIntentGroup, error) {

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	result, err := db.DBconn.Find(c.storeName, key, c.tagMetaData)
	if err != nil {
		return DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Get DeploymentIntentGroup error")
	}

	if result != nil {
		d := DeploymentIntentGroup{}
		err = db.DBconn.Unmarshal(result[0], &d)
		if err != nil {
			return DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Unmarshalling DeploymentIntentGroup")
		}
		return d, nil
	}

	return DeploymentIntentGroup{}, pkgerrors.New("Error getting DeploymentIntentGroup")

}

// GetAllDeploymentIntentGroups returns all the deploymentIntentGroups under a specific project, compositeApp and version
func (c *DeploymentIntentGroupClient) GetAllDeploymentIntentGroups(p string, ca string, v string) ([]DeploymentIntentGroup, error) {

	key := DeploymentIntentGroupKey{
		Name:         "",
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	//Check if project exists
	_, err := NewProjectClient().GetProject(p)
	if err != nil {
		return []DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Unable to find the project")
	}

	//check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return []DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Unable to find the composite-app, check CompositeAppName and Version")
	}
	var diList []DeploymentIntentGroup
	result, err := db.DBconn.Find(c.storeName, key, c.tagMetaData)
	if err != nil {
		return []DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Get DeploymentIntentGroup error")
	}

	for _, value := range result {
		di := DeploymentIntentGroup{}
		err = db.DBconn.Unmarshal(value, &di)
		if err != nil {
			return []DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Unmarshaling DeploymentIntentGroup")
		}
		diList = append(diList, di)
	}

	return diList, nil

}

// GetDeploymentIntentGroupContext returns the AppContent with a given DeploymentIntentname, project, compositeAppName and version of compositeApp
func (c *DeploymentIntentGroupClient) GetDeploymentIntentGroupContext(di string, p string, ca string, v string) (appcontext.AppContext, string, error) {

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	result, err := db.DBconn.Find(c.storeName, key, c.tagContext)
	if err != nil {
		return appcontext.AppContext{}, "", pkgerrors.Wrap(err, "Get DeploymentIntentGroup Context error")
	}

	if result != nil {
		ctxVal := string(result[0])
		var cc appcontext.AppContext
		_, err = cc.LoadAppContext(ctxVal)
		if err != nil {
			return appcontext.AppContext{}, "", pkgerrors.Wrap(err, "Error loading DeploymentIntentGroup Appcontext")
		}
		return cc, ctxVal, nil
	}

	return appcontext.AppContext{}, "", pkgerrors.New("Error getting DeploymentIntentGroup AppContext")
}

// DeleteDeploymentIntentGroup deletes a DeploymentIntentGroup
func (c *DeploymentIntentGroupClient) DeleteDeploymentIntentGroup(di string, p string, ca string, v string) error {
	k := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	_, _, err := c.GetDeploymentIntentGroupContext(di, p, ca, v)
	if err == nil {
		return pkgerrors.New("DeploymentIntentGroup must be terminated before it can be deleted " + di)
	}

	err = db.DBconn.Remove(c.storeName, k)
	if err != nil {
		return pkgerrors.Wrap(err, "Error deleting DeploymentIntentGroup entry")
	}
	return nil

}
