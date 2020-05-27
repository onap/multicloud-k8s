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

package installappserver

import (
	"context"
	"encoding/json"
	"log"

	"github.com/onap/multicloud-k8s/src/rsync/pkg/grpc/installapp"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
)

type installappServer struct {
	installapp.UnimplementedInstallappServer
}

func (cs *installappServer) InstallApp(ctx context.Context, req *installapp.InstallAppRequest) (*installapp.InstallAppResponse, error) {
	installAppReq, _ := json.Marshal(req)
	log.Println("GRPC Server received installAppRequest: ", string(installAppReq))

	// Insert call to Server Functionality here
	//
	//

	return &installapp.InstallAppResponse{AppContextInstalled: true}, nil
}

// NewInstallAppServer exported
func NewInstallAppServer() *installappServer {
	s := &installappServer{}
	return s
}
