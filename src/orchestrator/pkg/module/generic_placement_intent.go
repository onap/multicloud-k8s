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

// GenericPlacementIntent shall have 2 fields - metadata and spec
type GenericPlacementIntent struct {
	MetaData GenIntentMetaData `json:"metadata"`
}

// GenIntentMetaData has name, description, userdata1, userdata2
type GenIntentMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// GenericPlacementIntentManager is an interface which exposes the GenericPlacementIntentManager functionality
type GenericPlacementIntentManager interface {
	CreateGenericPlacementIntent(g GenericPlacementIntent, p string, ca string,
		v string, digName string) (GenericPlacementIntent, error)
	GetGenericPlacementIntent(intentName string, projectName string,
		compositeAppName string, version string, digName string) (GenericPlacementIntent, error)
	DeleteGenericPlacementIntent(intentName string, projectName string,
		compositeAppName string, version string, digName string) error

	GetAllGenericPlacementIntents(p string, ca string, v string, digName string) ([]GenericPlacementIntent, error)
}

// GenericPlacementIntentKey is used as the primary key
type GenericPlacementIntentKey struct {
	Name         string `json:"genericplacement"`
	Project      string `json:"project"`
	CompositeApp string `json:"compositeapp"`
	Version      string `json:"compositeappversion"`
	DigName      string `json:"deploymentintentgroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (gk GenericPlacementIntentKey) String() string {
	out, err := json.Marshal(gk)
	if err != nil {
		return ""
	}
	return string(out)
}

// GenericPlacementIntentClient implements the GenericPlacementIntentManager interface
type GenericPlacementIntentClient struct {
	storeName   string
	tagMetaData string
}

// NewGenericPlacementIntentClient return an instance of GenericPlacementIntentClient which implements GenericPlacementIntentManager
func NewGenericPlacementIntentClient() *GenericPlacementIntentClient {
	return &GenericPlacementIntentClient{
		storeName:   "orchestrator",
		tagMetaData: "genericplacementintentmetadata",
	}
}

// CreateGenericPlacementIntent creates an entry for GenericPlacementIntent in the database. Other Input parameters for it - projectName, compositeAppName, version and deploymentIntentGroupName
func (c *GenericPlacementIntentClient) CreateGenericPlacementIntent(g GenericPlacementIntent, p string, ca string,
	v string, digName string) (GenericPlacementIntent, error) {

	// check if the genericPlacement already exists.
	res, err := c.GetGenericPlacementIntent(g.MetaData.Name, p, ca, v, digName)
	if res != (GenericPlacementIntent{}) {
		return GenericPlacementIntent{}, pkgerrors.New("Intent already exists")
	}

	//Check if project exists
	_, err = NewProjectClient().GetProject(p)
	if err != nil {
		return GenericPlacementIntent{}, pkgerrors.New("Unable to find the project")
	}

	// check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return GenericPlacementIntent{}, pkgerrors.New("Unable to find the composite-app")
	}

	// check if the deploymentIntentGrpName exists
	_, err = NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(digName, p, ca, v)
	if err != nil {
		return GenericPlacementIntent{}, pkgerrors.New("Unable to find the deployment-intent-group-name")
	}

	gkey := GenericPlacementIntentKey{
		Name:         g.MetaData.Name,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		DigName:      digName,
	}

	err = db.DBconn.Insert(c.storeName, gkey, nil, c.tagMetaData, g)
	if err != nil {
		return GenericPlacementIntent{}, pkgerrors.Wrap(err, "Create DB entry error")
	}

	return g, nil
}

// GetGenericPlacementIntent shall take arguments - name of the intent, name of the project, name of the composite app, version of the composite app and deploymentIntentGroupName. It shall return the genericPlacementIntent if its present.
func (c *GenericPlacementIntentClient) GetGenericPlacementIntent(i string, p string, ca string, v string, digName string) (GenericPlacementIntent, error) {
	key := GenericPlacementIntentKey{
		Name:         i,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		DigName:      digName,
	}

	result, err := db.DBconn.Find(c.storeName, key, c.tagMetaData)
	if err != nil {
		return GenericPlacementIntent{}, pkgerrors.Wrap(err, "Get Intent error")
	}

	if result != nil {
		g := GenericPlacementIntent{}
		err = db.DBconn.Unmarshal(result[0], &g)
		if err != nil {
			return GenericPlacementIntent{}, pkgerrors.Wrap(err, "Unmarshalling GenericPlacement Intent")
		}
		return g, nil
	}

	return GenericPlacementIntent{}, pkgerrors.New("Error getting GenericPlacementIntent")

}

// GetAllGenericPlacementIntents returns all the generic placement intents for a given compsoite app name, composite app version, project and deploymentIntentGroupName
func (c *GenericPlacementIntentClient) GetAllGenericPlacementIntents(p string, ca string, v string, digName string) ([]GenericPlacementIntent, error) {

	//Check if project exists
	_, err := NewProjectClient().GetProject(p)
	if err != nil {
		return []GenericPlacementIntent{}, pkgerrors.Wrap(err, "Unable to find the project")
	}

	// check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return []GenericPlacementIntent{}, pkgerrors.Wrap(err, "Unable to find the composite-app, check compositeApp name and version")
	}

	key := GenericPlacementIntentKey{
		Name:         "",
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		DigName:      digName,
	}

	var gpList []GenericPlacementIntent
	values, err := db.DBconn.Find(c.storeName, key, c.tagMetaData)
	if err != nil {
		return []GenericPlacementIntent{}, pkgerrors.Wrap(err, "Getting GenericPlacementIntent")
	}

	for _, value := range values {
		gp := GenericPlacementIntent{}
		err = db.DBconn.Unmarshal(value, &gp)
		if err != nil {
			return []GenericPlacementIntent{}, pkgerrors.Wrap(err, "Unmarshaling GenericPlacementIntent")
		}
		gpList = append(gpList, gp)
	}

	return gpList, nil

}

// DeleteGenericPlacementIntent the intent from the database
func (c *GenericPlacementIntentClient) DeleteGenericPlacementIntent(i string, p string, ca string, v string, digName string) error {
	key := GenericPlacementIntentKey{
		Name:         i,
		Project:      p,
		CompositeApp: ca,
		Version:      v,
		DigName:      digName,
	}

	err := db.DBconn.Remove(c.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Project entry;")
	}
	return nil
}
