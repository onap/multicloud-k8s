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

package grpc

import (
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	controller "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
)

const RsyncName = "rsync"

// InitRsyncClient initializes connctions to the Resource Synchronizer serivice
func InitRsyncClient() bool {
	client := controller.NewControllerClient()

	vals, _ := client.GetControllers()
	found := false
	for _, v := range vals {
		if v.Metadata.Name == RsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rpc.UpdateRpcConn(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			found = true
			break
		}
	}
	return found
}
