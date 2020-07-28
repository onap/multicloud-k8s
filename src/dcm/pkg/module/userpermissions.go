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
	pkgerrors "github.com/pkg/errors"
)

// UserPermission contains the parameters needed for a user permission
type UserPermission struct {
	UserPermissionName string   `json:"name"`
	APIGroups          []string `json:"apiGroups"`
	Resources          []string `json:"resources"`
	Verbs              []string `json:"verbs"`
}

// UserPermissionKey is the key structure that is used in the database
type UserPermissionKey struct {
	Project            string `json:"project"`
	LogicalCloudName   string `json:"logical-cloud-name"`
	UserPermissionName string `json:"upname"`
}

// UserPermissionManager is an interface that exposes the connection
// functionality
type UserPermissionManager interface {
	CreateUserPerm(project, logicalCloud string, c UserPermission) (UserPermission, error)
	GetUserPerm(project, logicalCloud, name string) (UserPermission, error)
	GetAllUserPerms(project, logicalCloud string) ([]UserPermission, error)
	DeleteUserPerm(project, logicalCloud, name string) error
	UpdateUserPerm(project, logicalCloud, name string, c UserPermission) (UserPermission, error)
}

// UserPermissionClient implements the UserPermissionManager
// It will also be used to maintain some localized state
type UserPermissionClient struct {
	storeName string
	tagMeta   string
	util      Utility
}

// UserPermissionClient returns an instance of the UserPermissionClient
// which implements the UserPermissionManager
func NewUserPermissionClient() *UserPermissionClient {
	service := DBService{}
	return &UserPermissionClient{
		storeName: "orchestrator",
		tagMeta:   "userpermission",
		util:      service,
	}
}

// Create entry for the User Permission resource in the database
func (v *UserPermissionClient) CreateUserPerm(project, logicalCloud string, c UserPermission) (UserPermission, error) {

	//Construct key consisting of name
	key := UserPermissionKey{
		Project:            project,
		LogicalCloudName:   logicalCloud,
		UserPermissionName: c.UserPermissionName,
	}

	//Check if project exists
	err := v.util.CheckProject(project)
	if err != nil {
		return UserPermission{}, pkgerrors.New("Unable to find the project")
	}
	//check if logical cloud exists
	err = v.util.CheckLogicalCloud(project, logicalCloud)
	if err != nil {
		return UserPermission{}, pkgerrors.New("Unable to find the logical cloud")
	}

	//Check if this User Permission already exists
	_, err = v.GetUserPerm(project, logicalCloud, c.UserPermissionName)
	if err == nil {
		return UserPermission{}, pkgerrors.New("User Permission already exists")
	}

	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return UserPermission{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns User Permission for corresponding name
func (v *UserPermissionClient) GetUserPerm(project, logicalCloud, userPermName string) (UserPermission, error) {

	//Construct the composite key to select the entry
	key := UserPermissionKey{
		Project:            project,
		LogicalCloudName:   logicalCloud,
		UserPermissionName: userPermName,
	}

	value, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return UserPermission{}, pkgerrors.Wrap(err, "Get User Permission")
	}

	//value is a byte array
	if value != nil {
		up := UserPermission{}
		err = v.util.DBUnmarshal(value[0], &up)
		if err != nil {
			return UserPermission{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		return up, nil
	}

	return UserPermission{}, pkgerrors.New("User Permission does not exist")
}

// GetAll lists all user permissions
func (v *UserPermissionClient) GetAllUserPerms(project, logicalCloud string) ([]UserPermission, error) {
	//Construct the composite key to select the entry
	key := UserPermissionKey{
		Project:            project,
		LogicalCloudName:   logicalCloud,
		UserPermissionName: "",
	}
	var resp []UserPermission
	values, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return []UserPermission{}, pkgerrors.Wrap(err, "Get All User Permissions")
	}

	for _, value := range values {
		up := UserPermission{}
		err = v.util.DBUnmarshal(value, &up)
		if err != nil {
			return []UserPermission{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		resp = append(resp, up)
	}
	return resp, nil
}

// Delete the User Permission entry from database
func (v *UserPermissionClient) DeleteUserPerm(project, logicalCloud, userPermName string) error {
	//Construct the composite key to select the entry
	key := UserPermissionKey{
		Project:            project,
		LogicalCloudName:   logicalCloud,
		UserPermissionName: userPermName,
	}
	err := v.util.DBRemove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete User Permission")
	}
	return nil
}

// Update an entry for the User Permission in the database
func (v *UserPermissionClient) UpdateUserPerm(project, logicalCloud, userPermName string, c UserPermission) (
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
	_, err := v.GetUserPerm(project, logicalCloud, userPermName)
	if err != nil {
		return UserPermission{}, pkgerrors.New(
			"User Permission does not exist")
	}
	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return UserPermission{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}
