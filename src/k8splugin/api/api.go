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
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	"github.com/gorilla/mux"
)

// Helper for returning all /v1/ API prefixes
// according to context
func yieldV1Prefixes() []string {
	config := config.GetConfiguration()
	if config.PreserveV1BackwardCompatibility == "true" {
		return []string{"", "/v1", "/v1/v1"}
	}
	return []string{"/v1"}

}

// Helper for initializing all handlers
func setupHandlers(
	defClient rb.DefinitionManager,
	profileClient rb.ProfileManager,
	instClient app.InstanceManager,
	configClient app.ConfigManager,
	connectionClient connection.ConnectionManager,
	templateClient rb.ConfigTemplateManager) (
	instanceHandler,
	brokerInstanceHandler,
	connectionHandler,
	rbDefinitionHandler,
	rbProfileHandler,
	rbTemplateHandler,
	rbConfigHandler) {

	if instClient == nil {
		instClient = app.NewInstanceClient()
	}
	instHandler := instanceHandler{client: instClient}
	brokerHandler := brokerInstanceHandler{client: instClient}
	if connectionClient == nil {
		connectionClient = connection.NewConnectionClient()
	}
	connectionHandler := connectionHandler{client: connectionClient}
	if defClient == nil {
		defClient = rb.NewDefinitionClient()
	}
	defHandler := rbDefinitionHandler{client: defClient}
	if profileClient == nil {
		profileClient = rb.NewProfileClient()
	}
	profileHandler := rbProfileHandler{client: profileClient}
	if templateClient == nil {
		templateClient = rb.NewConfigTemplateClient()
	}
	templateHandler := rbTemplateHandler{client: templateClient}
	if configClient == nil {
		configClient = app.NewConfigClient()
	}
	configHandler := rbConfigHandler{client: configClient}

	return instHandler, brokerHandler, connectionHandler, defHandler,
		profileHandler, templateHandler, configHandler
}

// NewRouter creates a router that registers the various urls that are supported
func NewRouter(defClient rb.DefinitionManager,
	profileClient rb.ProfileManager,
	instClient app.InstanceManager,
	configClient app.ConfigManager,
	connectionClient connection.ConnectionManager,
	templateClient rb.ConfigTemplateManager) *mux.Router {

	router := mux.NewRouter()

	//Retrieve handlers from wrapper
	instHandler, brokerHandler, connectionHandler,
		defHandler, profileHandler, templateHandler,
		configHandler := setupHandlers(defClient,
		profileClient, instClient, configClient,
		connectionClient, templateClient)

	for _, prefix := range yieldV1Prefixes() {
		subRouter := router.PathPrefix(prefix).Subrouter()
		subRouter.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload", brokerHandler.createHandler).Methods("POST")
		subRouter.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload/{instID}", brokerHandler.getHandler).Methods("GET")
		subRouter.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload", brokerHandler.findHandler).Queries("name", "{name}").Methods("GET")
		subRouter.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload/{instID}", brokerHandler.deleteHandler).Methods("DELETE")
	}
	subRouter := router.PathPrefix("/v1").Subrouter()

	subRouter.HandleFunc("/healthcheck", healthCheckHandler).Methods("GET")

	subRouter.HandleFunc("/instance", instHandler.createHandler).Methods("POST")
	subRouter.HandleFunc("/instance", instHandler.listHandler).Methods("GET")
	subRouter.HandleFunc("/instance", instHandler.listHandler).
		Queries("rb-name", "{rb-name}",
			"rb-version", "{rb-version}",
			"profile-name", "{profile-name}").Methods("GET")

	subRouter.HandleFunc("/instance/{instID}", instHandler.getHandler).Methods("GET")
	subRouter.HandleFunc("/instance/{instID}/status", instHandler.statusHandler).Methods("GET")
	subRouter.HandleFunc("/instance/{instID}", instHandler.deleteHandler).Methods("DELETE")
	// (TODO): Fix update method
	// subRouter.HandleFunc("/{vnfInstanceId}", UpdateHandler).Methods("PUT")

	subRouter.HandleFunc("/connectivity-info", connectionHandler.createHandler).Methods("POST")
	subRouter.HandleFunc("/connectivity-info/{connname}", connectionHandler.getHandler).Methods("GET")
	subRouter.HandleFunc("/connectivity-info/{connname}", connectionHandler.deleteHandler).Methods("DELETE")

	resRouter := subRouter.PathPrefix("/rb").Subrouter()
	resRouter.HandleFunc("/definition", defHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/content", defHandler.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}", defHandler.listVersionsHandler).Methods("GET")
	resRouter.HandleFunc("/definition", defHandler.listAllHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}", defHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}", defHandler.deleteHandler).Methods("DELETE")

	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile", profileHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile", profileHandler.listHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/content", profileHandler.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}", profileHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}", profileHandler.deleteHandler).Methods("DELETE")

	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template", templateHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template/{tname}/content", templateHandler.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template/{tname}", templateHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/config-template/{tname}", templateHandler.deleteHandler).Methods("DELETE")

	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/config", configHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/config/{cfgname}", configHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/config/{cfgname}", configHandler.updateHandler).Methods("PUT")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/config/{cfgname}", configHandler.deleteHandler).Methods("DELETE")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/config/rollback", configHandler.rollbackHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/config/tagit", configHandler.tagitHandler).Methods("POST")

	return router
}
