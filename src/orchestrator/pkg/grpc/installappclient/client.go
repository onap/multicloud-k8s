/*
Copyright 2020 Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package installappclient

import (
	"context"
	"time"

	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	installpb "github.com/onap/multicloud-k8s/src/rsync/pkg/grpc/installapp"
	pkgerrors "github.com/pkg/errors"
)

// InvokeInstallApp will make the grpc call to the resource synchronizer
// or rsync controller.
// rsync will deply the resources in the app context to the clusters as
// prepared in the app context.
func InvokeInstallApp(appContextId string) error {
	var err error
	var rpcClient installpb.InstallappClient
	var installRes *installpb.InstallAppResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn := rpc.GetRpcConn("rsync")

	if conn != nil {
		rpcClient = installpb.NewInstallappClient(conn)
		installReq := new(installpb.InstallAppRequest)
		installReq.AppContext = appContextId
		installRes, err = rpcClient.InstallApp(ctx, installReq)
		if err == nil {
			log.Info("Response from InstappApp GRPC call", log.Fields{
				"Succeeded": installRes.AppContextInstalled,
				"Message":   installRes.AppContextInstallMessage,
			})
		}
	} else {
		return pkgerrors.Errorf("InstallApp Failed - Could not get InstallAppClient: %v", "rsync")
	}

	if err == nil {
		if installRes.AppContextInstalled {
			log.Info("InstallApp Success", log.Fields{
				"AppContext": appContextId,
				"Message":    installRes.AppContextInstallMessage,
			})
			return nil
		} else {
			return pkgerrors.Errorf("InstallApp Failed: %v", installRes.AppContextInstallMessage)
		}
	}
	return err
}
