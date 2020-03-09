/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"github.com/gorilla/mux"
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"
)

var moduleClient *moduleLib.Client

// NewRouter creates a router that registers the various urls that are supported
func NewRouter(projectClient moduleLib.ProjectManager,
	compositeAppClient moduleLib.CompositeAppManager,
	appClient moduleLib.AppManager,
	ControllerClient moduleLib.ControllerManager,
	clusterClient moduleLib.ClusterManager,
	genericPlacementIntentClient moduleLib.GenericPlacementIntentManager,
	appIntentClient moduleLib.AppIntentManager,
	deploymentIntentGrpClient moduleLib.DeploymentIntentGroupManager,
	intentClient moduleLib.IntentManager,
	compositeProfileClient moduleLib.CompositeProfileManager,
	appProfileClient moduleLib.AppProfileManager) *mux.Router {

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	moduleClient = moduleLib.NewClient()

	//setting routes for project
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

	//setting routes for compositeApp
	if compositeAppClient == nil {
		compositeAppClient = moduleClient.CompositeApp
	}
	compAppHandler := compositeAppHandler{
		client: compositeAppClient,
	}
	router.HandleFunc("/projects/{project-name}/composite-apps", compAppHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}", compAppHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}", compAppHandler.deleteHandler).Methods("DELETE")

	if appClient == nil {
		appClient = moduleClient.App
	}
	appHandler := appHandler{
		client: appClient,
	}

	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}/apps", appHandler.createAppHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}/apps/{app-name}", appHandler.getAppHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{version}/apps/{app-name}", appHandler.deleteAppHandler).Methods("DELETE")

	if compositeProfileClient == nil {
		compositeProfileClient = moduleClient.CompositeProfile
	}
	compProfilepHandler := compositeProfileHandler{
		client: compositeProfileClient,
	}
	if appProfileClient == nil {
		appProfileClient = moduleClient.AppProfile
	}
	appProfileHandler := appProfileHandler{
		client: appProfileClient,
	}

	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles", compProfilepHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles", compProfilepHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}", compProfilepHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}", compProfilepHandler.deleteHandler).Methods("DELETE")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles", appProfileHandler.createAppProfileHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles", appProfileHandler.getAppProfileHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles", appProfileHandler.getAppProfileHandler).Queries("app-name", "{app-name}")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles/{app-profile}", appProfileHandler.getAppProfileHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles/{app-profile}", appProfileHandler.deleteAppProfileHandler).Methods("DELETE")

	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles", appProfileHandler.createAppProfileHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles", appProfileHandler.getAppProfileHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles", appProfileHandler.getAppProfileHandler).Queries("app-name", "{app-name}")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles/{app-profile}", appProfileHandler.getAppProfileHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/composite-profiles/{composite-profile-name}/profiles/{app-profile}", appProfileHandler.deleteAppProfileHandler).Methods("DELETE")

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
	//setting routes for genericPlacementIntent
	if genericPlacementIntentClient == nil {
		genericPlacementIntentClient = moduleClient.GenericPlacementIntent
	}

	genericPlacementIntentHandler := genericPlacementIntentHandler{
		client: genericPlacementIntentClient,
	}
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents", genericPlacementIntentHandler.createGenericPlacementIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}", genericPlacementIntentHandler.getGenericPlacementHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}", genericPlacementIntentHandler.deleteGenericPlacementHandler).Methods("DELETE")

	//setting routes for AppIntent
	if appIntentClient == nil {
		appIntentClient = moduleClient.AppIntent
	}

	appIntentHandler := appIntentHandler{
		client: appIntentClient,
	}

	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}/app-intents", appIntentHandler.createAppIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}/app-intents/{app-intent-name}", appIntentHandler.getAppIntentHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/generic-placement-intents/{intent-name}/app-intents/{app-intent-name}", appIntentHandler.deleteAppIntentHandler).Methods("DELETE")

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

	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/intents", intentHandler.addIntentHandler).Methods("POST")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/intents/{intent-name}", intentHandler.getIntentHandler).Methods("GET")
	router.HandleFunc("/projects/{project-name}/composite-apps/{composite-app-name}/{composite-app-version}/deployment-intent-groups/{deployment-intent-group-name}/intents/{intent-name}", intentHandler.deleteIntentHandler).Methods("DELETE")

	return router
}
