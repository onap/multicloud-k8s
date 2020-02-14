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
        "encoding/json"

        "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

        pkgerrors "github.com/pkg/errors"
)

// LogicalCloud contains the parameters needed for a Logical Cloud
type LogicalCloud struct {
        MetaData            MetaDataList        `json:"metadata"`
        Specification       Spec                `json:"spec"`
}


// MetaData contains the parameters needed for metadata
type MetaDataList struct {
        LogicalCloudName    string  `json:"name"`
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
        LogicalCloudName    string `json:"name"`
        ClusterReference    string `json:"cluster-reference"`
}

/*type ClusterKey struct {
        ClusterReference    string `json:"cluster-reference"`
}*/

// We will use json marshalling to convert to string to
// preserve the underlying structure
func (dk LogicalCloudKey) String() string {
        out, err := json.Marshal(dk)
        if err != nil {
                return ""
        }

        return string(out)
}

// LogicalCloudManager is an interface that exposes the connection
// functionality
type LogicalCloudManager interface {
        Create(c LogicalCloud) (LogicalCloud, error)
        Get(name string) (LogicalCloud, error)
        Delete(name string) error
        Update(name string, c LogicalCloud) (LogicalCloud, error)
//      Associate(name string, c Cluster) (Cluster, error)
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
               storeName:   "logicalcloud",
               tagMeta:     "metadata",
        }
}

// Create entry for the logical cloud resource in the database
func (v *LogicalCloudClient) Create(c LogicalCloud) (LogicalCloud, error) {

        //Construct key consisting of name
        key := LogicalCloudKey{LogicalCloudName: c.MetaData.LogicalCloudName}

        //Check if this Logical CLoud already exists
        _, err := v.Get(c.MetaData.LogicalCloudName)
        if err == nil {
                return LogicalCloud{}, pkgerrors.New("Logical Cloud already exists")
        }

        err = db.DBconn.Create(v.storeName, key, v.tagMeta, c)
        if err != nil {
                return LogicalCloud{}, pkgerrors.Wrap(err, "Creating DB Entry")
        }

        return c, nil
}

// Get returns Logical Cloud for correspondin name
func (v *LogicalCloudClient) Get(name string) (LogicalCloud, error) {

        //Construct the composite key to select the entry
        key := LogicalCloudKey{LogicalCloudName: name}
        value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
        if err != nil {
                return LogicalCloud{}, pkgerrors.Wrap(err, "Get Logical Cloud")
        }

        //value is a byte array
        if value != nil {
                lc := LogicalCloud{}
                err = db.DBconn.Unmarshal(value, &lc)
                if err != nil {
                        return LogicalCloud{}, pkgerrors.Wrap(err, "Unmarshaling value")
                }
                return lc, nil
        }

        return LogicalCloud{}, pkgerrors.New("Error getting Logical Cloud")
}

// Delete the Logical Cloud entry from database
func (v *LogicalCloudClient) Delete(name string) error {
        //Construct the composite key to select the entry
        key := LogicalCloudKey{LogicalCloudName: name}
        err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
        if err != nil {
                return pkgerrors.Wrap(err, "Delete Logical Cloud")
        }
        return nil
}

// Download kubeconfig
/*func (v *LogicalCloudClient) Download(name string) (string, error) {
    cluster, err := v.Get(name)
    if err != nil {
        return "", pkgerrors.Wrap(err, "Getting Cluster info")
    }

    //Decode the Kubeconfig from base64 to string
    kubecontent, err := base64.StdEncoding.DecodeString(cluster.Kubeconfig)
    if err != nil {
        return "", pkgerrors.Wrap(err, "Converting fron base64")
    }

    // Create temp file to write kubeconfig
    // Assume this file will be deleted after usage by the consumer
    tempF, err := ioutil.TempFile("", "kube-config-temp")
    if err != nil {
        return "", pkgerrors.Wrap(err, "Creating temp file")
    }

    _, err = tempF.Write(kubeContent)
    if err != nil{
        return "", pkgerrors.Wrap(err, "Writing kubeconfig to file")
    }

    return tempF.Name(), nil
}*/

// Update an entry for the Logical Cloud in the database
func (v *LogicalCloudClient) Update(name string, c LogicalCloud) (LogicalCloud, error) {

    key := LogicalCloudKey{
        LogicalCloudName: name,
    }
    //Check if this Logical Cloud exists
    _, err := v.Get(name)
    if err != nil {
        return LogicalCloud{}, pkgerrors.New("Update Error - Logical Cloud doesn't exist")
    }
    err = db.DBconn.Update(v.storeName, key, v.tagMeta, c)
    if err != nil {
        return LogicalCloud{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}

// Associate cluster with logical cloud
/*func (v *LogicalCloudClient) Update(name string, c Cluster) (Cluster, error) {
    key := LogicalCloudKey {
        LogicalCloudName: name,
    }
    //Check if this Logical Cloud exists
    _, err := v.Get(name)
    if err != nil {
        return Cluster{}, pkgerrors.New("Error Associating Cluster - Logical Cloud doesn't exist")
    }
    err = db.DBconn.Update(v.storeName, key, v.tagMeta, c)
    if err != nil {
        return Cluster{}, pkgerrors.Wrap(err, "Associating cluster to Logical Cloud")
    }
    return c, nil
}*/
