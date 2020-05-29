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
	"context"
	"encoding/json"
	"time"

	"github.com/onap/multicloud-k8s/src/ncm/internal/grpc"
	oc "github.com/onap/multicloud-k8s/src/ncm/internal/ovncontroller"
	clusterPkg "github.com/onap/multicloud-k8s/src/ncm/pkg/cluster"
	ncmtypes "github.com/onap/multicloud-k8s/src/ncm/pkg/module/types"
	nettypes "github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents/types"
	appcontext "github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	installpb "github.com/onap/multicloud-k8s/src/rsync/pkg/grpc/installapp"

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
			TagContext: "clustercontext",
		},
	}
}

// Apply Network Intents associated with a cluster
func (v *SchedulerClient) ApplyNetworkIntents(clusterProvider, cluster string) error {

	_, err := clusterPkg.NewClusterClient().GetClusterContext(clusterProvider, cluster)
	if err == nil {
		return pkgerrors.Errorf("Cluster network intents have already been applied: %v, %v", clusterProvider, cluster)
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

	// save the context in the cluster db record
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	err = db.DBconn.Insert(v.db.StoreName, key, nil, v.db.TagContext, ctxVal)
	if err != nil {
		cleanuperr := ac.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext after DB insert failure", log.Fields{
				"cluster-provider": clusterProvider,
				"cluster":          cluster,
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
func (v *SchedulerClient) TerminateNetworkIntents(clusterProvider, cluster string) error {
	context, err := clusterPkg.NewClusterClient().GetClusterContext(clusterProvider, cluster)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error finding AppContext for cluster: %v, %v", clusterProvider, cluster)
	}

	// TODO: call resource synchronizer to terminate the CRs in the cluster

	// remove the app context
	cleanuperr := context.DeleteCompositeApp()
	if cleanuperr != nil {
		log.Warn("Error deleted AppContext", log.Fields{
			"cluster-provider": clusterProvider,
			"cluster":          cluster,
		})
	}

	// remove the app context field from the cluster db record
	key := clusterPkg.ClusterKey{
		ClusterProviderName: clusterProvider,
		ClusterName:         cluster,
	}
	err = db.DBconn.RemoveTag(v.db.StoreName, key, v.db.TagContext)
	if err != nil {
		log.Warn("Error removing AppContext from Cluster document", log.Fields{
			"cluster-provider": clusterProvider,
			"cluster":          cluster,
		})
	}
	return nil
}
