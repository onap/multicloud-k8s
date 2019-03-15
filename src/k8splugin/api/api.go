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
	"k8splugin/internal/rb"

	"github.com/gorilla/mux"
)

// NewRouter creates a router instance that serves the VNFInstance web methods
func NewRouter(kubeconfig string, defClient rb.DefinitionManager,
	profileClient rb.ProfileManager) *mux.Router {
	router := mux.NewRouter()

	vnfInstanceHandler := router.PathPrefix("/v1/vnf_instances").Subrouter()
	vnfInstanceHandler.HandleFunc("/", CreateHandler).Methods("POST").Name("VNFCreation")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}", ListHandler).Methods("GET")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}/{externalVNFID}", DeleteHandler).Methods("DELETE")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}/{externalVNFID}", GetHandler).Methods("GET")

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

	// (TODO): Fix update method
	// vnfInstanceHandler.HandleFunc("/{vnfInstanceId}", UpdateHandler).Methods("PUT")

	return router
}
