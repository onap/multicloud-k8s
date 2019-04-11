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
	"k8splugin/internal/app"
	"k8splugin/internal/connectivity"
	"k8splugin/internal/rb"

	"github.com/gorilla/mux"
)

// NewRouter creates a router that registers the various urls that are supported
func NewRouter(defClient rb.DefinitionManager,
	profileClient rb.ProfileManager,
	instClient app.InstanceManager) *mux.Router {

	router := mux.NewRouter()

	// Setup Instance handler routes
	if instClient == nil {
		instClient = app.NewInstanceClient()
	}
	instHandler := instanceHandler{client: instClient}
	instRouter := router.PathPrefix("/v1").Subrouter()
	instRouter.HandleFunc("/instance", instHandler.createHandler).Methods("POST")
	instRouter.HandleFunc("/instance/{instID}", instHandler.getHandler).Methods("GET")
	instRouter.HandleFunc("/instance/{instID}", instHandler.deleteHandler).Methods("DELETE")
	// (TODO): Fix update method
	// instRouter.HandleFunc("/{vnfInstanceId}", UpdateHandler).Methods("PUT")

	//Setup the broker handler here
	brokerHandler := brokerInstanceHandler{client: instClient}
	instRouter.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload", brokerHandler.createHandler).Methods("POST")
	instRouter.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload/{instID}",
		brokerHandler.getHandler).Methods("GET")
	instRouter.HandleFunc("/{cloud-owner}/{cloud-region}/infra_workload/{instID}",
		brokerHandler.deleteHandler).Methods("DELETE")

	//Setup the connectivity api handler here
	connectivityClient := connectivity.NewConnectivityClient()
	connectivityHandler := connectivity.ConnectivityHandler{Client: connectivityClient}
	instRouter.HandleFunc("/connectivity-info", connectivityHandler.CreateHandler).Methods("POST")
	instRouter.HandleFunc("/connectivity-info/{name}", connectivityHandler.GetHandler).Methods("GET")
	instRouter.HandleFunc("/connectivity-info/{name}", connectivityHandler.DeleteHandler).Methods("DELETE")

	//Setup resource bundle definition routes
	if defClient == nil {
		defClient = rb.NewDefinitionClient()
	}
	defHandler := rbDefinitionHandler{client: defClient}
	resRouter := router.PathPrefix("/v1/rb").Subrouter()
	resRouter.HandleFunc("/definition", defHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/content", defHandler.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}", defHandler.listVersionsHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}", defHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}", defHandler.deleteHandler).Methods("DELETE")

	//Setup resource bundle profile routes
	if profileClient == nil {
		profileClient = rb.NewProfileClient()
	}
	profileHandler := rbProfileHandler{client: profileClient}
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile", profileHandler.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/content", profileHandler.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}", profileHandler.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}", profileHandler.deleteHandler).Methods("DELETE")

	return router
}
