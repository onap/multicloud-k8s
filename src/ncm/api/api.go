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
	"fmt"
	"reflect"

	"github.com/gorilla/mux"
	moduleLib "github.com/onap/multicloud-k8s/src/ncm/pkg/module"
)

var moduleClient *moduleLib.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *moduleLib.ClusterClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.ClusterManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.ClusterManager)
			if ok {
				return c
			}
		}
	case *moduleLib.NetworkClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.NetworkManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.NetworkManager)
			if ok {
				return c
			}
		}
	case *moduleLib.ProviderNetClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.ProviderNetManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.ProviderNetManager)
			if ok {
				return c
			}
		}
	case *moduleLib.NetControlIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.NetControlIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.NetControlIntentManager)
			if ok {
				return c
			}
		}
	case *moduleLib.WorkloadIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.WorkloadIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.WorkloadIntentManager)
			if ok {
				return c
			}
		}
	case *moduleLib.WorkloadIfIntentClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*moduleLib.WorkloadIfIntentManager)(nil)).Elem()) {
			c, ok := testClient.(moduleLib.WorkloadIfIntentManager)
			if ok {
				return c
			}
		}
	default:
		fmt.Printf("unknown type %T\n", cl)
	}
	return client
}

// NewRouter creates a router that registers the various urls that are supported
// testClient parameter allows unit testing for a given client
func NewRouter(testClient interface{}) *mux.Router {

	moduleClient = moduleLib.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	clusterHandler := clusterHandler{
		client: setClient(moduleClient.Cluster, testClient).(moduleLib.ClusterManager),
	}
	router.HandleFunc("/cluster-providers", clusterHandler.createClusterProviderHandler).Methods("POST")
	router.HandleFunc("/cluster-providers", clusterHandler.getClusterProviderHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{name}", clusterHandler.getClusterProviderHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{name}", clusterHandler.deleteClusterProviderHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters", clusterHandler.createClusterHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters", clusterHandler.getClusterHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters", clusterHandler.getClusterHandler).Queries("label", "{label}")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{name}", clusterHandler.getClusterHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{name}", clusterHandler.deleteClusterHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{name}/apply", clusterHandler.applyClusterHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{name}/terminate", clusterHandler.terminateClusterHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels", clusterHandler.createClusterLabelHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels", clusterHandler.getClusterLabelHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels/{label}", clusterHandler.getClusterLabelHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/labels/{label}", clusterHandler.deleteClusterLabelHandler).Methods("DELETE")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs", clusterHandler.createClusterKvPairsHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs", clusterHandler.getClusterKvPairsHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs/{kvpair}", clusterHandler.getClusterKvPairsHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/kv-pairs/{kvpair}", clusterHandler.deleteClusterKvPairsHandler).Methods("DELETE")

	networkHandler := networkHandler{
		client: setClient(moduleClient.Network, testClient).(moduleLib.NetworkManager),
	}
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks", networkHandler.createNetworkHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks", networkHandler.getNetworkHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks/{name}", networkHandler.putNetworkHandler).Methods("PUT")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks/{name}", networkHandler.getNetworkHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks/{name}", networkHandler.deleteNetworkHandler).Methods("DELETE")

	providernetHandler := providernetHandler{
		client: setClient(moduleClient.ProviderNet, testClient).(moduleLib.ProviderNetManager),
	}
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks", providernetHandler.createProviderNetHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks", providernetHandler.getProviderNetHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks/{name}", providernetHandler.putProviderNetHandler).Methods("PUT")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks/{name}", providernetHandler.getProviderNetHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks/{name}", providernetHandler.deleteProviderNetHandler).Methods("DELETE")

	netcontrolintentHandler := netcontrolintentHandler{
		client: setClient(moduleClient.NetControlIntent, testClient).(moduleLib.NetControlIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent", netcontrolintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent", netcontrolintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{name}", netcontrolintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{name}", netcontrolintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{name}", netcontrolintentHandler.deleteHandler).Methods("DELETE")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{name}/apply", netcontrolintentHandler.applyHandler).Methods("POST")

	workloadintentHandler := workloadintentHandler{
		client: setClient(moduleClient.WorkloadIntent, testClient).(moduleLib.WorkloadIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents", workloadintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents", workloadintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents/{name}", workloadintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents/{name}", workloadintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents/{name}", workloadintentHandler.deleteHandler).Methods("DELETE")

	workloadifintentHandler := workloadifintentHandler{
		client: setClient(moduleClient.WorkloadIfIntent, testClient).(moduleLib.WorkloadIfIntentManager),
	}
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces", workloadifintentHandler.createHandler).Methods("POST")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces", workloadifintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces/{name}", workloadifintentHandler.putHandler).Methods("PUT")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces/{name}", workloadifintentHandler.getHandler).Methods("GET")
	router.HandleFunc("/projects/{project}/composite-apps/{composite-app-name}/{version}/network-controller-intent/{net-control-intent}/workload-intents/{workload-intent}/interfaces/{name}", workloadifintentHandler.deleteHandler).Methods("DELETE")

	return router
}
