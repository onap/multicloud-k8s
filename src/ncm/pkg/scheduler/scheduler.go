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

package scheduler

import (
	"encoding/json"
	"fmt"
	"time"

	clusterPkg "github.com/onap/multicloud-k8s/src/clm/pkg/cluster"
	oc "github.com/onap/multicloud-k8s/src/ncm/internal/ovncontroller"
	ncmtypes "github.com/onap/multicloud-k8s/src/ncm/pkg/module/types"
	nettypes "github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents/types"
	appcontext "github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/installappclient"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/state"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/status"

	pkgerrors "github.com/pkg/errors"
)

// rsyncName denotes the name of the rsync controller
const rsyncName = "rsync"

// ClusterManager is an interface exposes the Cluster functionality
type SchedulerManager interface {
	ApplyNetworkIntents(clusterProvider, cluster string) error
	NetworkIntentsStatus(clusterProvider, cluster, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) (ClusterStatus, error)
	TerminateNetworkIntents(clusterProvider, cluster string) error
}

// ClusterClient implements the Manager
// It will also be used to maintain some localized state
type SchedulerClient struct {
	db ncmtypes.ClientDbInfo
}

// NewSchedulerClient returns an instance of the SchedulerClient
// which implements the Manager
func NewSchedulerClient() *SchedulerClient {
	return &SchedulerClient{
		db: ncmtypes.ClientDbInfo{
			StoreName:  "cluster",
			TagMeta:    "clustermetadata",
			TagContent: "clustercontent",
			TagState:   "stateInfo",
		},
	}
}

// ClusterStatus holds the status data prepared for cluster network intent status queries
type ClusterStatus struct {
	status.StatusResult `json:",inline"`
}

func deleteAppContext(ac appcontext.AppContext) {
	err := ac.DeleteCompositeApp()
	if err != nil {
		log.Warn(":: Error deleting AppContext ::", log.Fields{"Error": err})
	}
}

/*
queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller
and then sets the RsyncInfo global variable.
*/
func queryDBAndSetRsyncInfo() (installappclient.RsyncInfo, error) {
	client := controller.NewControllerClient()
	vals, _ := client.GetControllers()
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfo := installappclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return installappclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

/*
callRsyncInstall method shall take in the app context id and invokes the rsync service via grpc
*/
func callRsyncInstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeInstallApp(appContextID)
	if err != nil {
		return err
	}
	return nil
}

/*
callRsyncUninstall method shall take in the app context id and invokes the rsync service via grpc
*/
func callRsyncUninstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeUninstallApp(appContextID)
	if err != nil {
		return err
	}
	return nil
}

// Apply Network Intents associated with a cluster
func (v *SchedulerClient) ApplyNetworkIntents(clusterProvider, cluster string) error {

	s, err := clusterPkg.NewClusterClient().GetClusterState(clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Errorf("Error finding cluster: %v %v", clusterProvider, cluster)
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from Cluster stateInfo: " + cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.Wrap(err, "Cluster is in an invalid state: "+cluster+" "+state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		break
	case state.StateEnum.Created:
		break
	case state.StateEnum.Applied:
		return nil
	case state.StateEnum.Instantiated:
		return pkgerrors.Wrap(err, "Cluster is in an invalid state: "+cluster+" "+state.StateEnum.Instantiated)
	default:
		return pkgerrors.Wrap(err, "Cluster is in an invalid state: "+cluster+" "+stateVal)
	}

	// Make an app context for the network intent resources
	ac := appcontext.AppContext{}
	ctxVal, err := ac.InitAppContext()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext")
	}
	handle, err := ac.CreateCompositeApp()
	if err != nil {
		deleteAppContext(ac)
		return pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}

	// Add an app (fixed value) to the app context
	apphandle, err := ac.AddApp(handle, nettypes.CONTEXT_CLUSTER_APP)
	if err != nil {
		deleteAppContext(ac)
		return pkgerrors.Wrap(err, "Error adding App to AppContext")
	}

	// Add an app order instruction
	appinstr := struct {
		Apporder []string `json:"apporder"`
	}{
		[]string{nettypes.CONTEXT_CLUSTER_APP},
	}
	jinstr, err := json.Marshal(appinstr)
	if err != nil {
		deleteAppContext(ac)
		return pkgerrors.Wrap(err, "Error marshalling network intent app order instruction")
	}

	appdepinstr := struct {
		Appdep map[string]string `json:"appdependency"`
	}{
		map[string]string{nettypes.CONTEXT_CLUSTER_APP: "go"},
	}
	jdep, err := json.Marshal(appdepinstr)
	if err != nil {
		deleteAppContext(ac)
		return pkgerrors.Wrap(err, "Error marshalling network intent app dependency instruction")
	}

	_, err = ac.AddInstruction(handle, "app", "order", string(jinstr))
	if err != nil {
		deleteAppContext(ac)
		return pkgerrors.Wrap(err, "Error adding network intent app order instruction")
	}
	_, err = ac.AddInstruction(handle, "app", "dependency", string(jdep))
	if err != nil {
		deleteAppContext(ac)
		return pkgerrors.Wrap(err, "Error adding network intent app dependency instruction")
	}

	// Add a cluster to the app
	_, err = ac.AddCluster(apphandle, clusterProvider+nettypes.SEPARATOR+cluster)
	if err != nil {
		deleteAppContext(ac)
		return pkgerrors.Wrap(err, "Error adding Cluster to AppContext")
	}

	// Pass the context to the appropriate controller (just default ovncontroller now)
	// for internal controller - pass the appcontext, cluster provider and cluster names in directly
	// external controllers will be given the appcontext id and wiil have to recontstruct
	// their own context
	err = oc.Apply(ctxVal, clusterProvider, cluster)
	if err != nil {
		deleteAppContext(ac)
		return pkgerrors.Wrap(err, "Error adding Cluster to AppContext")
	}

	// call resource synchronizer to instantiate the CRs in the cluster
	err = callRsyncInstall(ctxVal)
	if err != nil {
		deleteAppContext(ac)
		return err
	}

	// update the StateInfo in the cluster db record
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Applied,
		ContextId: ctxVal.(string),
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)

	err = db.DBconn.Insert(v.db.StoreName, key, nil, v.db.TagState, s)
	if err != nil {
		log.Warn(":: Error updating Cluster state in DB ::", log.Fields{"Error": err.Error(), "cluster": cluster, "cluster provider": clusterProvider, "AppContext": ctxVal.(string)})
		return pkgerrors.Wrap(err, "Error updating the stateInfo of cluster after Apply on network intents: "+cluster)
	}

	return nil
}

// Terminate Network Intents associated with a cluster
func (v *SchedulerClient) TerminateNetworkIntents(clusterProvider, cluster string) error {
	s, err := clusterPkg.NewClusterClient().GetClusterState(clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error finding StateInfo for cluster: %v, %v", clusterProvider, cluster)
	}
	stateVal, err := state.GetCurrentStateFromStateInfo(s)
	if err != nil {
		return pkgerrors.Errorf("Error getting current state from Cluster stateInfo: " + cluster)
	}
	switch stateVal {
	case state.StateEnum.Approved:
		return pkgerrors.Wrap(err, "Cluster is in an invalid state: "+cluster+" "+state.StateEnum.Approved)
	case state.StateEnum.Terminated:
		return nil
	case state.StateEnum.Created:
		return pkgerrors.Wrap(err, "Cluster network intents have not been applied: "+cluster)
	case state.StateEnum.Applied:
		break
	case state.StateEnum.Instantiated:
		return pkgerrors.Wrap(err, "Cluster is in an invalid state: "+cluster+" "+state.StateEnum.Instantiated)
	default:
		return pkgerrors.Wrap(err, "Cluster is in an invalid state: "+cluster+" "+stateVal)
	}

	// call resource synchronizer to terminate the CRs in the cluster
	contextId := state.GetLastContextIdFromStateInfo(s)
	err = callRsyncUninstall(contextId)
	if err != nil {
		return err
	}

	// update StateInfo
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	a := state.ActionEntry{
		State:     state.StateEnum.Terminated,
		ContextId: contextId,
		TimeStamp: time.Now(),
	}
	s.Actions = append(s.Actions, a)
	err = db.DBconn.Insert(v.db.StoreName, key, nil, v.db.TagState, s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of cluster: "+cluster)
	}

	return nil
}

/*
NetworkIntentsStatus takes in cluster provider, cluster and query parameters.
This method is responsible obtaining the status of
the cluster network intents, which is made available in the appcontext
*/
func (c SchedulerClient) NetworkIntentsStatus(clusterProvider, cluster, qInstance, qType, qOutput string, qApps, qClusters, qResources []string) (ClusterStatus, error) {

	s, err := clusterPkg.NewClusterClient().GetClusterState(clusterProvider, cluster)
	if err != nil {
		return ClusterStatus{}, pkgerrors.Wrap(err, "cluster state not found")
	}

	// Prepare the apps list (just one hardcoded value)
	allApps := make([]string, 0)
	allApps = append(allApps, nettypes.CONTEXT_CLUSTER_APP)

	statusResponse, err := status.PrepareStatusResult(s, allApps, qInstance, qType, qOutput, qApps, qClusters, qResources)
	if err != nil {
		return ClusterStatus{}, err
	}
	statusResponse.Name = clusterProvider + "+" + cluster
	clStatus := ClusterStatus{
		StatusResult: statusResponse,
	}

	return clStatus, nil
}
