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
	"log"
	"strings"

	controllerpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/controller"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

// RPC interface used to talk a concrete RPC Connection
var RPC map[string]controllerpb.ControllerClient

type ContextUpdateRequest interface {
}

type ContextUpdateResponse interface {
}

type InstallAppRequest interface {
}

type InstallAppResponse interface {
}

type HealthCheckRequest interface {
}

type HealthCheckResponse interface {
}

// createRpcClient creates the Rpc Client
func createClient(Host string, Port string, Name string) error {
	var err error
	var tls bool
	var opts []grpc.DialOption
	if RPC == nil {
		RPC = make(map[string]controllerpb.ControllerClient)
	}

	serverAddr := Host + ":" + Port
	serverHostVerify := config.GetConfiguration().GrpcServerHostVerify

	if strings.Contains(config.GetConfiguration().GrpcTLS, "enable") {
		tls = true
	} else {
		tls = false
	}
	//certFile := config.GetConfiguration().GrpcCert
	//keyFile := config.GetConfiguration().GrpcKey
	caFile := config.GetConfiguration().GrpcCAFile

	if tls {
		if caFile == "" {
			caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(caFile, serverHostVerify)
		if err != nil {
			log.Fatalf("Failed to create TLS credentials %v", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	RPC[Name] = controllerpb.NewControllerClient(conn)

	if err != nil {
		pkgerrors.Wrap(err, "Grpc Client Initialization failed with error")
	}

	return err
}

// InitializeRPC sets up the connection to the
// configured RPC Server to allow the application to talk to it.
func InitializeRPC(Host string, Port string, Name string) error {
	err := createClient(Host, Port, Name)
	if err != nil {
		return pkgerrors.Cause(err)
	}
	return nil
}
