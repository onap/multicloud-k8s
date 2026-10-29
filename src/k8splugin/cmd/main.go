/*
Copyright 2018 Intel Corporation.
Copyright © 2021 Samsung Electronics
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
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/onap/multicloud-k8s/src/k8splugin/api"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/auth"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/tracing"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/utils"
)

func main() {

	err := utils.CheckInitialSettings()
	if err != nil {
		log.Fatal("Invalid initial settings", log.Fields{"error": err})
	}

	rand.Seed(time.Now().UnixNano())

	i := tracing.NewInstrumentation()
	defer func() {
		if err := i.TracerProvider.Shutdown(context.Background()); err != nil {
			log.Error("Error shutting down tracer provider", log.Fields{"error": err})
		}
	}()

	httpRouter := api.NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	log.Info("Starting Kubernetes Multicloud API", nil)

	httpServer := &http.Server{
		Handler: httpRouter,
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
		log.Warn("Error Getting TLS Configuration. Starting without TLS...", log.Fields{"error": err})
		log.Fatal("HTTP server stopped", log.Fields{"error": httpServer.ListenAndServe()})
	} else {
		httpServer.TLSConfig = tlsConfig
		// empty strings because tlsconfig already has this information
		err = httpServer.ListenAndServeTLS("", "")
	}
}
