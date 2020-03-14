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
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// ProviderNet contains the parameters needed for dynamic networks
type ProviderNet struct {
	Metadata Metadata        `json:"metadata"`
	Spec     ProviderNetSpec `json:"spec"`
}

type ProviderNetSpec struct {
	CniType         string       `json:"cniType"`
	Ipv4Subnets     []Ipv4Subnet `json:"ipv4Subnets"`
	ProviderNetType string       `json:"providerNetType"`
	Vlan            Vlan         `json:"vlan"`
}

// ProviderNetKey is the key structure that is used in the database
type ProviderNetKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
	ProviderNetName     string `json:"providernet"`
}

// Manager is an interface exposing the ProviderNet functionality
type ProviderNetManager interface {
	CreateProviderNet(pr ProviderNet, clusterProvider, cluster string, exists bool) (ProviderNet, error)
	GetProviderNet(name, clusterProvider, cluster string) (ProviderNet, error)
	GetProviderNets(clusterProvider, cluster string) ([]ProviderNet, error)
	DeleteProviderNet(name, clusterProvider, cluster string) error
}

// ProviderNetClient implements the Manager
// It will also be used to maintain some localized state
type ProviderNetClient struct {
	db ClientDbInfo
}

// NewProviderNetClient returns an instance of the ProviderNetClient
// which implements the Manager
func NewProviderNetClient() *ProviderNetClient {
	return &ProviderNetClient{
		db: ClientDbInfo{
			storeName: "cluster",
			tagMeta:   "networkmetadata",
		},
	}
}

// CreateProviderNet - create a new ProviderNet
func (v *ProviderNetClient) CreateProviderNet(p ProviderNet, clusterProvider, cluster string, exists bool) (ProviderNet, error) {

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     p.Metadata.Name,
	}

	//Check if cluster exists
	_, err := NewClusterClient().GetCluster(clusterProvider, cluster)
	if err != nil {
		return ProviderNet{}, pkgerrors.New("Unable to find the cluster")
	}

	//Check if this ProviderNet already exists
	_, err = v.GetProviderNet(p.Metadata.Name, clusterProvider, cluster)
	if err == nil && !exists {
		return ProviderNet{}, pkgerrors.New("ProviderNet already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return ProviderNet{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetProviderNet returns the ProviderNet for corresponding name
func (v *ProviderNetClient) GetProviderNet(name, clusterProvider, cluster string) (ProviderNet, error) {

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ProviderNet{}, pkgerrors.Wrap(err, "Get ProviderNet")
	}

	//value is a byte array
	if value != nil {
		cp := ProviderNet{}
		err = db.DBconn.Unmarshal(value[0], &cp)
		if err != nil {
			return ProviderNet{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cp, nil
	}

	return ProviderNet{}, pkgerrors.New("Error getting ProviderNet")
}

// GetProviderNetList returns all of the ProviderNet for corresponding name
func (v *ProviderNetClient) GetProviderNets(clusterProvider, cluster string) ([]ProviderNet, error) {

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     "",
	}

	var resp []ProviderNet
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ProviderNet{}, pkgerrors.Wrap(err, "Get ProviderNets")
	}

	for _, value := range values {
		cp := ProviderNet{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ProviderNet{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// Delete the  ProviderNet from database
func (v *ProviderNetClient) DeleteProviderNet(name, clusterProvider, cluster string) error {

	//Construct key and tag to select the entry
	key := ProviderNetKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		ProviderNetName:     name,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete ProviderNet Entry;")
	}

	return nil
}
