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
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"github.com/onap/multicloud-k8s/src/ncm/api"
	register "github.com/onap/multicloud-k8s/src/ncm/pkg/grpc"
	"github.com/onap/multicloud-k8s/src/ncm/pkg/grpc/contextupdateserver"
	"github.com/onap/multicloud-k8s/src/ncm/pkg/grpc/healthcheckserver"
	updatepb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/contextupdate"
	healthpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/healthcheck"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/auth"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
	contextDb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

func startGrpcServer() error {
	var tls bool
	var certFile string
	var keyFile string

	tls = false
	certFile = ""
	keyFile = ""

	host, port := register.GetServerHostPort()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
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
	healthpb.RegisterHealthcheckServer(grpcServer, healthcheckserver.NewHealthcheckServer())
	updatepb.RegisterContextupdateServer(grpcServer, contextupdateserver.NewContextupdateServer())

	log.Println("Starting Network Configuration Manager gRPC Server")
	grpcServer.Serve(lis)
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	err := db.InitializeDatabaseConnection("mco")
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	httpRouter := api.NewRouter(nil)
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	log.Println("Starting Network Customization Manager")

	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + config.GetConfiguration().ServicePort,
	}

	go func() {
		err := startGrpcServer()
		if err != nil {
			log.Fatalf("GRPC server failed to start")
		}
	}()
	// Register GRPC server to the database
	err = register.RegisterGrpcServer(register.GetServerHostPort())
	if err != nil {
		log.Fatalf("GRPC server failed to register with DB")
	}

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		httpServer.Shutdown(context.Background())
		close(connectionsClose)
	}()

	tlsConfig, err := auth.GetTLSConfig("ca.cert", "server.cert", "server.key")
	if err != nil {
		log.Println("Error Getting TLS Configuration. Starting without TLS...")
		log.Fatal(httpServer.ListenAndServe())
	} else {
		httpServer.TLSConfig = tlsConfig
		// empty strings because tlsconfig already has this information
		err = httpServer.ListenAndServeTLS("", "")
	}
}
