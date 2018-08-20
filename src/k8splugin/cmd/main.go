/*
Copyright 2018 Intel Corporation.
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
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/gorilla/handlers"
	"k8s.io/client-go/util/homedir"

	"k8splugin/api"
)

func main() {
	var kubeconfig string

	home := homedir.HomeDir()
	if home != "" {
		kubeconfig = *flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	}
	flag.Parse()

	err := api.CheckInitialSettings()
	if err != nil {
		log.Fatal(err)
	}

	httpRouter := api.NewRouter(kubeconfig)
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	log.Println("Starting Kubernetes Multicloud API")

	httpServer := &http.Server{
		Handler: loggedRouter,
		Addr:    ":8081", // Remove hardcoded port number
	}

	connectionsClose := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		httpServer.Shutdown(context.Background())
		close(connectionsClose)
	}()

	log.Fatal(httpServer.ListenAndServe())
}
