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

package contextupdateserver

import (
	"context"
	"encoding/json"
	"log"

	contextpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/contextupdate"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
)

type contextupdateServer struct {
	contextpb.UnimplementedContextupdateServer
}

func (cs *contextupdateServer) UpdateAppContext(ctx context.Context, req *contextpb.ContextUpdateRequest) (*contextpb.ContextUpdateResponse, error) {
	contextUpdateReq, _ := json.Marshal(req)
	log.Println("GRPC Server received contextupdateRequest: ", string(contextUpdateReq))

	// Insert call to Server Functionality here
	//
	//

	return &contextpb.ContextUpdateResponse{AppContextUpdated: true}, nil
}

// NewContextUpdateServer exported
func NewContextupdateServer() *contextupdateServer {
	s := &contextupdateServer{}
	return s
}
