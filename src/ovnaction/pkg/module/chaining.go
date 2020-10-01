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

// Chain defines the high level structure of a network chain document
type Chain struct {
	Metadata Metadata            `json:"metadata" yaml:"metadata"`
	Spec     NetworkChainingSpec `json:"spec" yaml:"spec"`
}

// NetworkChainingSpec contains the specification of a network chain
type NetworkChainingSpec struct {
	ChainType   string    `json:"type"`
	RoutingSpec RouteSpec `json:"routingSpec"`
}

// RouteSpec contains the routing specificaiton of a network chain
type RouteSpec struct {
	LeftNetwork  []RoutingNetwork `json:"leftNetwork"`
	RightNetwork []RoutingNetwork `json:"rightNetwork"`
	NetworkChain string           `json:"networkChain"`
	Namespace    string           `json:"namespace"`
}

// RoutingNetwork contains the route networkroute network details for en element of a network chain
type RoutingNetwork struct {
	NetworkName string `json:"networkName"`
	GatewayIP   string `json:"gatewayIp"`
	Subnet      string `json:"subnet"`
}

// ChainKey is the key structure that is used in the database
type ChainKey struct {
	Project             string `json:"project"`
	CompositeApp        string `json:"compositeapp"`
	CompositeAppVersion string `json:"compositeappversion"`
	DigName             string `json:"deploymentintentgroup"`
	NetControlIntent    string `json:"netcontrolintent"`
	NetworkChain        string `json:"networkchain"`
}

// CrChain is the structure for the Network Chain Custom Resource
type CrChain struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Chain      Chain
}

// RoutingChainType is currently only defined chaining type
const RoutingChainType = "routing"

// ChainingAPIVersion is the kubernetes version of a network chain custom resource
const ChainingAPIVersion = "k8s.plugin.opnfv.org/v1"

// ChainingKind is the Kind string for a network chain
const ChainingKind = "NetworkChaining"

// ChainManager is an interface exposing the Chain functionality
type ChainManager interface {
	CreateChain(ch Chain, pr, ca, caver, dig, netctrlint string, exists bool) (Chain, error)
	GetChain(name, pr, ca, caver, dig, netctrlint string) (Chain, error)
	GetChains(pr, ca, caver, dig, netctrlint string) ([]Chain, error)
	DeleteChain(name, pr, ca, caver, dig, netctrlint string) error
}

// ChainClient implements the Manager
// It will also be used to maintain some localized state
type ChainClient struct {
	db ClientDbInfo
}

// NewChainClient returns an instance of the ChainClient
// which implements the Manager
func NewChainClient() *ChainClient {
	return &ChainClient{
		db: ClientDbInfo{
			storeName: "orchestrator",
			tagMeta:   "chainmetadata",
		},
	}
}

// CreateChain - create a new Chain
func (v *ChainClient) CreateChain(ch Chain, pr, ca, caver, dig, netctrlint string, exists bool) (Chain, error) {
	//Construct key and tag to select the entry
	key := ChainKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		NetworkChain:        ch.Metadata.Name,
	}

	//Check if the Network Control Intent exists
	_, err := NewNetControlIntentClient().GetNetControlIntent(netctrlint, pr, ca, caver, dig)
	if err != nil {
		return Chain{}, pkgerrors.Errorf("Network Control Intent %v does not exist", netctrlint)
	}

	//Check if this Chain already exists
	_, err = v.GetChain(ch.Metadata.Name, pr, ca, caver, dig, netctrlint)
	if err == nil && !exists {
		return Chain{}, pkgerrors.New("Chain already exists")
	}

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, ch)
	if err != nil {
		return Chain{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return ch, nil
}

// GetChain returns the Chain for corresponding name
func (v *ChainClient) GetChain(name, pr, ca, caver, dig, netctrlint string) (Chain, error) {
	//Construct key and tag to select the entry
	key := ChainKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		NetworkChain:        name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return Chain{}, pkgerrors.Wrap(err, "Get Chain")
	}

	//value is a byte array
	if value != nil {
		ch := Chain{}
		err = db.DBconn.Unmarshal(value[0], &ch)
		if err != nil {
			return Chain{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return ch, nil
	}

	return Chain{}, pkgerrors.New("Error getting Chain")
}

// GetChains returns all of the Chains for for the given network control intent
func (v *ChainClient) GetChains(pr, ca, caver, dig, netctrlint string) ([]Chain, error) {
	//Construct key and tag to select the entry
	key := ChainKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		NetworkChain:        "",
	}

	var resp []Chain
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []Chain{}, pkgerrors.Wrap(err, "Get Chains")
	}

	for _, value := range values {
		cp := Chain{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []Chain{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteChain deletes the Chain from the database
func (v *ChainClient) DeleteChain(name, pr, ca, caver, dig, netctrlint string) error {

	//Construct key and tag to select the entry
	key := ChainKey{
		Project:             pr,
		CompositeApp:        ca,
		CompositeAppVersion: caver,
		DigName:             dig,
		NetControlIntent:    netctrlint,
		NetworkChain:        name,
	}

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Chain Entry;")
	}

	return nil
}
