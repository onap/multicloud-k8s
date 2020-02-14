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

package cluster

import (
        "encoding/json"

       "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

        pkgerrors "github.com/pkg/errors"
)

// Cluster contains the parameters needed for a Cluster
type Cluster struct {
        ClusterName         string      `json:"cluster-name"`
        LoadBalancerIP      string      `json:"loadbalancer-ip"`
}


type ClusterKey struct {
        ClusterName         string      `json:"cluster-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure
func (dk ClusterKey) String() string {
        out, err := json.Marshal(dk)
        if err != nil {
                return ""
        }

        return string(out)
}

// ClusterManager is an interface that exposes the connection
// functionality
type ClusterManager interface {
        Create(c Cluster) (Cluster, error)
        Get(name string) (Cluster, error)
        Delete(name string) error
        Update(name string, c Cluster) (Cluster, error)
}

// ClusterClient implements the ClusterManager
// It will also be used to maintain some localized state
type ClusterClient struct {
        storeName   string
        tagMeta     string
}

// ClusterClient returns an instance of the ClusterClient
// which implements the ClusterManager
func NewClusterClient() *ClusterClient {
        return &ClusterClient{
               storeName:   "cluster",
               tagMeta:     "metadata",
        }
}

// Create entry for the cluster resource in the database
func (v *ClusterClient) Create(c Cluster) (Cluster, error) {

        //Construct key consisting of name
        key := ClusterKey{ClusterName: c.ClusterName}

        //Check if this Cluster already exists
        _, err := v.Get(c.ClusterName)
        if err == nil {
                return Cluster{}, pkgerrors.New("Cluster already exists")
        }

        err = db.DBconn.Create(v.storeName, key, v.tagMeta, c)
        if err != nil {
                return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
        }

        return c, nil
}

// Get returns  Cluster for corresponding name
func (v *ClusterClient) Get(name string) (Cluster, error) {

        //Construct the composite key to select the entry
        key := ClusterKey{ClusterName: name}
        value, err := db.DBconn.Read(v.storeName, key, v.tagMeta)
        if err != nil {
                return Cluster{}, pkgerrors.Wrap(err, "Get Cluster")
        }

        //value is a byte array
        if value != nil {
                cl := Cluster{}
                err = db.DBconn.Unmarshal(value, &cl)
                if err != nil {
                        return Cluster{}, pkgerrors.Wrap(err, "Unmarshaling value")
                }
                return cl, nil
        }

        return Cluster{}, pkgerrors.New("Error getting Cluster")
}

// Delete the Cluster entry from database
func (v *ClusterClient) Delete(name string) error {
        //Construct the composite key to select the entry
        key := ClusterKey{ClusterName: name}
        err := db.DBconn.Delete(v.storeName, key, v.tagMeta)
        if err != nil {
                return pkgerrors.Wrap(err, "Delete Cluster")
        }
        return nil
}

// Update an entry for the Cluster in the database
func (v *ClusterClient) Update(name string, c Cluster) (Cluster, error) {

    key := ClusterKey{
        ClusterName: name,
    }
    //Check if this Cluster exists
    _, err := v.Get(name)
    if err != nil {
        return Cluster{}, pkgerrors.New("Update Error - Cluster doesn't exist")
    }
    err = db.DBconn.Update(v.storeName, key, v.tagMeta, c)
    if err != nil {
        return Cluster{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}
