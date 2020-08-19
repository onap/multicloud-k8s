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

package main

import (
	"fmt"
	register "github.com/onap/multicloud-k8s/src/rsync/pkg/grpc"
	installpb "github.com/onap/multicloud-k8s/src/rsync/pkg/grpc/installapp"
	"github.com/onap/multicloud-k8s/src/rsync/pkg/grpc/installappserver"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
	contextDb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

func startGrpcServer() error {
	var tls bool

	if strings.Contains(config.GetConfiguration().GrpcEnableTLS, "enable") {
		tls = true
	} else {
		tls = false
	}
	certFile := config.GetConfiguration().GrpcServerCert
	keyFile := config.GetConfiguration().GrpcServerKey

	_, port := register.GetServerHostPort()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Could not listen to port: %v", err)
	}
	var opts []grpc.ServerOption
	if tls {
		if certFile == "" {
			certFile = testdata.Path("server.pem")
		}
		if keyFile == "" {
			keyFile = testdata.Path("server.key")
		}
		creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
		if err != nil {
			log.Fatalf("Could not generate credentials %v", err)
		}
		opts = []grpc.ServerOption{grpc.Creds(creds)}
	}
	grpcServer := grpc.NewServer(opts...)
	installpb.RegisterInstallappServer(grpcServer, installappserver.NewInstallAppServer())

	log.Println("Starting rsync gRPC Server")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("rsync grpc server is not serving %v", err)
	}
	return err
}

func main() {

	rand.Seed(time.Now().UnixNano())

	// Initialize the mongodb
	err := db.InitializeDatabaseConnection("mco")
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	// Initialize contextdb
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	// Start grpc
	log.Println("starting rsync GRPC server..")
	err = startGrpcServer()
	if err != nil {
		log.Fatalf("GRPC server failed to start")
	}
}
