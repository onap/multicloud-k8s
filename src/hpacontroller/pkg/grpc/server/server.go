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

package server

import (
	"context"
	"encoding/json"
	"log"

	pb "github.com/onap/multicloud-k8s/src/hpacontroller/pkg/grpc/controller"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
)

type controllerServer struct {
	pb.UnimplementedControllerServer
}

func (cs *controllerServer) UpdateAppContext(ctx context.Context, req *pb.ContextUpdateRequest) (*pb.ContextUpdateResponse, error) {
	contextUpdateReq, _ := json.Marshal(req) 
	log.Println("GRPC Server received UpdateAppRequest: ", string(contextUpdateReq))

	// Insert call to Server Functionality here
	//
	//

	return &pb.ContextUpdateResponse{AppContextUpdated: true}, nil
}

func (cs *controllerServer) InstallApp(ctx context.Context, req *pb.InstallAppRequest) (*pb.InstallAppResponse, error) {
	appInstallReq, _ := json.Marshal(req) 
	log.Println("GRPC Server received InstallAppRequest: ", string(appInstallReq))

	// Insert call to Server Functionality here
	//
	//

	// Replace return below with not implemented for HPA Controller Code
	// return nil, status.Errorf(codes.Unimplemented, "method InstallApp not implemented")
	return &pb.InstallAppResponse{AppContextInstalled: false}, nil
}
func (cs *controllerServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	log.Println("GRPC Server received HealthCheckRequest")
	return &pb.HealthCheckResponse{ConnectionWorking: true}, nil
}

// NewControllerServer exported
func NewControllerServer() *controllerServer {
	s := &controllerServer{}
	return s
}
