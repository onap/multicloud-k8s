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

package contextupdateclient

import (
	"context"
	"time"

	contextpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/contextupdate"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	pkgerrors "github.com/pkg/errors"
)

// InvokeContextUpdate will make the grpc call to the specified controller
// The controller will take the specified intentName and update the AppContext
// appropriatly based on its operation as a placement or action controller.
func InvokeContextUpdate(controllerName, intentName, appContextId string) error {
	var err error
	var rpcClient contextpb.ContextupdateClient
	var updateRes *contextpb.ContextUpdateResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn := rpc.GetRpcConn(controllerName)

	if conn != nil {
		rpcClient = contextpb.NewContextupdateClient(conn)
		updateReq := new(contextpb.ContextUpdateRequest)
		updateReq.AppContext = appContextId
		updateReq.IntentName = intentName
		updateRes, err = rpcClient.UpdateAppContext(ctx, updateReq)
	} else {
		return pkgerrors.Errorf("ContextUpdate Failed - Could not get ContextupdateClient: %v", controllerName)
	}

	if err == nil {
		if updateRes.AppContextUpdated {
			log.Info("ContextUpdate Passed", log.Fields{
				"Controller": controllerName,
				"Intent":     intentName,
				"AppContext": appContextId,
				"Message":    updateRes.AppContextUpdateMessage,
			})
			return nil
		} else {
			return pkgerrors.Errorf("ContextUpdate Failed: %v", updateRes.AppContextUpdateMessage)
		}
	}
	return err
}
