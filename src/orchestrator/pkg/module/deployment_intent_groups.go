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
	"time"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/state"

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
	LogicalCloud string `json:"logical-cloud"`
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
	GetDeploymentIntentGroupState(di string, p string, ca string, v string) (state.StateInfo, error)
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
	tagState    string
}

// NewDeploymentIntentGroupClient return an instance of DeploymentIntentGroupClient which implements DeploymentIntentGroupManager
func NewDeploymentIntentGroupClient() *DeploymentIntentGroupClient {
	return &DeploymentIntentGroupClient{
		storeName:   "orchestrator",
		tagMetaData: "deploymentintentgroupmetadata",
		tagState:    "stateInfo",
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

	// Add the stateInfo record
	s := state.StateInfo{}
	a := state.ActionEntry{
		State:     state.StateEnum.Created,
		ContextId: "",
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(c.storeName, gkey, nil, c.tagState, s)
	if err != nil {
		return DeploymentIntentGroup{}, pkgerrors.Wrap(err, "Error updating the stateInfo of the DeploymentIntentGroup: "+d.MetaData.Name)
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

// GetDeploymentIntentGroupState returns the AppContent with a given DeploymentIntentname, project, compositeAppName and version of compositeApp
func (c *DeploymentIntentGroupClient) GetDeploymentIntentGroupState(di string, p string, ca string, v string) (state.StateInfo, error) {

	key := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}

	result, err := db.DBconn.Find(c.storeName, key, c.tagState)
	if err != nil {
		return state.StateInfo{}, pkgerrors.Wrap(err, "Get DeploymentIntentGroup StateInfo error")
	}

	if result != nil {
		s := state.StateInfo{}
		err = db.DBconn.Unmarshal(result[0], &s)
		if err != nil {
			return state.StateInfo{}, pkgerrors.Wrap(err, "Unmarshalling DeploymentIntentGroup StateInfo")
		}
		return s, nil
	}

	return state.StateInfo{}, pkgerrors.New("Error getting DeploymentIntentGroup StateInfo")
}

// DeleteDeploymentIntentGroup deletes a DeploymentIntentGroup
func (c *DeploymentIntentGroupClient) DeleteDeploymentIntentGroup(di string, p string, ca string, v string) error {
	k := DeploymentIntentGroupKey{
		Name:         di,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
	}
	s, err := c.GetDeploymentIntentGroupState(di, p, ca, v)
	if err != nil {
		return pkgerrors.Errorf("Error getting stateInfo from DeploymentIntentGroup: " + di)
	}

	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from DeploymentIntentGroup stateInfo: " + di)
	}

	if stateVal == state.StateEnum.Instantiated {
		return pkgerrors.Errorf("DeploymentIntentGroup must be terminated before it can be deleted " + di)
	}

	// remove the app contexts associated with thie Deployment Intent Group
	if stateVal == state.StateEnum.Terminated {
		// Verify that the appcontext has completed terminating
		ctxid := state.GetLastContextIdFromStateInfo(s)
		acStatus, err := state.GetAppContextStatus(ctxid)
		if err == nil &&
			!(acStatus.Status == appcontext.AppContextStatusEnum.Terminated || acStatus.Status == appcontext.AppContextStatusEnum.TerminateFailed) {
			return pkgerrors.Errorf("DeploymentIntentGroup has not completed terminating: " + di)
		}

		for _, id := range state.GetContextIdsFromStateInfo(s) {
			context, err := state.GetAppContextFromId(id)
			if err != nil {
				return pkgerrors.Wrap(err, "Error getting appcontext from Deployment Intent Group StateInfo")
			}
			err = context.DeleteCompositeApp()
			if err != nil {
				return pkgerrors.Wrap(err, "Error deleting appcontext for Deployment Intent Group")
			}
		}
	}

	err = db.DBconn.Remove(c.storeName, k)
	if err != nil {
		return pkgerrors.Wrap(err, "Error deleting DeploymentIntentGroup entry")
	}
	return nil

}
