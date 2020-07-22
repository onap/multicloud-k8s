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

	clusterPkg "github.com/onap/multicloud-k8s/src/clm/pkg/cluster"
	oc "github.com/onap/multicloud-k8s/src/ncm/internal/ovncontroller"
	ncmtypes "github.com/onap/multicloud-k8s/src/ncm/pkg/module/types"
	nettypes "github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents/types"
	appcontext "github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/installappclient"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/state"

	pkgerrors "github.com/pkg/errors"
)

// ClusterManager is an interface exposes the Cluster functionality
type SchedulerManager interface {
	ApplyNetworkIntents(clusterProvider, cluster string) error
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

// Apply Network Intents associated with a cluster
func (v *SchedulerClient) ApplyNetworkIntents(clusterProvider, cluster string) error {

	s, err := clusterPkg.NewClusterClient().GetClusterState(clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Errorf("Error finding cluster: %v %v", clusterProvider, cluster)
	}
	switch s.State {
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
		return pkgerrors.Wrap(err, "Cluster is in an invalid state: "+cluster+" "+s.State)
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
	apphandle, err := ac.AddApp(handle, nettypes.CONTEXT_CLUSTER_APP)
	if err != nil {
		cleanuperr := ac.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext CompositeApp create failure", log.Fields{
				"cluster-provider": clusterProvider,
				"cluster":          cluster,
			})
		}
		return pkgerrors.Wrap(err, "Error adding App to AppContext")
	}

	// Add an app order instruction
	appinstr := struct {
		Apporder []string `json:"apporder"`
	}{
		[]string{nettypes.CONTEXT_CLUSTER_APP},
	}
	jinstr, _ := json.Marshal(appinstr)

	appdepinstr := struct {
		Appdep map[string]string `json:"appdependency"`
	}{
		map[string]string{nettypes.CONTEXT_CLUSTER_APP: "go"},
	}
	jdep, _ := json.Marshal(appdepinstr)

	_, err = ac.AddInstruction(handle, "app", "order", string(jinstr))
	_, err = ac.AddInstruction(handle, "app", "dependency", string(jdep))

	// Add a cluster to the app
	_, err = ac.AddCluster(apphandle, clusterProvider+nettypes.SEPARATOR+cluster)
	if err != nil {
		cleanuperr := ac.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext after add cluster failure", log.Fields{
				"cluster-provider": clusterProvider,
				"cluster":          cluster,
			})
		}
		return pkgerrors.Wrap(err, "Error adding Cluster to AppContext")
	}

	// Pass the context to the appropriate controller (just default ovncontroller now)
	// for internal controller - pass the appcontext, cluster provider and cluster names in directly
	// external controllers will be given the appcontext id and wiil have to recontstruct
	// their own context
	err = oc.Apply(ctxVal, clusterProvider, cluster)
	if err != nil {
		cleanuperr := ac.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext after controller failure", log.Fields{
				"cluster-provider": clusterProvider,
				"cluster":          cluster,
			})
		}
		return pkgerrors.Wrap(err, "Error adding Cluster to AppContext")
	}

	// update the StateInfo in the cluster db record
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	stateInfo := state.StateInfo{
		State:     state.StateEnum.Applied,
		ContextId: ctxVal.(string),
	}

	err = db.DBconn.Insert(v.db.StoreName, key, nil, v.db.TagState, stateInfo)
	if err != nil {
		cleanuperr := ac.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext after DB insert failure", log.Fields{
				"cluster-provider": clusterProvider,
				"cluster":          cluster,
			})
		}
		return pkgerrors.Wrap(err, "Error updating the stateInfo of cluster: "+cluster)
	}

	// call resource synchronizer to instantiate the CRs in the cluster
	err = installappclient.InvokeInstallApp(ctxVal.(string))
	if err != nil {
		return err
	}

	return nil
}

// Terminate Network Intents associated with a cluster
func (v *SchedulerClient) TerminateNetworkIntents(clusterProvider, cluster string) error {
	s, err := clusterPkg.NewClusterClient().GetClusterState(clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error finding StateInfo for cluster: %v, %v", clusterProvider, cluster)
	}
	switch s.State {
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
		return pkgerrors.Wrap(err, "Cluster is in an invalid state: "+cluster+" "+s.State)
	}

	// call resource synchronizer to terminate the CRs in the cluster
	err = installappclient.InvokeUninstallApp(s.ContextId)
	if err != nil {
		return err
	}

	// remove the app context
	context, err := state.GetAppContextFromStateInfo(s)
	if err != nil {
		return pkgerrors.Wrap(err, "Error getting appcontext from cluster StateInfo : "+clusterProvider+" "+cluster)
	}
	err = context.DeleteCompositeApp()
	if err != nil {
		return pkgerrors.Wrap(err, "Error deleting appcontext of cluster : "+clusterProvider+" "+cluster)
	}

	// update StateInfo
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	stateInfo := state.StateInfo{
		State:     state.StateEnum.Terminated,
		ContextId: "",
	}

	err = db.DBconn.Insert(v.db.StoreName, key, nil, v.db.TagState, stateInfo)
	if err != nil {
		return pkgerrors.Wrap(err, "Error updating the stateInfo of cluster: "+cluster)
	}

	return nil
}
