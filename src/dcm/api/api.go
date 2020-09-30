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
	"github.com/onap/multicloud-k8s/src/dcm/pkg/module"

	"github.com/gorilla/mux"
)

// NewRouter creates a router that registers the various urls that are
// supported
func NewRouter(
	logicalCloudClient module.LogicalCloudManager,
	clusterClient module.ClusterManager,
	userPermissionClient module.UserPermissionManager,
	quotaClient module.QuotaManager,
	keyValueClient module.KeyValueManager) *mux.Router {

	router := mux.NewRouter()

	// Set up Logical Cloud handler routes
	if logicalCloudClient == nil {
		logicalCloudClient = module.NewLogicalCloudClient()
	}

	if clusterClient == nil {
		clusterClient = module.NewClusterClient()
	}

	if quotaClient == nil {
		quotaClient = module.NewQuotaClient()
	}

	// Set up Logical Cloud API
	logicalCloudHandler := logicalCloudHandler{client: logicalCloudClient,
		clusterClient: clusterClient,
		quotaClient:   quotaClient,
	}
	lcRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
	lcRouter.HandleFunc(
		"/logical-clouds",
		logicalCloudHandler.createHandler).Methods("POST")
	lcRouter.HandleFunc(
		"/logical-clouds",
		logicalCloudHandler.getAllHandler).Methods("GET")
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
	lcRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/terminate",
		logicalCloudHandler.terminateHandler).Methods("POST")

	// Set up Cluster API
	clusterHandler := clusterHandler{client: clusterClient}
	clusterRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
	clusterRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/cluster-references",
		clusterHandler.createHandler).Methods("POST")
	clusterRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/cluster-references",
		clusterHandler.getAllHandler).Methods("GET")
	clusterRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/cluster-references/{cluster-reference}",
		clusterHandler.getHandler).Methods("GET")
	clusterRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/cluster-references/{cluster-reference}",
		clusterHandler.updateHandler).Methods("PUT")
	clusterRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/cluster-references/{cluster-reference}",
		clusterHandler.deleteHandler).Methods("DELETE")
	// Get kubeconfig for cluster of logical cloud
	clusterRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/cluster-references/{cluster-reference}/kubeconfig",
		clusterHandler.getConfigHandler).Methods("GET")

	// Set up User Permission API
	if userPermissionClient == nil {
		userPermissionClient = module.NewUserPermissionClient()
	}
	userPermissionHandler := userPermissionHandler{client: userPermissionClient}
	upRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
	upRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/user-permissions",
		userPermissionHandler.createHandler).Methods("POST")
	upRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/user-permissions",
		userPermissionHandler.getAllHandler).Methods("GET")
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
	quotaHandler := quotaHandler{client: quotaClient}
	quotaRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
	quotaRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/cluster-quotas",
		quotaHandler.createHandler).Methods("POST")
	quotaRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/cluster-quotas",
		quotaHandler.getAllHandler).Methods("GET")
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
		keyValueClient = module.NewKeyValueClient()
	}
	keyValueHandler := keyValueHandler{client: keyValueClient}
	kvRouter := router.PathPrefix("/v2/projects/{project-name}").Subrouter()
	kvRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/kv-pairs",
		keyValueHandler.createHandler).Methods("POST")
	kvRouter.HandleFunc(
		"/logical-clouds/{logical-cloud-name}/kv-pairs",
		keyValueHandler.getAllHandler).Methods("GET")
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
