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
        "dcm/internal/logicalcloud/keyvalue"
        "dcm/internal/logicalcloud"
        "dcm/internal/logicalcloud/cluster" 
        "dcm/internal/logicalcloud/userpermission"
        "dcm/internal/logicalcloud/quota"


        "github.com/gorilla/mux"
)

// NewRouter creates a router that registers the various urls that are
// supported

func NewRouter(
    logicalCloudClient logicalcloud.LogicalCloudManager,
    clusterClient cluster.ClusterManager,
    userPermissionClient userpermission.UserPermissionManager,
    quotaClient quota.QuotaManager,
    keyValueClient keyvalue.KeyValueManager) *mux.Router {

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

    lcRouter.HandleFunc(
        "/logical-clouds/{name}",
        logicalCloudHandler.updateHandler).Methods("PUT")
    // To Do
    // get kubeconfig
    /*lcRouter.HandleFunc(
         "/logical-clouds/{name}/kubeconfig?cluster-reference={cluster}",
         logicalCloudHandler.getConfigHandler).Methods("GET")
    // apply
    lcRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/apply",
        logicalCloudHandler.getConfigHandler).Methods("PATCH")
    //get status    
    lcRouter.HandleFunc(
        "/logical-clouds/{name}/cluster-references/",
        logicalCloudHandler.associateHandler).Methods("GET")*/

    // Set up Cluster API
    if clusterClient == nil {
        clusterClient = cluster.NewClusterClient()
    }
    clusterHandler := clusterHandler{client: clusterClient}
    clusterRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-references",
        clusterHandler.createHandler).Methods("POST")

    // To Do - this should list all the cluster in the logical cloud        
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-references/{name}",
        clusterHandler.getHandler).Methods("GET")
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-references/{name}",
        clusterHandler.updateHandler).Methods("PUT")
    clusterRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-references/{name}",
        clusterHandler.deleteHandler).Methods("DELETE")

    // Set up User Permission API
    if userPermissionClient == nil {
        userPermissionClient = userpermission.NewUserPermissionClient()
    }
    userPermissionHandler := userPermissionHandler{client: userPermissionClient}
    upRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    upRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/user-permissions",
        userPermissionHandler.createHandler).Methods("POST")
    upRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/user-permissions/{name}",
        userPermissionHandler.getHandler).Methods("GET")
    upRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/user-permissions/{name}",
        userPermissionHandler.updateHandler).Methods("PUT")
    upRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/user-permissions/{name}",
        userPermissionHandler.deleteHandler).Methods("DELETE")

    // Set up Quota API
    if quotaClient == nil {
        quotaClient = quota.NewQuotaClient()
    }
    quotaHandler := quotaHandler{client: quotaClient}
    quotaRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    quotaRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-quotas",
        quotaHandler.createHandler).Methods("POST")
    quotaRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-quotas/{name}",
        quotaHandler.getHandler).Methods("GET")
    quotaRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-quotas/{name}",
        quotaHandler.updateHandler).Methods("PUT")
    quotaRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/cluster-quotas/{name}",
        quotaHandler.deleteHandler).Methods("DELETE")

    // Ser up Key Value API
    if keyValueClient == nil {
        keyValueClient = keyvalue.NewKeyValueClient()
    }
    keyValueHandler := keyValueHandler{client: keyValueClient}
    kvRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
    kvRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/kv-pairs",
        keyValueHandler.createHandler).Methods("POST")
    kvRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/kv-pairs/{name}",
        keyValueHandler.getHandler).Methods("GET")
    kvRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/kv-pairs/{name}",
        keyValueHandler.updateHandler).Methods("PUT")
    kvRouter.HandleFunc(
        "/logical-clouds/{logical-cloud-name}/kv-pairs/{name}",
        keyValueHandler.deleteHandler).Methods("DELETE")
        return router
}
