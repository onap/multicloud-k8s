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

// Cluster contains the parameters needed for a Cluster
type Cluster struct {
	MetaData      ClusterMeta `json:"metadata"`
	Specification ClusterSpec `json:"spec"`
}

type ClusterMeta struct {
	ClusterReference string `json:"name"`
	Description      string `json:"description"`
	UserData1        string `json:"userData1"`
	UserData2        string `json:"userData2"`
}

type ClusterSpec struct {
	ClusterProvider string `json:"cluster-provider"`
	ClusterName     string `json:"cluster-name"`
	LoadBalancerIP  string `json:"loadbalancer-ip"`
}

type ClusterKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logical-cloud-name"`
	ClusterReference string `json:"clname"`
}

// ClusterManager is an interface that exposes the connection
// functionality
type ClusterManager interface {
	CreateCluster(project, logicalCloud string, c Cluster) (Cluster, error)
	GetCluster(project, logicalCloud, name string) (Cluster, error)
	GetAllClusters(project, logicalCloud string) ([]Cluster, error)
	DeleteCluster(project, logicalCloud, name string) error
	UpdateCluster(project, logicalCloud, name string, c Cluster) (Cluster, error)
}

// ClusterClient implements the ClusterManager
// It will also be used to maintain some localized state
type ClusterClient struct {
	storeName string
	tagMeta   string
	util      Utility
}

// ClusterClient returns an instance of the ClusterClient
// which implements the ClusterManager
func NewClusterClient() *ClusterClient {
	service := DBService{}
	return &ClusterClient{
		storeName: "orchestrator",
		tagMeta:   "cluster",
		util:      service,
	}
}

// Create entry for the cluster reference resource in the database
func (v *ClusterClient) CreateCluster(project, logicalCloud string, c Cluster) (Cluster, error) {

	//Construct key consisting of name
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: c.MetaData.ClusterReference,
	}

	//Check if project exists
	err := v.util.CheckProject(project)
	if err != nil {
		return Cluster{}, pkgerrors.New("Unable to find the project")
	}
	//check if logical cloud exists
	err = v.util.CheckLogicalCloud(project, logicalCloud)
	if err != nil {
		return Cluster{}, pkgerrors.New("Unable to find the logical cloud")
	}
	//Check if this Cluster reference already exists
	_, err = v.GetCluster(project, logicalCloud, c.MetaData.ClusterReference)
	if err == nil {
		return Cluster{}, pkgerrors.New("Cluster reference already exists")
	}

	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns  Cluster for corresponding cluster reference
func (v *ClusterClient) GetCluster(project, logicalCloud, clusterReference string) (Cluster, error) {

	//Construct the composite key to select the entry
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	value, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Get Cluster reference")
	}

	//value is a byte array
	if value != nil {
		cl := Cluster{}
		err = v.util.DBUnmarshal(value[0], &cl)
		if err != nil {
			return Cluster{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		return cl, nil
	}

	return Cluster{}, pkgerrors.New("Error getting Cluster")
}

// GetAll returns all cluster references in the logical cloud
func (v *ClusterClient) GetAllClusters(project, logicalCloud string) ([]Cluster, error) {
	//Construct the composite key to select clusters
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: "",
	}
	var resp []Cluster
	values, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return []Cluster{}, pkgerrors.Wrap(err, "Get All Cluster references")
	}

	for _, value := range values {
		cl := Cluster{}
		err = v.util.DBUnmarshal(value, &cl)
		if err != nil {
			return []Cluster{}, pkgerrors.Wrap(err, "Unmarshaling values")
		}
		resp = append(resp, cl)
	}

	return resp, nil
}

// Delete the Cluster reference entry from database
func (v *ClusterClient) DeleteCluster(project, logicalCloud, clusterReference string) error {
	//Construct the composite key to select the entry
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}
	err := v.util.DBRemove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Cluster Reference")
	}
	return nil
}

// Update an entry for the Cluster reference in the database
func (v *ClusterClient) UpdateCluster(project, logicalCloud, clusterReference string, c Cluster) (Cluster, error) {

	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	//Check for name mismatch in cluster reference
	if c.MetaData.ClusterReference != clusterReference {
		return Cluster{}, pkgerrors.New("Update Error - Cluster reference mismatch")
	}
	//Check if this Cluster reference exists
	_, err := v.GetCluster(project, logicalCloud, clusterReference)
	if err != nil {
		return Cluster{}, pkgerrors.New("Update Error - Cluster reference doesn't exist")
	}
	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}
