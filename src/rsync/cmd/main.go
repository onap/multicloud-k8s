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
	"log"
	"math/rand"
	"time"
	"net"
	"google.golang.org/grpc"
	contextDb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
	installpb "rsync/pkg/grpc/installapp"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	register "rsync/pkg/grpc"
)

func startGrpcServer() error {

	host, port := register.GetServerHostPort()
        //lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", "localhost", port))
        lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
                log.Fatalf("Could not listen to port: %v", err)
        }
        var opts []grpc.ServerOption
/*
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
*/
        grpcServer := grpc.NewServer(opts...)
	installpb.RegisterContextupdateServer(grpcServer, installappserver.NewInstallAppServer())

        log.Println("Starting rsync gRPC Server")
        grpcServer.Serve(lis)
        return nil
}

func main() {

	rand.Seed(time.Now().UnixNano())

	// Initialize the mongodb
	err := db.InitializeDatabaseConnection("mco")
	if err != nil {
		fmt.Println(" Exiting mongod ")
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	// Initialize contextdb
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		fmt.Println(" Exiting etcd")
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	// Start grpc
	fmt.Println("starting rsync GRPC server..")
	err = startGrpcServer()
	if err != nil {
		log.Fatalf("GRPC server failed to start")
	}
}
