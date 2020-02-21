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
	"encoding/json"
	"strings"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"

	pkgerrors "github.com/pkg/errors"
)

// ClusterProvider contains the parameters needed for ClusterProviders
// It implements the interface for managing the ClusterProviders
type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

type ClusterProvider struct {
	Metadata Metadata `json:"metadata"`
}

type Cluster struct {
	Metadata Metadata `json:"metadata"`
}

type ClusterContent struct {
	Kubeconfig string `json:"kubeconfig,omitempty"`
}

type ClusterLabel struct {
	LabelName string `json:"label-name"`
}

type ClusterKvPairs struct {
	Metadata Metadata      `json:"metadata"`
	Spec     ClusterKvSpec `json:"spec"`
}

type ClusterKvSpec struct {
	Kv []map[string]interface{} `json:"kv"`
}

// ClusterProviderKey is the key structure that is used in the database
type ClusterProviderKey struct {
	ClusterProviderName string `json:"provider"`
}

// ClusterKey is the key structure that is used in the database
type ClusterKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
}

// ClusterLabelKey is the key structure that is used in the database
type ClusterLabelKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
	ClusterLabelName    string `json:"label"`
}

// ClusterKvPairsKey is the key structure that is used in the database
type ClusterKvPairsKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
	ClusterKvPairsName  string `json:"kvname"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (pk ClusterProviderKey) String() string {
	_, err := json.Marshal(pk)
	if err != nil {
		return ""
	}
	return strings.Join([]string{"provider", pk.ClusterProviderName}, ",")
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (pk ClusterKey) String() string {
	_, err := json.Marshal(pk)
	if err != nil {
		return ""
	}
	return strings.Join([]string{"provider", pk.ClusterProviderName, "cluster", pk.ClusterName}, ",")
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (pk ClusterLabelKey) String() string {
	_, err := json.Marshal(pk)
	if err != nil {
		return ""
	}
	return strings.Join([]string{"provider", pk.ClusterProviderName, "cluster", pk.ClusterName, "label", pk.ClusterLabelName}, ",")
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (pk ClusterKvPairsKey) String() string {
	_, err := json.Marshal(pk)
	if err != nil {
		return ""
	}
	return strings.Join([]string{"provider", pk.ClusterProviderName, "cluster", pk.ClusterName, "kvpair", pk.ClusterKvPairsName}, ",")
}

// Manager is an interface exposes the Cluster functionality
type ClusterManager interface {
	CreateClusterProvider(pr ClusterProvider) (ClusterProvider, error)
	GetClusterProvider(name string) (ClusterProvider, error)
	GetClusterProviders() ([]ClusterProvider, error)
	DeleteClusterProvider(name string) error
	CreateCluster(provider string, pr Cluster, qr ClusterContent) (Cluster, error)
	GetCluster(provider, name string) (Cluster, error)
	GetClusterContent(provider, name string) (ClusterContent, error)
	GetClusters(provider string) ([]Cluster, error)
	DeleteCluster(provider, name string) error
	CreateClusterLabel(provider, cluster string, pr ClusterLabel) (ClusterLabel, error)
	GetClusterLabel(provider, cluster, label string) (ClusterLabel, error)
	GetClusterLabels(provider, cluster string) ([]ClusterLabel, error)
	DeleteClusterLabel(provider, cluster, label string) error
	CreateClusterKvPairs(provider, cluster string, pr ClusterKvPairs) (ClusterKvPairs, error)
	GetClusterKvPairs(provider, cluster, kvpair string) (ClusterKvPairs, error)
	GetAllClusterKvPairs(provider, cluster string) ([]ClusterKvPairs, error)
	DeleteClusterKvPairs(provider, cluster, kvpair string) error
}

// ClusterClient implements the Manager
// It will also be used to maintain some localized state
type ClusterClient struct {
	storeName  string
	tagMeta    string
	tagContent string
}

// NewClusterClient returns an instance of the ClusterClient
// which implements the Manager
func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		storeName:  "cluster",
		tagMeta:    "clustermetadata",
		tagContent: "clustercontent",
	}
}

// CreateClusterProvider a new collection based on the project
func (v *ClusterClient) CreateClusterProvider(p ClusterProvider) (ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: p.Metadata.Name,
	}

	//Check if this ClusterProvider already exists
	_, err := v.GetClusterProvider(p.Metadata.Name)
	if err == nil {
		return ClusterProvider{}, pkgerrors.New("ClusterProvider already exists")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, p)
	if err != nil {
		return ClusterProvider{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterProvider returns the ClusterProvider for corresponding name
func (v *ClusterClient) GetClusterProvider(name string) (ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: name,
	}

	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return ClusterProvider{}, pkgerrors.Wrap(err, "Get ClusterProvider")
	}

	//value is a byte array
	if value != nil {
		proj := ClusterProvider{}
		err = db.DBconn.Unmarshal(value[0], &proj)
		if err != nil {
			return ClusterProvider{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return proj, nil
	}

	return ClusterProvider{}, pkgerrors.New("Error getting ClusterProvider")
}

// GetClusterProviderList returns all of the ClusterProvider for corresponding name
func (v *ClusterClient) GetClusterProviders() ([]ClusterProvider, error) {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: "",
	}

	var resp []ClusterProvider
	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []ClusterProvider{}, pkgerrors.Wrap(err, "Get ClusterProviders")
	}

	for _, value := range values {
		cp := ClusterProvider{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterProvider{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteClusterProvider the  ClusterProvider from database
func (v *ClusterClient) DeleteClusterProvider(name string) error {

	//Construct key and tag to select the entry
	key := ClusterProviderKey{
		ClusterProviderName: name,
	}

	err := db.DBconn.Remove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete ClusterProvider Entry;")
	}

	//TODO: Delete the collection when the project is deleted
	return nil
}

// CreateCluster a new collection based on the project
func (v *ClusterClient) CreateCluster(provider string, p Cluster, q ClusterContent) (Cluster, error) {

	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         p.Metadata.Name,
	}

	//Verify ClusterProvider already exists
	_, err := v.GetClusterProvider(provider)
	if err != nil {
		return Cluster{}, pkgerrors.New("ClusterProvider does not exist")
	}

	//Check if this Cluster already exists
	_, err = v.GetCluster(provider, p.Metadata.Name)
	if err == nil {
		return Cluster{}, pkgerrors.New("Cluster already exists")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, p)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	err = db.DBconn.Insert(v.storeName, key, nil, v.tagContent, q)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetCluster returns the Cluster for corresponding provider and name
func (v *ClusterClient) GetCluster(provider, name string) (Cluster, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Get Cluster")
	}

	//value is a byte array
	if value != nil {
		proj := Cluster{}
		err = db.DBconn.Unmarshal(value[0], &proj)
		if err != nil {
			return Cluster{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return proj, nil
	}

	return Cluster{}, pkgerrors.New("Error getting Cluster")
}

// GetClusterContent returns the ClusterContent for corresponding provider and name
func (v *ClusterClient) GetClusterContent(provider, name string) (ClusterContent, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	value, err := db.DBconn.Find(v.storeName, key, v.tagContent)
	if err != nil {
		return ClusterContent{}, pkgerrors.Wrap(err, "Get Cluster Content")
	}

	//value is a byte array
	if value != nil {
		proj := ClusterContent{}
		err = db.DBconn.Unmarshal(value[0], &proj)
		if err != nil {
			return ClusterContent{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return proj, nil
	}

	return ClusterContent{}, pkgerrors.New("Error getting Cluster Content")
}

// GetClusters returns all the Clusters for corresponding provider
func (v *ClusterClient) GetClusters(provider string) ([]Cluster, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         "",
	}

	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []Cluster{}, pkgerrors.Wrap(err, "Get Clusters")
	}

	var resp []Cluster

	for _, value := range values {
		cp := Cluster{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []Cluster{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteCluster the  Cluster from database
func (v *ClusterClient) DeleteCluster(provider, name string) error {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	err := db.DBconn.Remove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Cluster Entry;")
	}

	//TODO: Delete the collection when the project is deleted
	return nil
}

// CreateClusterLabel a new collection based on the project
func (v *ClusterClient) CreateClusterLabel(provider string, cluster string, p ClusterLabel) (ClusterLabel, error) {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    p.LabelName,
	}

	//Verify Cluster already exists
	_, err := v.GetCluster(provider, cluster)
	if err != nil {
		return ClusterLabel{}, pkgerrors.New("Cluster does not exist")
	}

	//Check if this ClusterLabel already exists
	_, err = v.GetClusterLabel(provider, cluster, p.LabelName)
	if err == nil {
		return ClusterLabel{}, pkgerrors.New("Cluster Label already exists")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, p)
	if err != nil {
		return ClusterLabel{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterLabel returns the Cluster for corresponding provider, cluster and label
func (v *ClusterClient) GetClusterLabel(provider, cluster, label string) (ClusterLabel, error) {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    label,
	}

	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return ClusterLabel{}, pkgerrors.Wrap(err, "Get Cluster")
	}

	//value is a byte array
	if value != nil {
		proj := ClusterLabel{}
		err = db.DBconn.Unmarshal(value[0], &proj)
		if err != nil {
			return ClusterLabel{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return proj, nil
	}

	return ClusterLabel{}, pkgerrors.New("Error getting Cluster")
}

// GetClusterLabels returns the Cluster Labels for corresponding provider and cluster
func (v *ClusterClient) GetClusterLabels(provider, cluster string) ([]ClusterLabel, error) {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    "",
	}

	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []ClusterLabel{}, pkgerrors.Wrap(err, "Get Cluster Labels")
	}

	var resp []ClusterLabel

	for _, value := range values {
		cp := ClusterLabel{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterLabel{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// Delete the Cluster Label from database
func (v *ClusterClient) DeleteClusterLabel(provider, cluster, label string) error {
	//Construct key and tag to select the entry
	key := ClusterLabelKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterLabelName:    label,
	}

	err := db.DBconn.Remove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete ClusterLabel Entry;")
	}

	//TODO: Delete the collection when the project is deleted
	return nil
}

// CreateClusterKvPairs a new collection based on the project
func (v *ClusterClient) CreateClusterKvPairs(provider string, cluster string, p ClusterKvPairs) (ClusterKvPairs, error) {
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  p.Metadata.Name,
	}

	//Verify Cluster already exists
	_, err := v.GetCluster(provider, cluster)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.New("Cluster does not exist")
	}

	//Check if this ClusterKvPairs already exists
	_, err = v.GetClusterKvPairs(provider, cluster, p.Metadata.Name)
	if err == nil {
		return ClusterKvPairs{}, pkgerrors.New("Cluster KV Pair already exists")
	}

	err = db.DBconn.Insert(v.storeName, key, nil, v.tagMeta, p)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return p, nil
}

// GetClusterKvPairs returns the Cluster KeyValue pair for corresponding provider, cluster and KV pair name
func (v *ClusterClient) GetClusterKvPairs(provider, cluster, kvpair string) (ClusterKvPairs, error) {
	//Construct key and tag to select entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  kvpair,
	}

	value, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.Wrap(err, "Get Cluster")
	}

	//value is a byte array
	if value != nil {
		proj := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value[0], &proj)
		if err != nil {
			return ClusterKvPairs{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return proj, nil
	}

	return ClusterKvPairs{}, pkgerrors.New("Error getting Cluster")
}

// GetAllClusterKvPairs returns the Cluster Kv Pairs for corresponding provider and cluster
func (v *ClusterClient) GetAllClusterKvPairs(provider, cluster string) ([]ClusterKvPairs, error) {
	//Construct key and tag to select the entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  "",
	}

	values, err := db.DBconn.Find(v.storeName, key, v.tagMeta)
	if err != nil {
		return []ClusterKvPairs{}, pkgerrors.Wrap(err, "Get Cluster KV Pairs")
	}

	var resp []ClusterKvPairs

	for _, value := range values {
		cp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterKvPairs{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// DeleteClusterKvPairs the  ClusterKvPairs from database
func (v *ClusterClient) DeleteClusterKvPairs(provider, cluster, kvpair string) error {
	//Construct key and tag to select entry
	key := ClusterKvPairsKey{
		ClusterProviderName: provider,
		ClusterName:         cluster,
		ClusterKvPairsName:  kvpair,
	}

	err := db.DBconn.Remove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete ClusterKvPairs Entry;")
	}

	//TODO: Delete the collection when the project is deleted
	return nil
}
