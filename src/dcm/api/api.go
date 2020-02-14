/*
Copyright 2020 Intel Corporation.
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
        "dcm/internal/logicalcloud"

        "github.com/gorilla/mux"
)

// NewRouter creates a router that registers the various urls that are
// supported

func NewRouter(logicalCloudClient logicalcloud.LogicalCloudManager) *mux.Router {
    router := mux.NewRouter()

    // Set up Logical Cloud handler routes
    if logicalCloudClient == nil {
        logicalCloudClient = logicalcloud.NewLogicalCloudClient()
    }
    logicalCloudHandler := logicalCloudHandler{client: logicalCloudClient}
    lcRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    lcRouter.HandleFunc(
        "/logical-clouds",
        logicalCloudHandler.createHandler).Methods("POST")
    lcRouter.HandleFunc(
        "/logical-clouds/{name}",
        logicalCloudHandler.getHandler).Methods("GET")
    lcRouter.HandleFunc(
        "/logical-clouds/{name}",
        logicalCloudHandler.deleteHandler).Methods("DELETE")
    /*lcRouter.HandleFunc(
        "/logical-clouds/{name}/kubeconfig?cluster-reference={cluster}",
        logicalCloudHandler.getConfigHandler).Methods("GET")*/
    lcRouter.HandleFunc(
        "/logical-clouds/{name}",
        logicalCloudHandler.updateHandler).Methods("PUT")
    lcRouter.HandleFunc(
        "/logical-clouds/{name}/cluster-references/",
        logicalCloudHandler.associateHandler).Methods("POST")

        return router
}
