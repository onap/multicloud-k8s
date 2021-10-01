/*
Copyright 2018 Intel Corporation.
Copyright © 2021 Samsung Electronics
Copyright © 2021 Orange

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
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/healthcheck"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	"github.com/gorilla/mux"
)

// NewRouter creates a router that registers the various urls that are supported
func NewRouter(defClient rb.DefinitionManager,
	profileClient rb.ProfileManager,
	instClient app.InstanceManager,
	queryClient app.QueryManager,
	configClient app.ConfigManager,
	connectionClient connection.ConnectionManager,
	templateClient rb.ConfigTemplateManager,
	healthcheckClient healthcheck.InstanceHCManager) *mux.Router {

	router := mux.NewRouter()

	// Setup Instance handler routes
	if instClient == nil {
		instClient = app.NewInstanceClient()
	}
	instHandler := instanceHandler{client: instClient}
	instRouter := router.PathPrefix("/v1").Subrouter()
	instRouter.HandleFunc("/instance", instHandler.createHandler).Methods("POST")
	instRouter.HandleFunc("/instance", instHandler.listHandler).Methods("GET")
	// Match rb-names, versions or profiles
	instRouter.HandleFunc("/instance", instHandler.listHandler).
		Queries("rb-name", "{rb-name}",
			"rb-version", "{rb-version}",
			"profile-name", "{profile-name}").Methods("GET")
	//Want to get full Data -> add query param: /install/{instID}?full=true
	instRouter.HandleFunc("/instance/{instID}", instHandler.getHandler).Methods("GET")
	instRouter.HandleFunc("/instance/{instID}/status", instHandler.statusHandler).Methods("GET")
	instRouter.HandleFunc("/instance/{instID}/query", instHandler.queryHandler).Methods("GET")
	instRouter.HandleFunc("/instance/{instID}/query", instHandler.queryHandler).
		Queries("ApiVersion", "{ApiVersion}",
			"Kind", "{Kind}",
			"Name", "{Name}",
			"Labels", "{Labels}").Methods("GET")
	instRouter.HandleFunc("/instance/{instID}", instHandler.deleteHandler).Methods("DELETE")

	// Query handler routes
	if queryClient == nil {
		queryClient = app.NewQueryClient()
	}
	queryHandler := queryHandler{client: queryClient}
	queryRouter := router.PathPrefix("/v1").Subrouter()
	queryRouter.HandleFunc("/query", queryHandler.queryHandler).Methods("GET")
	queryRouter.HandleFunc("/query", queryHandler.queryHandler).
		Queries("Namespace", "{Namespace}",
			"CloudRegion", "{CloudRegion}",
			"ApiVersion", "{ApiVersion}",
			"Kind", "{Kind}",
			"Name", "{Name}",
			"Labels", "{Labels}").Methods("GET")

	//Setup the broker handler here
	//Use the base router without any path prefixes
	brokerHandler := brokerInstanceHandler{client: instClient}
	router.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload", brokerHandler.createHandler).Methods("POST")
	router.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload/{instID}", brokerHandler.getHandler).Methods("GET")
	router.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload", brokerHandler.findHandler).Queries("name", "{name}").Methods("GET")
	router.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload/{instID}", brokerHandler.deleteHandler).Methods("DELETE")

	//Setup the connectivity api handler here
	if connectionClient == nil {
		connectionClient = connection.NewConnectionClient()
	}
	connectionHandler := connectionHandler{client: connectionClient}
	instRouter.HandleFunc("/connectivity-info", connectionHandler.createHandler).Methods("POST")
	instRouter.HandleFunc("/connectivity-info/{connname}", connectionHandler.getHandler).Methods("GET")
	instRouter.HandleFunc("/connectivity-info/{connname}", connectionHandler.deleteHandler).Methods("DELETE")

	//Setup resource bundle definition routes
	if defClient == nil {
		defClient = rb.NewDefinitionClient()
	}
	defHandler := rbDefinitionHandler{client: defClient}
	resRouter := router.PathPrefix("/v1/rb").Subrouter()
	resRouter.HandleFunc("/definition", defHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/content", defHandler.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}", defHandler.listVersionsHandler).Methods("GET")
	resRouter.HandleFunc("/definition", defHandler.listAllHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}", defHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}", defHandler.deleteHandler).Methods("DELETE")

	//Setup resource bundle profile routes
	if profileClient == nil {
		profileClient = rb.NewProfileClient()
	}
	profileHandler := rbProfileHandler{client: profileClient}
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile", profileHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile", profileHandler.listHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/content", profileHandler.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}", profileHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}", profileHandler.deleteHandler).Methods("DELETE")

	// Config Template
	if templateClient == nil {
		templateClient = rb.NewConfigTemplateClient()
	}
	templateHandler := rbTemplateHandler{client: templateClient}
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template", templateHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template", templateHandler.listHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template/{tname}/content", templateHandler.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template/{tname}", templateHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template/{tname}", templateHandler.deleteHandler).Methods("DELETE")

	// Config value
	if configClient == nil {
		configClient = app.NewConfigClient()
	}
	configHandler := rbConfigHandler{client: configClient}
	instRouter.HandleFunc("/instance/{instID}/config", configHandler.createHandler).Methods("POST")
	instRouter.HandleFunc("/instance/{instID}/config", configHandler.listHandler).Methods("GET")
	instRouter.HandleFunc("/instance/{instID}/config/{cfgname}", configHandler.getHandler).Methods("GET")
	instRouter.HandleFunc("/instance/{instID}/config/{cfgname}", configHandler.updateHandler).Methods("PUT")
	instRouter.HandleFunc("/instance/{instID}/config/{cfgname}", configHandler.deleteAllHandler).Methods("DELETE")
	instRouter.HandleFunc("/instance/{instID}/config/{cfgname}/delete", configHandler.deleteHandler).Methods("POST")
	instRouter.HandleFunc("/instance/{instID}/config/{cfgname}/rollback", configHandler.rollbackHandler).Methods("POST")
	instRouter.HandleFunc("/instance/{instID}/config/{cfgname}/tagit", configHandler.tagitHandler).Methods("POST")

	// Instance Healthcheck API
	if healthcheckClient == nil {
		healthcheckClient = healthcheck.NewHCClient()
	}
	healthcheckHandler := instanceHCHandler{client: healthcheckClient}
	instRouter.HandleFunc("/instance/{instID}/healthcheck", healthcheckHandler.listHandler).Methods("GET")
	instRouter.HandleFunc("/instance/{instID}/healthcheck", healthcheckHandler.createHandler).Methods("POST")
	instRouter.HandleFunc("/instance/{instID}/healthcheck/{hcID}", healthcheckHandler.getHandler).Methods("GET")
	instRouter.HandleFunc("/instance/{instID}/healthcheck/{hcID}", healthcheckHandler.deleteHandler).Methods("DELETE")

	// Add healthcheck path
	instRouter.HandleFunc("/healthcheck", healthCheckHandler).Methods("GET")

	return router
}
