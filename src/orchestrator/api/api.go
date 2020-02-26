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
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)

var moduleClient *moduleLib.Client

// NewRouter creates a router that registers the various urls that are supported
func NewRouter(projectClient moduleLib.ProjectManager,
	compositeAppClient moduleLib.CompositeAppManager,
	ControllerClient moduleLib.ControllerManager,
	clusterClient moduleLib.ClusterManager,
	deploymentIntentGrpClient moduleLib.DeploymentIntentGroupManager,
	intentClient moduleLib.IntentManager) *mux.Router {

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()
	moduleClient = moduleLib.NewClient()
	if projectClient == nil {
		projectClient = moduleClient.Project
	}
	projHandler := projectHandler{
		client: projectClient,
	}
	if ControllerClient == nil {
		ControllerClient = moduleClient.Controller
	}
	controlHandler := controllerHandler{
		client: ControllerClient,
	}
	if clusterClient == nil {
		clusterClient = moduleClient.Cluster
	}
	clusterHandler := clusterHandler{
		client: clusterClient,
	}
	router.HandleFunc("/projects", projHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}", projHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}", projHandler.deleteHandler).Methods("DELETE")

	if compositeAppClient == nil {
		compositeAppClient = moduleClient.CompositeApp
	}
	compAppHandler := compositeAppHandler{
		client: compositeAppClient,
	}

	router.HandleFunc("/projects/{project-name}/composite-apps", compAppHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}", compAppHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}", compAppHandler.deleteHandler).Methods("DELETE")

	router.HandleFunc("/controllers", controlHandler.createHandler).Methods("POST")
	router.HandleFunc("/controllers", controlHandler.createHandler).Methods("PUT")
	router.HandleFunc("/controllers/{controller-name}", controlHandler.getHandler).Methods("GET")
	router.HandleFunc("/controllers/{controller-name}", controlHandler.deleteHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers", clusterHandler.createClusterProviderHandler).Methods("POST")
	router.HandleFunc("/cluster-providers", clusterHandler.getClusterProviderHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{name}", clusterHandler.getClusterProviderHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{name}", clusterHandler.deleteClusterProviderHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters", clusterHandler.createClusterHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters", clusterHandler.getClusterHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{name}", clusterHandler.getClusterHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{name}", clusterHandler.deleteClusterHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels", clusterHandler.createClusterLabelHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels", clusterHandler.getClusterLabelHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels/{label}", clusterHandler.getClusterLabelHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels/{label}", clusterHandler.deleteClusterLabelHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs", clusterHandler.createClusterKvPairsHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs", clusterHandler.getClusterKvPairsHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs/{kvpair}", clusterHandler.getClusterKvPairsHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs/{kvpair}", clusterHandler.deleteClusterKvPairsHandler).Methods("DELETE")

	//setting routes for deploymentIntentGroup
	if deploymentIntentGrpClient == nil {
		deploymentIntentGrpClient = moduleClient.DeploymentIntentGroup
	}

	deploymentIntentGrpHandler := deploymentIntentGroupHandler{
		client: deploymentIntentGrpClient,
	}
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups", deploymentIntentGrpHandler.createDeploymentIntentGroupHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}", deploymentIntentGrpHandler.getDeploymentIntentGroupHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}", deploymentIntentGrpHandler.deleteDeploymentIntentGroupHandler).Methods("DELETE")

	// setting routes for AddingIntents
	if intentClient == nil {
		intentClient = moduleClient.Intent
	}

	intentHandler := intentHandler{
		client: intentClient,
	}

	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/intents",intentHandler.addIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/intents/{intent-name}",intentHandler.getIntentHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/intents/{intent-name}",intentHandler.deleteIntentHandler).Methods("DELETE")

	return router
}
