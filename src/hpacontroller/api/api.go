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
package api

import (
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)
var moduleClient *moduleLib.Client

func NewRouter(projectClient moduleLib.ProjectManager) *mux.Router {

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()
	moduleClient = moduleLib.NewClient()

	if projectClient == nil {
		projectClient = moduleClient.Project
	}
	projHandler := projectHandler{
		client: projectClient,
	}
	router.HandleFunc("/projects", projHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}", projHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}", projHandler.deleteHandler).Methods("DELETE")

	return router
}
