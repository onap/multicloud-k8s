/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package api

import (
	"github.com/gorilla/mux"
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"
)

var moduleClient *moduleLib.Client

// NewRouter creates a router that registers the various urls that are supported

func NewRouter(projectClient moduleLib.ProjectManager,
	compositeAppClient moduleLib.CompositeAppManager,
	genericPlacementIntentClient moduleLib.GenericPlacementIntentManager,
	appIntentClient moduleLib.AppIntentManager) *mux.Router {

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	moduleClient = moduleLib.NewClient()

	//setting routes for project
	if projectClient == nil {
		projectClient = moduleClient.Project

	}
	projHandler := projectHandler{
		client: projectClient,
	}
	router.HandleFunc("/projects", projHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}", projHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}", projHandler.deleteHandler).Methods("DELETE")

	//setting routes for compositeApp
	if compositeAppClient == nil {
		compositeAppClient = moduleClient.CompositeApp
	}
	compAppHandler := compositeAppHandler{
		client: compositeAppClient,
	}
	router.HandleFunc("/projects/{project-name}/composite-apps", compAppHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}", compAppHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}", compAppHandler.deleteHandler).Methods("DELETE")

	//setting routes for genericPlacementIntent
	if genericPlacementIntentClient == nil {
		genericPlacementIntentClient = moduleClient.GenericPlacementIntent
	}

	genericPlacementIntentHandler := genericPlacementIntentHandler{
		client: genericPlacementIntentClient,
	}
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents", genericPlacementIntentHandler.createGenericPlacementIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}", genericPlacementIntentHandler.getGenericPlacementHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}", genericPlacementIntentHandler.deleteGenericPlacementHandler).Methods("DELETE")

	//setting routes for AppIntent
	if appIntentClient == nil {
		appIntentClient = moduleClient.AppIntent
	}

	appIntentHandler := appIntentHandler{
		client: appIntentClient,
	}

	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}/app-intents", appIntentHandler.createAppIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}/app-intents/{app-intent-name}", appIntentHandler.getAppIntentHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}/app-intents/{app-intent-name}", appIntentHandler.deleteAppIntentHandler).Methods("DELETE")

	return router
}
