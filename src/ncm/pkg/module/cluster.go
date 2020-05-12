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
	"context"
	"time"

	"github.com/onap/multicloud-k8s/src/ncm/pkg/grpc"
	appcontext "github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	installpb "github.com/onap/multicloud-k8s/src/rsync/pkg/grpc/installapp"
	"gopkg.in/yaml.v2"

	pkgerrors "github.com/pkg/errors"
)

// ClusterProvider contains the parameters needed for ClusterProviders
type ClusterProvider struct {
	Metadata Metadata `json:"metadata"`
}

type Cluster struct {
	Metadata Metadata `json:"metadata"`
}

type ClusterContent struct {
	Kubeconfig string `json:"kubeconfig"`
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

// LabelKey is the key structure that is used in the database
type LabelKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterLabelName    string `json:"label"`
}

// ClusterKvPairsKey is the key structure that is used in the database
type ClusterKvPairsKey struct {
	ClusterProviderName string `json:"provider"`
	ClusterName         string `json:"cluster"`
	ClusterKvPairsName  string `json:"kvname"`
}

const SEPARATOR = "+"
const CONTEXT_CLUSTER_APP = "network-intents"
const CONTEXT_CLUSTER_RESOURCE = "network-intents"

// ClusterManager is an interface exposes the Cluster functionality
type ClusterManager interface {
	CreateClusterProvider(pr ClusterProvider) (ClusterProvider, error)
	GetClusterProvider(name string) (ClusterProvider, error)
	GetClusterProviders() ([]ClusterProvider, error)
	DeleteClusterProvider(name string) error
	CreateCluster(provider string, pr Cluster, qr ClusterContent) (Cluster, error)
	GetCluster(provider, name string) (Cluster, error)
	GetClusterContent(provider, name string) (ClusterContent, error)
	GetClusterContext(provider, name string) (appcontext.AppContext, error)
	GetClusters(provider string) ([]Cluster, error)
	GetClustersWithLabel(provider, label string) ([]string, error)
	DeleteCluster(provider, name string) error
	ApplyNetworkIntents(provider, name string) error
	TerminateNetworkIntents(provider, name string) error
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
	db ClientDbInfo
}

// NewClusterClient returns an instance of the ClusterClient
// which implements the Manager
func NewClusterClient() *ClusterClient {
	return &ClusterClient{
		db: ClientDbInfo{
			storeName:  "cluster",
			tagMeta:    "clustermetadata",
			tagContent: "clustercontent",
			tagContext: "clustercontext",
		},
	}
}

// CreateClusterProvider - create a new Cluster Provider
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

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
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

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterProvider{}, pkgerrors.Wrap(err, "Get ClusterProvider")
	}

	//value is a byte array
	if value != nil {
		cp := ClusterProvider{}
		err = db.DBconn.Unmarshal(value[0], &cp)
		if err != nil {
			return ClusterProvider{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cp, nil
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
	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterProvider{}, pkgerrors.Wrap(err, "Get ClusterProviders")
	}

	for _, value := range values {
		cp := ClusterProvider{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterProvider{}, pkgerrors.Wrap(err, "Unmarshalling Value")
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

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete ClusterProvider Entry;")
	}

	return nil
}

// CreateCluster - create a new Cluster for a cluster-provider
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

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}
	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagContent, q)
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

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Get Cluster")
	}

	//value is a byte array
	if value != nil {
		cl := Cluster{}
		err = db.DBconn.Unmarshal(value[0], &cl)
		if err != nil {
			return Cluster{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cl, nil
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

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagContent)
	if err != nil {
		return ClusterContent{}, pkgerrors.Wrap(err, "Get Cluster Content")
	}

	//value is a byte array
	if value != nil {
		cc := ClusterContent{}
		err = db.DBconn.Unmarshal(value[0], &cc)
		if err != nil {
			return ClusterContent{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cc, nil
	}

	return ClusterContent{}, pkgerrors.New("Error getting Cluster Content")
}

// GetClusterContext returns the AppContext for corresponding provider and name
func (v *ClusterClient) GetClusterContext(provider, name string) (appcontext.AppContext, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagContext)
	if err != nil {
		return appcontext.AppContext{}, pkgerrors.Wrap(err, "Get Cluster Context")
	}

	//value is a byte array
	if value != nil {
		ctxVal := string(value[0])
		var cc appcontext.AppContext
		_, err = cc.LoadAppContext(ctxVal)
		if err != nil {
			return appcontext.AppContext{}, pkgerrors.Wrap(err, "Reinitializing Cluster AppContext")
		}
		return cc, nil
	}

	return appcontext.AppContext{}, pkgerrors.New("Error getting Cluster AppContext")
}

// GetClusters returns all the Clusters for corresponding provider
func (v *ClusterClient) GetClusters(provider string) ([]Cluster, error) {
	//Construct key and tag to select the entry
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         "",
	}

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []Cluster{}, pkgerrors.Wrap(err, "Get Clusters")
	}

	var resp []Cluster

	for _, value := range values {
		cp := Cluster{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []Cluster{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		resp = append(resp, cp)
	}

	return resp, nil
}

// GetClustersWithLabel returns all the Clusters with Labels for provider
// Support Query like /cluster-providers/{Provider}/clusters?label={label}
func (v *ClusterClient) GetClustersWithLabel(provider, label string) ([]string, error) {
	//Construct key and tag to select the entry
	key := LabelKey{
		ClusterProviderName: provider,
		ClusterLabelName:    label,
	}

	values, err := db.DBconn.Find(v.db.storeName, key, "cluster")
	if err != nil {
		return []string{}, pkgerrors.Wrap(err, "Get Clusters by label")
	}
	var resp []string

	for _, value := range values {
		cp := string(value)
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
	_, err := v.GetClusterContext(provider, name)
	if err == nil {
		return pkgerrors.Errorf("Cannot delete cluster until context is deleted: %v, %v", provider, name)
	}

	err = db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Cluster Entry;")
	}

	return nil
}

// Apply Network Intents associated with a cluster
func (v *ClusterClient) ApplyNetworkIntents(provider, name string) error {

	_, err := v.GetClusterContext(provider, name)
	if err == nil {
		return pkgerrors.Errorf("Cluster network intents have already been applied: %v, %v", provider, name)
	}

	type resource struct {
		name  string
		value string
	}

	var resources []resource

	// Find all Network Intents for this cluster
	networkIntents, err := NewNetworkClient().GetNetworks(provider, name)
	if err != nil {
		return pkgerrors.Wrap(err, "Error finding Network Intents")
	}
	for _, intent := range networkIntents {
		var crNetwork = CrNetwork{
			ApiVersion: NETWORK_APIVERSION,
			Kind:       NETWORK_KIND,
		}
		crNetwork.Network = intent
		// Produce the yaml CR document for each intent
		y, err := yaml.Marshal(&crNetwork)
		if err != nil {
			log.Info("Error marshalling network intent to yaml", log.Fields{
				"error":  err,
				"intent": intent,
			})
			continue
		}
		resources = append(resources, resource{
			name:  intent.Metadata.Name + SEPARATOR + NETWORK_KIND,
			value: string(y),
		})
	}

	// Find all Provider Network Intents for this cluster
	providerNetworkIntents, err := NewProviderNetClient().GetProviderNets(provider, name)
	if err != nil {
		return pkgerrors.Wrap(err, "Error finding Provider Network Intents")
	}
	for _, intent := range providerNetworkIntents {
		var crProviderNet = CrProviderNet{
			ApiVersion: PROVIDER_NETWORK_APIVERSION,
			Kind:       PROVIDER_NETWORK_KIND,
		}
		crProviderNet.ProviderNet = intent
		// Produce the yaml CR document for each intent
		y, err := yaml.Marshal(&crProviderNet)
		if err != nil {
			log.Info("Error marshalling provider network intent to yaml", log.Fields{
				"error":  err,
				"intent": intent,
			})
			continue
		}
		resources = append(resources, resource{
			name:  intent.Metadata.Name + SEPARATOR + PROVIDER_NETWORK_KIND,
			value: string(y),
		})
	}

	if len(resources) == 0 {
		return nil
	}

	// Make an app context for the network intent resources
	ac := appcontext.AppContext{}
	ctxVal, err := ac.InitAppContext()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext")
	}
	handle, err := ac.CreateCompositeApp()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}

	// Add an app (fixed value) to the app context
	apphandle, err := ac.AddApp(handle, CONTEXT_CLUSTER_APP)
	if err != nil {
		cleanuperr := ac.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext CompositeApp create failure", log.Fields{
				"cluster-provider": provider,
				"cluster":          name,
			})
		}
		return pkgerrors.Wrap(err, "Error adding App to AppContext")
	}

	// Add a cluster to the app
	clusterhandle, err := ac.AddCluster(apphandle, provider+SEPARATOR+name)
	if err != nil {
		cleanuperr := ac.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext after add cluster failure", log.Fields{
				"cluster-provider": provider,
				"cluster":          name,
			})
		}
		return pkgerrors.Wrap(err, "Error adding Cluster to AppContext")
	}

	// add the resources to the app context
	for _, resource := range resources {
		_, err = ac.AddResource(clusterhandle, resource.name, resource.value)
		if err != nil {
			cleanuperr := ac.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add resource failure", log.Fields{
					"cluster-provider": provider,
					"cluster":          name,
					"resource":         resource.name,
				})
			}
			return pkgerrors.Wrap(err, "Error adding Resource to AppContext")
		}
	}

	// save the context in the cluster db record
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}
	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagContext, ctxVal)
	if err != nil {
		cleanuperr := ac.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext after DB insert failure", log.Fields{
				"cluster-provider": provider,
				"cluster":          name,
			})
		}
		return pkgerrors.Wrap(err, "Error adding AppContext to DB")
	}

	// call resource synchronizer to instantiate the CRs in the cluster
	conn := rpc.GetRpcConn(grpc.RsyncName)
	if conn == nil {
		grpc.InitRsyncClient()
		conn = rpc.GetRpcConn(grpc.RsyncName)
	}

	var rpcClient installpb.InstallappClient
	var installRes *installpb.InstallAppResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if conn != nil {
		rpcClient = installpb.NewInstallappClient(conn)
		installReq := new(installpb.InstallAppRequest)
		installReq.AppContext = ctxVal.(string)
		installRes, err = rpcClient.InstallApp(ctx, installReq)
		if err == nil {
			log.Info("Response from InstappApp GRPC call", log.Fields{
				"Succeeded": installRes.AppContextInstalled,
				"Message":   installRes.AppContextInstallMessage,
			})
		}
	} else {
		return pkgerrors.Errorf("InstallApp Failed - Could not get InstallAppClient: %v", grpc.RsyncName)
	}

	return nil
}

// Terminate Network Intents associated with a cluster
func (v *ClusterClient) TerminateNetworkIntents(provider, name string) error {
	context, err := v.GetClusterContext(provider, name)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error finding AppContext for cluster: %v, %v", provider, name)
	}

	// TODO: call resource synchronizer to terminate the CRs in the cluster

	// remove the app context
	cleanuperr := context.DeleteCompositeApp()
	if cleanuperr != nil {
		log.Warn("Error deleted AppContext", log.Fields{
			"cluster-provider": provider,
			"cluster":          name,
		})
	}

	// remove the app context field from the cluster db record
	key := ClusterKey{
		ClusterProviderName: provider,
		ClusterName:         name,
	}
	err = db.DBconn.RemoveTag(v.db.storeName, key, v.db.tagContext)
	if err != nil {
		log.Warn("Error removing AppContext from Cluster document", log.Fields{
			"cluster-provider": provider,
			"cluster":          name,
		})
	}
	return nil
}

// CreateClusterLabel - create a new Cluster Label mongo document for a cluster-provider/cluster
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

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
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

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterLabel{}, pkgerrors.Wrap(err, "Get Cluster")
	}

	//value is a byte array
	if value != nil {
		cl := ClusterLabel{}
		err = db.DBconn.Unmarshal(value[0], &cl)
		if err != nil {
			return ClusterLabel{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return cl, nil
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

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterLabel{}, pkgerrors.Wrap(err, "Get Cluster Labels")
	}

	var resp []ClusterLabel

	for _, value := range values {
		cp := ClusterLabel{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterLabel{}, pkgerrors.Wrap(err, "Unmarshalling Value")
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

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete ClusterLabel Entry;")
	}

	return nil
}

// CreateClusterKvPairs - Create a New Cluster KV pairs document
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

	err = db.DBconn.Insert(v.db.storeName, key, nil, v.db.tagMeta, p)
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

	value, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return ClusterKvPairs{}, pkgerrors.Wrap(err, "Get Cluster")
	}

	//value is a byte array
	if value != nil {
		ckvp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value[0], &ckvp)
		if err != nil {
			return ClusterKvPairs{}, pkgerrors.Wrap(err, "Unmarshalling Value")
		}
		return ckvp, nil
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

	values, err := db.DBconn.Find(v.db.storeName, key, v.db.tagMeta)
	if err != nil {
		return []ClusterKvPairs{}, pkgerrors.Wrap(err, "Get Cluster KV Pairs")
	}

	var resp []ClusterKvPairs

	for _, value := range values {
		cp := ClusterKvPairs{}
		err = db.DBconn.Unmarshal(value, &cp)
		if err != nil {
			return []ClusterKvPairs{}, pkgerrors.Wrap(err, "Unmarshalling Value")
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

	err := db.DBconn.Remove(v.db.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete ClusterKvPairs Entry;")
	}

	return nil
}
