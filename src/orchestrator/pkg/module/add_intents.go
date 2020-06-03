/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by Addlicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package module

/*
This files deals with the backend implementation of adding
genericPlacementIntents to deployementIntentGroup
*/

import (
	"encoding/json"
	"reflect"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// Intent shall have 2 fields - MetaData and Spec
type Intent struct {
	MetaData IntentMetaData `json:"metadata"`
	Spec     IntentSpecData `json:"spec"`
}

// IntentMetaData has Name, Description, userdata1, userdata2
type IntentMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

// IntentSpecData has Intent
type IntentSpecData struct {
	Intent map[string]string `json:"intent"`
}

// ListOfIntents is a list of intents
type ListOfIntents struct {
	ListOfIntents []map[string]string `json:"intent"`
}

// IntentManager is an interface which exposes the IntentManager functionality
type IntentManager interface {
	AddIntent(a Intent, p string, ca string, v string, di string) (Intent, error)
	GetIntent(i string, p string, ca string, v string, di string) (Intent, error)
	GetAllIntents(p, ca, v, di string) (ListOfIntents, error)
	GetIntentByName(i, p, ca, v, di string) (IntentSpecData, error)
	DeleteIntent(i string, p string, ca string, v string, di string) error
}

// IntentKey consists of Name if the intent, Project name, CompositeApp name,
// CompositeApp version
type IntentKey struct {
	Name                  string `json:"intentname"`
	Project               string `json:"project"`
	CompositeApp          string `json:"compositeapp"`
	Version               string `json:"compositeappversion"`
	DeploymentIntentGroup string `json:"deploymentintentgroup"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (ik IntentKey) String() string {
	out, err := json.Marshal(ik)
	if err != nil {
		return ""
	}
	return string(out)
}

// IntentClient implements the AddIntentManager interface
type IntentClient struct {
	storeName   string
	tagMetaData string
}

// NewIntentClient returns an instance of AddIntentClient
func NewIntentClient() *IntentClient {
	return &IntentClient{
		storeName:   "orchestrator",
		tagMetaData: "addintent",
	}
}

/*
AddIntent adds a given intent to the deployment-intent-group and stores in the db.
Other input parameters for it - projectName, compositeAppName, version, DeploymentIntentgroupName
*/
func (c *IntentClient) AddIntent(a Intent, p string, ca string, v string, di string) (Intent, error) {

	//Check for the AddIntent already exists here.
	res, err := c.GetIntent(a.MetaData.Name, p, ca, v, di)
	if !reflect.DeepEqual(res, Intent{}) {
		return Intent{}, pkgerrors.New("Intent already exists")
	}

	//Check if project exists
	_, err = NewProjectClient().GetProject(p)
	if err != nil {
		return Intent{}, pkgerrors.New("Unable to find the project")
	}

	//check if compositeApp exists
	_, err = NewCompositeAppClient().GetCompositeApp(ca, v, p)
	if err != nil {
		return Intent{}, pkgerrors.New("Unable to find the composite-app")
	}

	//check if DeploymentIntentGroup exists
	_, err = NewDeploymentIntentGroupClient().GetDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		return Intent{}, pkgerrors.New("Unable to find the intent")
	}

	akey := IntentKey{
		Name:                  a.MetaData.Name,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	err = db.DBconn.Insert(c.storeName, akey, nil, c.tagMetaData, a)
	if err != nil {
		return Intent{}, pkgerrors.Wrap(err, "Create DB entry error")
	}
	return a, nil
}

/*
GetIntent takes in an IntentName, ProjectName, CompositeAppName, Version and DeploymentIntentGroup.
It returns the Intent.
*/
func (c *IntentClient) GetIntent(i string, p string, ca string, v string, di string) (Intent, error) {

	k := IntentKey{
		Name:                  i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return Intent{}, pkgerrors.Wrap(err, "Get Intent error")
	}

	if result != nil {
		a := Intent{}
		err = db.DBconn.Unmarshal(result[0], &a)
		if err != nil {
			return Intent{}, pkgerrors.Wrap(err, "Unmarshalling  AppIntent")
		}
		return a, nil

	}
	return Intent{}, pkgerrors.New("Error getting Intent")
}

/*
GetIntentByName takes in IntentName, projectName, CompositeAppName, CompositeAppVersion
and deploymentIntentGroupName returns the list of intents under the IntentName.
*/
func (c IntentClient) GetIntentByName(i string, p string, ca string, v string, di string) (IntentSpecData, error) {
	k := IntentKey{
		Name:                  i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}
	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return IntentSpecData{}, pkgerrors.Wrap(err, "Get AppIntent error")
	}
	var a Intent
	err = db.DBconn.Unmarshal(result[0], &a)
	if err != nil {
		return IntentSpecData{}, pkgerrors.Wrap(err, "Unmarshalling  Intent")
	}
	return a.Spec, nil
}

/*
GetAllIntents takes in projectName, CompositeAppName, CompositeAppVersion,
DeploymentIntentName . It returns ListOfIntents.
*/
func (c IntentClient) GetAllIntents(p string, ca string, v string, di string) (ListOfIntents, error) {
	k := IntentKey{
		Name:                  "",
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	result, err := db.DBconn.Find(c.storeName, k, c.tagMetaData)
	if err != nil {
		return ListOfIntents{}, pkgerrors.Wrap(err, "Get AppIntent error")
	}
	var a Intent
	var listOfMapOfIntents []map[string]string

	if len(result) != 0 {
		for i := range result {
			a = Intent{}
			err = db.DBconn.Unmarshal(result[i], &a)
			if err != nil {
				return ListOfIntents{}, pkgerrors.Wrap(err, "Unmarshalling Intent")
			}
			listOfMapOfIntents = append(listOfMapOfIntents, a.Spec.Intent)
		}
		return ListOfIntents{listOfMapOfIntents}, nil
	}
	return ListOfIntents{}, err
}

// DeleteIntent deletes a given intent tied to project, composite app and deployment intent group
func (c IntentClient) DeleteIntent(i string, p string, ca string, v string, di string) error {
	k := IntentKey{
		Name:                  i,
		Project:               p,
		CompositeApp:          ca,
		Version:               v,
		DeploymentIntentGroup: di,
	}

	err := db.DBconn.Remove(c.storeName, k)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Project entry;")
	}
	return nil
}
