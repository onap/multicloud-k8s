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
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/handlers"
	"k8s.io/client-go/util/homedir"

	"k8-plugin-multicloud/api"
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

	router := api.NewRouter(kubeconfig)
	loggedRouter := handlers.LoggingHandler(os.Stdout, router)
	log.Println("Starting Kubernetes Multicloud API")
	log.Fatal(http.ListenAndServe(":8081", loggedRouter)) // Remove hardcode.
}
