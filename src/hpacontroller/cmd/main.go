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
	"github.com/onap/multicloud-k8s/src/hpacontroller/api"
	"context"
	db "github.com/onap/multicloud-k8s/src/hpacontroller/pkg/infra/controllerdb"
	pb "github.com/onap/multicloud-k8s/src/hpacontroller/pkg/grpc/controller"
	server "github.com/onap/multicloud-k8s/src/hpacontroller/pkg/grpc/server"

	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net"
	"os"
	"os/signal"
	"time"
	//"strconv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/testdata"
	"google.golang.org/grpc/credentials"

	"github.com/gorilla/handlers"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/auth"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
	contextDb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
)

func main() {
	var err error
	var port int
	var tls bool
	var certFile string
	var keyFile string

	rand.Seed(time.Now().UnixNano())

	err = db.InitializeDatabaseConnection("mco")
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	err = db.InitializeDatabaseConnection("hpac")
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

	
	/*
	if port, err = strconv.Atoi(config.GetConfiguration().GrpcPort); err != nil { port = 9028}
	if tls, err = strconv.ParseBool(config.GetConfiguration().GrpcTls); err != nil { tls = false}
	if certFile, err = config.GetConfiguration().GrpcCert; err != nil { certFile = ""}
	if keyFile, err = config.GetConfiguration().GrpcKey; err != nil { keyFile = ""} 
	*/

	port = 9029
	tls = false
	certFile = ""
	keyFile = ""

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
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
	pb.RegisterControllerServer(grpcServer, server.NewControllerServer())
	grpcServer.Serve(lis)
	log.Println("Starting HPA Placement Controller gRPC Server")

	httpRouter := api.NewRouter(nil)
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	log.Println("Starting HPA Placement Controller API")

	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":" + config.GetConfiguration().ServicePort,
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
