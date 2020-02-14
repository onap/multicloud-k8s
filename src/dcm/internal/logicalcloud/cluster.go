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

// Cluster contains the parameters needed for a Cluster
type Cluster struct {
    ClusterName         string      `json:"cluster-name"`
    LoadBalancerIP      string      `json:"loadbalancer-ip"`
}


type ClusterKey struct {
    Project             string      `json:"project"`
    LogicalCloudName    string      `json:"logical-cloud-name"`
    ClusterProvider     string      `json:"cluster-provider-name"`
    ClusterName         string      `json:"cluster-name"`
}

// ClusterManager is an interface that exposes the connection
// functionality
type ClusterManager interface {
    Create(project, logicalCloud, clusterProvider string, c Cluster) (Cluster, error)
    Get(project, logicalCloud, clusterProvider, name string) (Cluster, error)
    GetAll(project, logicalCloud, clusterProvider string) ([]Cluster, error)
    Delete(project, logicalCloud, clusterProvider, name string) error
    Update(project, logicalCloud, clusterProvider, name string, c Cluster) (Cluster, error)
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
        storeName:   "dcm",
        tagMeta:     "cluster",
    }
}

// Create entry for the cluster resource in the database
func (v *ClusterClient) Create(project, logicalCloud, clusterProvider string, c Cluster) (Cluster, error) {

    //Construct key consisting of name
    key := ClusterKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        ClusterProvider:    clusterProvider,
        ClusterName: c.ClusterName,
    }

    //Check if project exists
    _, err := module.NewProjectClient().GetProject(project)
    if err != nil {
        return Cluster{}, pkgerrors.New("Unable to find the project")
    }
    //check if logical cloud exists
    _, err = NewLogicalCloudClient().Get(project, logicalCloud)
    if err != nil {
        return Cluster{}, pkgerrors.New("Unable to find the logical cloud")
    }
    //Check if this Cluster already exists
    _, err = v.Get(project, logicalCloud, clusterProvider, c.ClusterName)
    if err == nil {
        return Cluster{}, pkgerrors.New("Cluster already exists")
    }

    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
    if err != nil {
        return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
    }

    return c, nil
}

// Get returns  Cluster for corresponding cluster name
func (v *ClusterClient) Get(project, logicalCloud, clusterProvider, clusterName string)(Cluster, error) {

    //Construct the composite key to select the entry
    key := ClusterKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        ClusterProvider:    clusterProvider,
        ClusterName:        clusterName,
    }

    value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
    if err != nil {
        return Cluster{}, pkgerrors.Wrap(err, "Get Cluster")
    }

    //value is a byte array
    if value != nil {
        cl := Cluster{}
        err = db.DBconn.Unmarshal(value[0], &cl)
        if err != nil {
            return Cluster{}, pkgerrors.Wrap(err, "Unmarshaling value")
        }
        return cl, nil
    }

    return Cluster{}, pkgerrors.New("Error getting Cluster")
}


// GetAll returns all clusters in the logical cloud
func (v *ClusterClient) GetAll(project, logicalCloud, clusterProvider string)([]Cluster, error) {
    //Construct the composite key to select clusters
    key := ClusterKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        ClusterProvider:    clusterProvider,
        ClusterName:        "",
    }
    var resp  []Cluster
    values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
    if err != nil {
        return []Cluster{}, pkgerrors.Wrap(err, "Get All Clusters")
    }

    for _, value := range values {
        cl := Cluster{}
        err = db.DBconn.Unmarshal(value, &cl)
        if err != nil {
            return []Cluster{}, pkgerrors.Wrap(err, "Unmarshaling value")
        }
        resp = append(resp, cl)
    }

    return resp, nil
}

// Delete the Cluster entry from database
func (v *ClusterClient) Delete(project, logicalCloud, clusterProvider, clusterName string) error {
    //Construct the composite key to select the entry
    key := ClusterKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        ClusterProvider:    clusterProvider,
        ClusterName:        clusterName,
    }
    err := db.DBconn.Remove(v.storeName, key)
    if err != nil {
        return pkgerrors.Wrap(err, "Delete Cluster")
    }
    return nil
}

// Update an entry for the Cluster in the database
func (v *ClusterClient) Update(project, logicalCloud, clusterProvider, clusterName string, c Cluster) (Cluster, error) {

    key := ClusterKey{
        Project:            project,
        LogicalCloudName:   logicalCloud,
        ClusterProvider:    clusterProvider,
        ClusterName: clusterName,
    }

    //Check for nae mismatch in cluster name
    if c.ClusterName != clusterName {
        return Cluster{}, pkgerrors.New("Update Error - Cluster name mismatch")
    }
    //Check if this Cluster exists
    _, err := v.Get(project, logicalCloud, clusterProvider, clusterName)
    if err != nil {
        return Cluster{}, pkgerrors.New("Update Error - Cluster doesn't exist")
    }
    err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, c)
    if err != nil {
        return Cluster{}, pkgerrors.Wrap(err, "Updating DB Entry")
    }
    return c, nil
}
