/*
=======================================================================
Copyright (c) 2017-2020 Aarna Networks, Inc.
All rights reserved.
======================================================================
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
          http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
========================================================================
*/

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"example.com/middleend/app"
	"example.com/middleend/authproxy"
	"example.com/middleend/db"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

/* This is the main package of the middleend. This package
 * implements the http server which exposes service ar 9891.
 * It also intialises an API router which handles the APIs with
 * subpath /v1.
 */
func main() {
	depHandler := app.NewAppHandler()
	authProxyHandler := authproxy.NewAppHandler()
	configFile, err := os.Open("/opt/emco/config/middleend.conf")
	if err != nil {
		fmt.Printf("Failed to read middleend configuration")
		return
	}
	defer configFile.Close()

	// Read the configuration json
	byteValue, _ := ioutil.ReadAll(configFile)
	json.Unmarshal(byteValue, &depHandler.MiddleendConf)
	json.Unmarshal(byteValue, &authProxyHandler.AuthProxyConf)

	// Connect to the DB
	err = db.CreateDBClient("mongo", "mco", depHandler.MiddleendConf.Mongo)
	if err != nil {
		fmt.Println("Failed to connect to DB")
		return
	}
	// Get an instance of the OrchestrationHandler, this type implements
	// the APIs i.e CreateApp, ShowApp, DeleteApp.
	httpRouter := mux.NewRouter().PathPrefix("/middleend").Subrouter()
	loggedRouter := handlers.LoggingHandler(os.Stdout, httpRouter)
	log.Println("Starting middle end service")

	httpServer := &http.Server{
		Handler:      loggedRouter,
		Addr:         ":" + depHandler.MiddleendConf.OwnPort,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	httpRouter.HandleFunc("/healthcheck", depHandler.GetHealth).Methods("GET")

	// POST, GET, DELETE composite apps
	httpRouter.HandleFunc("/projects/{project-name}/composite-apps", depHandler.CreateApp).Methods("POST")
	//httpRouter.HandleFunc("/projects/{project-name}/composite-apps", depHandler.GetAllCaps).Methods("GET")
	httpRouter.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}",
		depHandler.GetSvc).Methods("GET")
	httpRouter.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}",
		depHandler.DelSvc).Methods("DELETE")
	// POST, GET, DELETE deployment intent groups
	httpRouter.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups",
		depHandler.CreateDig).Methods("POST")
	httpRouter.HandleFunc("/projects/{project-name}/deployment-intent-groups", depHandler.GetAllDigs).Methods("GET")
	httpRouter.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}/deployment-intent-groups/{deployment-intent-group-name}",
		depHandler.DelDig).Methods("DELETE")

	// Authproxy relates APIs
	httpRouter.HandleFunc("/login", authProxyHandler.LoginHandler).Methods("GET")
	httpRouter.HandleFunc("/callback", authProxyHandler.CallbackHandler).Methods("GET")
	httpRouter.HandleFunc("/auth", authProxyHandler.AuthHandler).Methods("GET")
	// Cluster createion API
	httpRouter.HandleFunc("/clusterproviders/{cluster-provider-name}/clusters", depHandler.CheckConnection).Methods("POST")

	// Start server in a go routine.
	go func() {
		log.Fatal(httpServer.ListenAndServe())
	}()

	// Gracefull shutdown of the server,
	// create a channel and wait for SIGINT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	log.Println("wait for signal")
	<-c
	log.Println("Bye Bye")
	httpServer.Shutdown(context.Background())
}
