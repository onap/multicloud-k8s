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

// UserPermission contains the parameters needed for a user permission
type UserPermission struct {
    UserPermissionName      string          `json:"permission-name"`
    APIGroups               []string        `json:"apiGroups"`
    Resources               []string        `json:"resources"`
    Verbs                   []string        `json:"verbs"`
}

// UserPermissionKey is the key structure that is used in the database
type UserPermissionKey struct {
    Project                 string      `json:"project"`
    LogicalCloudName        string      `json:"logical-cloud-name"`
    UserPermissionName      string      `json:"name"`
}

// UserPermissionManager is an interface that exposes the connection
// functionality
type UserPermissionManager interface {
    Create(project, logicalCloud string, c UserPermission) (UserPermission, error)
    Get(project, logicalCloud, name string) (UserPermission, error)
    GetAll(project, logicalCloud string) ([]UserPermission, error)
    Delete(project, logicalCloud, name string) error
    Update(project, logicalCloud, name string, c UserPermission) (UserPermission, error)
}

// UserPermissionClient implements the UserPermissionManager
// It will also be used to maintain some localized state
type UserPermissionClient struct {
    storeName   string
    tagMeta     string
}

// UserPermissionClient returns an instance of the UserPermissionClient
// which implements the UserPermissionManager
func NewUserPermissionClient() *UserPermissionClient {
    return &UserPermissionClient{
        storeName:   "dcm",
        tagMeta:     "metadata",
    }
}

// Create entry for the User Permission resource in the database
func (v *UserPermissionClient) Create(project, logicalCloud string, c UserPermission) (UserPermission, error) {

    //Construct key consisting of name
    key := UserPermissionKey {
        Project:            project,
        LogicalCloudName:   logicalCloud,
        UserPermissionName: c.UserPermissionName,
    }

    //Check if project exists
    _, err := module.NewProjectClient().GetProject(project)
    if err != nil {
        return UserPermission{}, pkgerrors.New("Unable to find the project")
    }
    //check if logical cloud exists
    _, err = NewLogicalCloudClient().Get(project, logicalCloud)
    if err != nil {
        return UserPermission{}, pkgerrors.New("Unable to find the logical cloud")
    }

    //Check if this User Permission already exists
    _, err = v.Get(project, logicalCloud, c.UserPermissionName)
    if err == nil {
        return UserPermission{}, pkgerrors.New("User Permission already exists")
    }

    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
    if err != nil {
        return UserPermission{}, pkgerrors.Wrap(err, "Creating DB Entry")
    }

    return c, nil
}

// Get returns User Permission for corresponding name
func (v *UserPermissionClient) Get(project, logicalCloud, userPermName string) (UserPermission, error) {

    //Construct the composite key to select the entry
    key := UserPermissionKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        UserPermissionName: userPermName,
    }

    value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
    if err != nil {
        return UserPermission{}, pkgerrors.Wrap(err, "Get User Permission")
    }

    //value is a byte array
    if value != nil {
        up := UserPermission{}
        err = db.DBconn.Unmarshal(value[0], &up)
        if err != nil {
            return UserPermission{}, pkgerrors.Wrap(err, "Unmarshaling value")
        }
        return up, nil
    }

    return UserPermission{}, pkgerrors.New("Error getting User Permission")
}

// GetAll lists all user permissions
func (v *UserPermissionClient) GetAll(project, logicalCloud string) ([]UserPermission, error) {
    //Construct the composite key to select the entry
    key := UserPermissionKey {
        Project:            project,
        LogicalCloudName:   logicalCloud,
        UserPermissionName: "",
    }
    var resp  []UserPermission
    values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
    if err != nil {
        return []UserPermission{}, pkgerrors.Wrap(err, "Get All User Permissions")
    }

    for _, value := range values {
        up := UserPermission{}
        err = db.DBconn.Unmarshal(value, &up)
        if err != nil {
            return []UserPermission{}, pkgerrors.Wrap(err, "Unmarshaling value")
        }
        resp = append(resp, up)
    }
    return resp, nil
}
// Delete the User Permission entry from database
func (v *UserPermissionClient) Delete(project, logicalCloud, userPermName string) error {
    //Construct the composite key to select the entry
    key := UserPermissionKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        UserPermissionName: userPermName,
    }
    err := db.DBconn.Remove(v.storeName, key)
        if err != nil {
                return pkgerrors.Wrap(err, "Delete User Permission")
        }
        return nil
}

// Update an entry for the User Permission in the database
func (v *UserPermissionClient) Update(project, logicalCloud, userPermName string, c UserPermission) (
    UserPermission, error) {

    key := UserPermissionKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        UserPermissionName: userPermName,
    }
    //Check for URL name and json permission name mismatch
    if c.UserPermissionName != userPermName {
        return UserPermission{}, pkgerrors.New("Update Error - Permission name mismatch")
    }
    //Check if this User Permission exists
    _, err := v.Get(project, logicalCloud, userPermName)
    if err != nil {
        return UserPermission{}, pkgerrors.New(
            "Update Error - User Permission doesn't exist")
    }
    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
    if err != nil {
        return UserPermission{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}
