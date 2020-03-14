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

// Network contains the parameters needed for dynamic networks
type Network struct {
	Metadata Metadata    `json:"metadata"`
	Spec     NetworkSpec `json:"spec"`
}

type NetworkSpec struct {
	CniType     string       `json:"cniType"`
	Ipv4Subnets []Ipv4Subnet `json:"ipv4Subnets"`
}

// NetworkKey is the key structure that is used in the database
type NetworkKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
	NetworkName         string `json:"network"`
}

// Manager is an interface exposing the Network functionality
type NetworkManager interface {
	CreateNetwork(pr Network, clusterProvider, cluster string, exists bool) (Network, error)
	GetNetwork(name, clusterProvider, cluster string) (Network, error)
	GetNetworks(clusterProvider, cluster string) ([]Network, error)
	DeleteNetwork(name, clusterProvider, cluster string) error
}

// NetworkClient implements the Manager
// It will also be used to maintain some localized state
type NetworkClient struct {
	db ClientDbInfo
}

// NewNetworkClient returns an instance of the NetworkClient
// which implements the Manager
func NewNetworkClient() *NetworkClient {
	return &NetworkClient{
		db: ClientDbInfo{
			storeName: "cluster",
			tagMeta:   "networkmetadata",
		},
	}
}

// CreateNetwork - create a new Network
func (v *NetworkClient) CreateNetwork(p Network, clusterProvider, cluster string, exists bool) (Network, error) {

	//Construct key and tag to select the entry
	key := NetworkKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		NetworkName:         p.Metadata.Name,
	}

	//Check if cluster exists
	_, err := NewClusterClient().GetCluster(clusterProvider, cluster)
	if err != nil {
		return Network{}, pkgerrors.New("Unable to find the cluster")
	}

	//Check if this Network already exists
	_, err = v.GetNetwork(p.Metadata.Name, clusterProvider, cluster)
	if err == nil && !exists {
		return Network{}, pkgerrors.New("Network already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return Network{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetNetwork returns the Network for corresponding name
func (v *NetworkClient) GetNetwork(name, clusterProvider, cluster string) (Network, error) {

	//Construct key and tag to select the entry
	key := NetworkKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		NetworkName:         name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return Network{}, pkgerrors.Wrap(err, "Get Network")
	}

	//value is a byte array
	if value != nil {
		cp := Network{}
		err = db.DBconn.Unmarshal(value[0], &cp)
		if err != nil {
			return Network{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cp, nil
	}

	return Network{}, pkgerrors.New("Error getting Network")
}

// GetNetworkList returns all of the Network for corresponding name
func (v *NetworkClient) GetNetworks(clusterProvider, cluster string) ([]Network, error) {

	//Construct key and tag to select the entry
	key := NetworkKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		NetworkName:         "",
	}

	var resp []Network
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []Network{}, pkgerrors.Wrap(err, "Get Networks")
	}

	for _, value := range values {
		cp := Network{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []Network{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// Delete the  Network from database
func (v *NetworkClient) DeleteNetwork(name, clusterProvider, cluster string) error {

	//Construct key and tag to select the entry
	key := NetworkKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
		NetworkName:         name,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Network Entry;")
	}

	return nil
}
