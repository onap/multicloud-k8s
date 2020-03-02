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

package rpc

import (
	"context"
	"log"

	controllerpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/controller/controller"
	syncpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/controller/sync"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

// Rpc interface used to talk a concrete RPC Connection
var RPC map[string]Client

//  is an interface for accessing RPC
type Client interface {
	// Returns nil if grpc health is good
	HealthCheck() error
	// Sends rpc to update the application context to a controller
	UpdateAppContext(ctx context.Context, in *controllerpb.ContextUpdateRequest, opts ...grpc.CallOption) (*controllerpb.ContextUpdateResponse, error)
	// Sends rpc to update the application context to a controller
	InstallApp(ctx context.Context, in *syncpb.InstallAppRequest, opts ...grpc.CallOption) (*syncpb.InstallAppResponse, error)
}

// createRpcClient creates the Rpc Client
func createClient(Host string, Port string, Name string) error {
	var err error
	var opts []grpc.DialOption

	serverAddr := Host + ":" + Port
	serverHostVerify := config.GetConfiguration().GrpcServerHostVerify
	tls := config.GetConfiguration().GrpcTLS
	//certFile := config.GetConfiguration().GrpcCert
	//keyFile := config.GetConfiguration().GrpcKey
	caFile := config.GetConfiguration().GrpcCAFile

	if *tls {
		if *caFile == "" {
			*caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(*caFile, *serverHostVerify)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(serverAddr, opts...)

	if Contains(Name, "sync") {
		RPC[Name], err = controllerpb.NewSyncClient(conn)
	} else {
		RPC[Name], err = syncpb.NewControllerClient(conn)
	}
	if err != nil {
		pkgerrors.Wrap(err, "Grpc Client Initialization failed with error")
	}

	return err
}

// InitializeRpc sets up the connection to the
// configured RPC Server to allow the application to talk to it.
func InitializeRPC(Host string, Port string, Name string) error {
	// Only support Etcd for now
	err := createClient(Host, Port, Name)
	if err != nil {
		return pkgerrors.Cause(err)
	}
	err = RPC[Name].HealthCheck()
	if err != nil {
		return pkgerrors.Cause(err)
	}
	return nil
}
