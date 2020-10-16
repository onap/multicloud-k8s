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

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	pkgerrors "github.com/pkg/errors"
)

// LogicalCloud contains the parameters needed for a Logical Cloud
type LogicalCloud struct {
	MetaData      MetaDataList `json:"metadata"`
	Specification Spec         `json:"spec"`
}

// MetaData contains the parameters needed for metadata
type MetaDataList struct {
	LogicalCloudName string `json:"name"`
	Description      string `json:"description"`
	UserData1        string `json:"userData1"`
	UserData2        string `json:"userData2"`
}

// Spec contains the parameters needed for spec
type Spec struct {
	NameSpace string   `json:"namespace"`
	User      UserData `json:"user"`
}

// UserData contains the parameters needed for user
type UserData struct {
	UserName        string     `json:"user-name"`
	Type            string     `json:"type"`
	UserPermissions []UserPerm `json:"user-permissions"`
}

//  UserPerm contains the parameters needed for user permissions
type UserPerm struct {
	PermName  string   `json:"permission-name"`
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
	Verbs     []string `json:"verbs"`
}

// LogicalCloudKey is the key structure that is used in the database
type LogicalCloudKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logical-cloud-name"`
}

// LogicalCloudManager is an interface that exposes the connection
// functionality
type LogicalCloudManager interface {
	Create(project string, c LogicalCloud) (LogicalCloud, error)
	Get(project, name string) (LogicalCloud, error)
	GetAll(project string) ([]LogicalCloud, error)
	Delete(project, name string) error
	Update(project, name string, c LogicalCloud) (LogicalCloud, error)
}

// Interface facilitates unit testing by mocking functions
type Utility interface {
	DBInsert(storeName string, key db.Key, query interface{}, meta string, c interface{}) error
	DBFind(storeName string, key db.Key, meta string) ([][]byte, error)
	DBUnmarshal(value []byte, out interface{}) error
	DBRemove(storeName string, key db.Key) error
	CheckProject(project string) error
	CheckLogicalCloud(project, logicalCloud string) error
	GetLogicalCloudContext(storeName string, key db.Key, meta string, project string, name string) (appcontext.AppContext, string, error)
	GetAppContextStatus(ac appcontext.AppContext) (*appcontext.AppContextStatus, error)
}

// LogicalCloudClient implements the LogicalCloudManager
// It will also be used to maintain some localized state
type LogicalCloudClient struct {
	storeName  string
	tagMeta    string
	tagContext string
	util       Utility
}

// Added for unit testing; implements Utility interface
type DBService struct{}

// LogicalCloudClient returns an instance of the LogicalCloudClient
// which implements the LogicalCloudManager
func NewLogicalCloudClient() *LogicalCloudClient {
	service := DBService{}
	return &LogicalCloudClient{
		storeName:  "orchestrator",
		tagMeta:    "logicalcloud",
		tagContext: "lccontext",
		util:       service,
	}
}

// Create entry for the logical cloud resource in the database
func (v *LogicalCloudClient) Create(project string, c LogicalCloud) (LogicalCloud, error) {

	//Construct key consisting of name
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: c.MetaData.LogicalCloudName,
	}

	//Check if project exists
	err := v.util.CheckProject(project)
	if err != nil {
		return LogicalCloud{}, pkgerrors.New("Unable to find the project")
	}

	//Check if this Logical Cloud already exists
	_, err = v.Get(project, c.MetaData.LogicalCloudName)
	if err == nil {
		return LogicalCloud{}, pkgerrors.New("Logical Cloud already exists")
	}

	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Error creating DB Entry")
	}

	return c, nil
}

// Get returns Logical Cloud corresponding to logical cloud name
func (v *LogicalCloudClient) Get(project, logicalCloudName string) (LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	value, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Error getting Logical Cloud")
	}

	//value is a byte array
	if value != nil {
		lc := LogicalCloud{}
		err = v.util.DBUnmarshal(value[0], &lc)
		if err != nil {
			return LogicalCloud{}, pkgerrors.Wrap(err, "Error unmarshaling value")
		}
		return lc, nil
	}

	return LogicalCloud{}, pkgerrors.New("Logical Cloud does not exist")
}

// GetAll returns Logical Clouds in the project
func (v *LogicalCloudClient) GetAll(project string) ([]LogicalCloud, error) {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: "",
	}

	var resp []LogicalCloud
	values, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return []LogicalCloud{}, pkgerrors.Wrap(err, "Get Logical Clouds")
	}

	for _, value := range values {
		lc := LogicalCloud{}
		err = v.util.DBUnmarshal(value, &lc)
		if err != nil {
			return []LogicalCloud{}, pkgerrors.Wrap(err, "Unmarshaling values")
		}
		resp = append(resp, lc)
	}

	return resp, nil
}

// Delete the Logical Cloud entry from database
func (v *LogicalCloudClient) Delete(project, logicalCloudName string) error {

	//Construct the composite key to select the entry
	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	//Check if this Logical Cloud exists
	_, err := v.Get(project, logicalCloudName)
	if err != nil {
		return pkgerrors.New("Logical Cloud does not exist")
	}

	context, _, err := v.util.GetLogicalCloudContext(v.storeName, key, v.tagMeta, project, logicalCloudName)
	// If there's no context for Logical Cloud, just go ahead and delete it now
	if err != nil {
		err = v.util.DBRemove(v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Error when deleting Logical Cloud")
		}
		return nil
	}

	// Make sure rsync status for this logical cloud is Terminated,
	// otherwise we can't remove appcontext yet
	acStatus, _ := v.util.GetAppContextStatus(context)
	switch acStatus.Status {
	case appcontext.AppContextStatusEnum.Terminated:
		// remove the appcontext
		err := context.DeleteCompositeApp()
		if err != nil {
			return pkgerrors.Wrap(err, "Error deleting AppContext CompositeApp Logical Cloud")
		}

		err = v.util.DBRemove(v.storeName, key)
		if err != nil {
			return pkgerrors.Wrap(err, "Error when deleting Logical Cloud")
		}
		return nil
	case appcontext.AppContextStatusEnum.Terminating:
		return pkgerrors.New("The Logical Cloud can't be deleted yet, it is being terminated.")
	case appcontext.AppContextStatusEnum.Instantiated:
		return pkgerrors.New("The Logical Cloud is applied, please terminate first.")
	default:
		return pkgerrors.New("The Logical Cloud can't be deleted yet at this point.")
	}
}

// Update an entry for the Logical Cloud in the database
func (v *LogicalCloudClient) Update(project, logicalCloudName string, c LogicalCloud) (LogicalCloud, error) {

	key := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloudName,
	}
	// Check for mismatch, logicalCloudName and payload logical cloud name
	if c.MetaData.LogicalCloudName != logicalCloudName {
		return LogicalCloud{}, pkgerrors.New("Logical Cloud name mismatch")
	}
	//Check if this Logical Cloud exists
	_, err := v.Get(project, logicalCloudName)
	if err != nil {
		return LogicalCloud{}, pkgerrors.New("Logical Cloud does not exist")
	}
	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return LogicalCloud{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}

// GetLogicalCloudContext returns the AppContext for corresponding provider and name
func (d DBService) GetLogicalCloudContext(storeName string, key db.Key, meta string, project string, name string) (appcontext.AppContext, string, error) {

	value, err := d.DBFind(storeName, key, meta)
	if err != nil {
		return appcontext.AppContext{}, "", pkgerrors.Wrap(err, "Get Logical Cloud Context")
	}

	//value is a [][]byte
	if value != nil {
		ctxVal := string(value[0])
		var lcc appcontext.AppContext
		_, err = lcc.LoadAppContext(ctxVal)
		if err != nil {
			return appcontext.AppContext{}, "", pkgerrors.Wrap(err, "Reinitializing Logical Cloud AppContext")
		}
		return lcc, ctxVal, nil
	}

	return appcontext.AppContext{}, "", pkgerrors.New("Error getting Logical Cloud AppContext")
}

func (d DBService) DBInsert(storeName string, key db.Key, query interface{}, meta string, c interface{}) error {

	err := db.DBconn.Insert(storeName, key, nil, meta, c)
	if err != nil {
		return pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return nil
}

func (d DBService) DBFind(storeName string, key db.Key, meta string) ([][]byte, error) {

	value, err := db.DBconn.Find(storeName, key, meta)
	if err != nil {
		return [][]byte{}, pkgerrors.Wrap(err, "Get Resource")
	}

	return value, nil
}

func (d DBService) DBUnmarshal(value []byte, out interface{}) error {

	err := db.DBconn.Unmarshal(value, out)
	if err != nil {
		return pkgerrors.Wrap(err, "Unmarshaling Value")
	}

	return nil
}

func (d DBService) DBRemove(storeName string, key db.Key) error {

	err := db.DBconn.Remove(storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Resource")
	}

	return nil
}

func (d DBService) CheckProject(project string) error {
	// Check if project exists
	_, err := module.NewProjectClient().GetProject(project)
	if err != nil {
		return pkgerrors.New("Unable to find the project")
	}

	return nil
}

func (d DBService) CheckLogicalCloud(project, logicalCloud string) error {
	// Check if logical cloud exists
	_, err := NewLogicalCloudClient().Get(project, logicalCloud)
	if err != nil {
		return pkgerrors.New("Unable to find the logical cloud")
	}

	return nil
}

func (d DBService) GetAppContextStatus(ac appcontext.AppContext) (*appcontext.AppContextStatus, error) {

	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return nil, err
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if err != nil {
		return nil, err
	}
	s, err := ac.GetValue(sh)
	if err != nil {
		return nil, err
	}
	acStatus := appcontext.AppContextStatus{}
	js, _ := json.Marshal(s)
	json.Unmarshal(js, &acStatus)

	return &acStatus, nil

}
