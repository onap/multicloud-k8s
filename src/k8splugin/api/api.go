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
func NewRouter(kubeconfig string) *mux.Router {
	router := mux.NewRouter()

	vnfInstanceHandler := router.PathPrefix("/v1/vnf_instances").Subrouter()
	vnfInstanceHandler.HandleFunc("/", CreateHandler).Methods("POST").Name("VNFCreation")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}", ListHandler).Methods("GET")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}/{externalVNFID}", DeleteHandler).Methods("DELETE")
	vnfInstanceHandler.HandleFunc("/{cloudRegionID}/{namespace}/{externalVNFID}", GetHandler).Methods("GET")

	//rbd is resource bundle definition
	resRouter := router.PathPrefix("/v1/rb").Subrouter()
	rbdef := rbDefinitionHandler{client: rb.NewDefinitionClient()}
	resRouter.HandleFunc("/definition", rbdef.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/content", rbdef.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}", rbdef.listVersionsHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}", rbdef.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}", rbdef.deleteHandler).Methods("DELETE")

	//rbp is resource bundle profile
	rbprofile := rbProfileHandler{client: rb.NewProfileClient()}
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile", rbprofile.createHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}/content", rbprofile.uploadHandler).Methods("POST")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}", rbprofile.getHandler).Methods("GET")
	resRouter.HandleFunc("/definition/{rbname}/{rbversion}/profile/{prname}", rbprofile.deleteHandler).Methods("DELETE")

	// (TODO): Fix update method
	// vnfInstanceHandler.HandleFunc("/{vnfInstanceId}", UpdateHandler).Methods("PUT")

	return router
}
