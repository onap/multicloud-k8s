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

package userpermission

import (
        "encoding/json"

        "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

        pkgerrors "github.com/pkg/errors"
)

// UserPermission contains the parameters needed for a user permission
type UserPermission struct {
        UserPermissionName      string          `json:"name"`
        APIGroups               []string        `json:"apiGroups"`
        Resources               []string        `json:"resources"`
        Verbs                   []string        `json:"verbs"`
}

// UserPermissionKey is the key structure that is used in the database
type UserPermissionKey struct {
        UserPermissionName      string          `json:"name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure
func (dk UserPermissionKey) String() string {
        out, err := json.Marshal(dk)
        if err != nil {
                return ""
        }

        return string(out)
}

// UserPermissionManager is an interface that exposes the connection
// functionality
type UserPermissionManager interface {
        Create(c UserPermission) (UserPermission, error)
        Get(name string) (UserPermission, error)
        Delete(name string) error
        Update(name string, c UserPermission) (UserPermission, error)
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
               storeName:   "userpermission",
               tagMeta:     "metadata",
        }
}

// Create entry for the User Permission resource in the database
func (v *UserPermissionClient) Create(c UserPermission) (UserPermission, error) {

        //Construct key consisting of name
        key := UserPermissionKey{UserPermissionName: c.UserPermissionName}

        //Check if this User Permission already exists
        _, err := v.Get(c.UserPermissionName)
        if err == nil {
                return UserPermission{}, pkgerrors.New(
                    "User Permission already exists")
        }

        err = db.DBconn.Create(v.storeName, key, v.tagMeta, c)
        if err != nil {
                return UserPermission{}, pkgerrors.Wrap(err, "Creating DB Entry")
        }

        return c, nil
}

// Get returns User Permission for corresponding name
func (v *UserPermissionClient) Get(name string) (UserPermission, error) {

        //Construct the composite key to select the entry
        key := UserPermissionKey{UserPermissionName: name}
        value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
        if err != nil {
                return UserPermission{}, pkgerrors.Wrap(err,
                "Get User Permission")
        }

        //value is a byte array
        if value != nil {
                up := UserPermission{}
                err = db.DBconn.Unmarshal(value, &up)
                if err != nil {
                        return UserPermission{}, pkgerrors.Wrap(err,
                        "Unmarshaling value")
                }
                return up, nil
        }

        return UserPermission{}, pkgerrors.New("Error getting User Permission")
}

// Delete the User Permission entry from database
func (v *UserPermissionClient) Delete(name string) error {
        //Construct the composite key to select the entry
        key := UserPermissionKey{UserPermissionName: name}
        err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
        if err != nil {
                return pkgerrors.Wrap(err, "Delete User Permission")
        }
        return nil
}

// Update an entry for the User Permission in the database
func (v *UserPermissionClient) Update(name string, c UserPermission) (
    UserPermission, error) {

    key := UserPermissionKey{
        UserPermissionName: name,
    }
    //Check if this User Permission exists
    _, err := v.Get(name)
    if err != nil {
        return UserPermission{}, pkgerrors.New(
            "Update Error - User Permission doesn't exist")
    }
    err = db.DBconn.Update(v.storeName, key, v.tagMeta, c)
    if err != nil {
        return UserPermission{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}
