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

package logicalcloud

import (
    "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
    "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

    pkgerrors "github.com/pkg/errors"
)

// LogicalCloud contains the parameters needed for a Logical Cloud
type LogicalCloud struct {
    MetaData            MetaDataList        `json:"metadata"`
    Specification       Spec                `json:"spec"`
}


// MetaData contains the parameters needed for metadata
type MetaDataList struct {
    LogicalCloudName    string  `json:"logical-cloud-name"`
    Description         string  `json:"description"`
    UserData1           string  `json:"userData1"`
    UserData2           string  `json:"userData2"`
}

// Spec contains the parameters needed for spec
type Spec struct {
    NameSpace   string      `json:"namespace"`
    User        UserData    `json:"user"`

}

// UserData contains the parameters needed for user
type UserData struct {
    UserName            string        `json:"user-name"`
    Type                string        `json:"type"`
    UserPermissions     []UserPerm    `json:"user-permissions"`
}

//  UserPerm contains the parameters needed for user permissions
type UserPerm struct {
    PermName        string          `json:"permission-name"`
    APIGroups       []string        `json:"apiGroups"`
    Resources       []string        `json:"resources"`
    Verbs           []string        `json:"verbs"`
}

// LogicalCloudKey is the key structure that is used in the database
type LogicalCloudKey struct {
    Project             string  `json:"project"`
    LogicalCloudName    string  `json:"logical-cloud-name"`
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

// LogicalCloudClient implements the LogicalCloudManager
// It will also be used to maintain some localized state
type LogicalCloudClient struct {
    storeName   string
    tagMeta     string
}

// LogicalCloudClient returns an instance of the LogicalCloudClient
// which implements the LogicalCloudManager
func NewLogicalCloudClient() *LogicalCloudClient {
    return &LogicalCloudClient{
        storeName:   "dcm",
        tagMeta:     "logicalcloud",
    }
}

// Create entry for the logical cloud resource in the database
func (v *LogicalCloudClient) Create(project string, c LogicalCloud) (LogicalCloud, error) {

    //Construct key consisting of name
    key := LogicalCloudKey{
        Project:            project,
        LogicalCloudName:   c.MetaData.LogicalCloudName,
    }

    //Check if project exists
    _, err := module.NewProjectClient().GetProject(project)
    if err != nil {
        return LogicalCloud{}, pkgerrors.New("Unable to find the project")
    }

    //Check if this Logical Cloud already exists
    _, err = v.Get(project, c.MetaData.LogicalCloudName)
    if err == nil {
        return LogicalCloud{}, pkgerrors.New("Logical Cloud already exists")
    }

    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
    if err != nil {
        return LogicalCloud{}, pkgerrors.Wrap(err, "Creating DB Entry")
    }

    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c) 

    return c, nil
}

// Get returns Logical Cloud for corresponding to logical cloud name
func (v *LogicalCloudClient) Get(project, logicalCloudName string) (LogicalCloud, error) {

    //Construct the composite key to select the entry
    key := LogicalCloudKey{
        Project:            project,
        LogicalCloudName:   logicalCloudName,
    }
    value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
    if err != nil {
        return LogicalCloud{}, pkgerrors.Wrap(err, "Get Logical Cloud")
    }

    //value is a byte array
    if value != nil {
        lc := LogicalCloud{}
        err = db.DBconn.Unmarshal(value[0], &lc)
        if err != nil {
            return LogicalCloud{}, pkgerrors.Wrap(err, "Unmarshaling value")
        }
        return lc, nil
    }

    return LogicalCloud{}, pkgerrors.New("Error getting Logical Cloud")
}

// GetAll returns Logical Clouds in the project
func (v *LogicalCloudClient) GetAll(project string) ([]LogicalCloud, error) {

    //Construct the composite key to select the entry
    key := LogicalCloudKey{
        Project:            project,
        LogicalCloudName:   "",
    }

    var resp []LogicalCloud
    values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
    if err != nil {
        return []LogicalCloud{}, pkgerrors.Wrap(err, "Get Logical Clouds")
    }

    for _, value := range values {
        lc := LogicalCloud{}
        err = db.DBconn.Unmarshal(value, &lc)
        if err != nil {
            return []LogicalCloud{}, pkgerrors.Wrap(err, "Unmarshaling value")
        }
        resp = append(resp, lc)
    }

    return resp, nil
}

// Delete the Logical Cloud entry from database
func (v *LogicalCloudClient) Delete(project, logicalCloudName string) error {

    //Construct the composite key to select the entry
    key := LogicalCloudKey{
        Project:            project,
        LogicalCloudName:   logicalCloudName,
    }
    err := db.DBconn.Remove(v.storeName, key)
    if err != nil {
        return pkgerrors.Wrap(err, "Delete Logical Cloud")
    }

    return nil
}

// Update an entry for the Logical Cloud in the database
func (v *LogicalCloudClient) Update(project, logicalCloudName string, c LogicalCloud) (LogicalCloud, error) {

    key := LogicalCloudKey{
        Project:            project,
        LogicalCloudName: logicalCloudName,
    }
    // Check for mismatch, logicalCloudName and payload logical cloud name
    if c.MetaData.LogicalCloudName != logicalCloudName {
        return LogicalCloud{}, pkgerrors.New("Update Error - Logical Cloud name mismatch")
    }
    //Check if this Logical Cloud exists
    _, err := v.Get(project, logicalCloudName)
    if err != nil {
        return LogicalCloud{}, pkgerrors.New("Update Error - Logical Cloud doesn't exist")
    }
    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
    if err != nil {
        return LogicalCloud{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}
