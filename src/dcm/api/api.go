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

    "github.com/onap/multicloud-k8s/src/dcm/internal/logicalcloud"

    "github.com/gorilla/mux"
)

// NewRouter creates a router that registers the various urls that are
// supported

func NewRouter(
    logicalCloudClient logicalcloud.LogicalCloudManager,
    clusterClient logicalcloud.ClusterManager,
    userPermissionClient logicalcloud.UserPermissionManager,
    quotaClient logicalcloud.QuotaManager,
    keyValueClient logicalcloud.KeyValueManager) *mux.Router {

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
        "/logical-clouds",
        logicalCloudHandler.getHandler).Methods("GET")
    lcRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}",
        logicalCloudHandler.getHandler).Methods("GET")
    lcRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}",
        logicalCloudHandler.deleteHandler).Methods("DELETE")
    lcRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}",
        logicalCloudHandler.updateHandler).Methods("PUT")
    lcRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/apply",
        logicalCloudHandler.applyHandler).Methods("POST")
    // To Do
    // get kubeconfig
    /*lcRouter.HandleFunc(
         "/logical-clouds/{name}/kubeconfig?cluster-reference={cluster}",
         logicalCloudHandler.getConfigHandler).Methods("GET")
    //get status
    lcRouter.HandleFunc(
        "/logical-clouds/{name}/cluster-references/",
        logicalCloudHandler.associateHandler).Methods("GET")*/

    // Set up Cluster API
    if clusterClient == nil {
        clusterClient = logicalcloud.NewClusterClient()
    }
    clusterHandler := clusterHandler{client: clusterClient}
    clusterRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-providers/{cluster-provider-name}/cluster-references",
        clusterHandler.createHandler).Methods("POST")
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-providers/{cluster-provider-name}/cluster-references",
        clusterHandler.getHandler).Methods("GET")
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-providers/{cluster-provider-name}/cluster-references/{cluster-name}",
        clusterHandler.getHandler).Methods("GET")
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-providers/{cluster-provider-name}/cluster-references/{cluster-name}",
        clusterHandler.updateHandler).Methods("PUT")
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-providers/{cluster-provider-name}/cluster-references/{cluster-name}",
        clusterHandler.deleteHandler).Methods("DELETE")

    // Set up User Permission API
    if userPermissionClient == nil {
        userPermissionClient = logicalcloud.NewUserPermissionClient()
    }
    userPermissionHandler := userPermissionHandler{client: userPermissionClient}
    upRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    upRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/user-permissions",
        userPermissionHandler.createHandler).Methods("POST")
    upRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/user-permissions/{permission-name}",
        userPermissionHandler.getHandler).Methods("GET")
    upRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/user-permissions/{permission-name}",
        userPermissionHandler.updateHandler).Methods("PUT")
    upRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/user-permissions/{permission-name}",
        userPermissionHandler.deleteHandler).Methods("DELETE")

    // Set up Quota API
    if quotaClient == nil {
        quotaClient = logicalcloud.NewQuotaClient()
    }
    quotaHandler := quotaHandler{client: quotaClient}
    quotaRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    quotaRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-quotas",
        quotaHandler.createHandler).Methods("POST")
    quotaRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-quotas/{quota-name}",
        quotaHandler.getHandler).Methods("GET")
    quotaRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-quotas/{quota-name}",
        quotaHandler.updateHandler).Methods("PUT")
    quotaRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-quotas/{quota-name}",
        quotaHandler.deleteHandler).Methods("DELETE")

    // Set up Key Value API
    if keyValueClient == nil {
        keyValueClient = logicalcloud.NewKeyValueClient()
    }
    keyValueHandler := keyValueHandler{client: keyValueClient}
    kvRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    kvRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/kv-pairs",
        keyValueHandler.createHandler).Methods("POST")
    kvRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/kv-pairs/{kv-pair-name}",
        keyValueHandler.getHandler).Methods("GET")
    kvRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/kv-pairs/{kv-pair-name}",
        keyValueHandler.updateHandler).Methods("PUT")
    kvRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/kv-pairs/{kv-pair-name}",
        keyValueHandler.deleteHandler).Methods("DELETE")
        return router
}
