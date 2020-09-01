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
	"github.com/onap/multicloud-k8s/src/ncm/pkg/module"
	"github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents"
	"github.com/onap/multicloud-k8s/src/ncm/pkg/scheduler"
)

var moduleClient *module.Client

// For the given client and testClient, if the testClient is not null and
// implements the client manager interface corresponding to client, then
// return the testClient, otherwise return the client.
func setClient(client, testClient interface{}) interface{} {
	switch cl := client.(type) {
	case *networkintents.NetworkClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*networkintents.NetworkManager)(nil)).Elem()) {
			c, ok := testClient.(networkintents.NetworkManager)
			if ok {
				return c
			}
		}
	case *networkintents.ProviderNetClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*networkintents.ProviderNetManager)(nil)).Elem()) {
			c, ok := testClient.(networkintents.ProviderNetManager)
			if ok {
				return c
			}
		}
	case *scheduler.SchedulerClient:
		if testClient != nil && reflect.TypeOf(testClient).Implements(reflect.TypeOf((*scheduler.SchedulerManager)(nil)).Elem()) {
			c, ok := testClient.(scheduler.SchedulerManager)
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

	moduleClient = module.NewClient()

	router := mux.NewRouter().PathPrefix("/v2").Subrouter()

	networkHandler := networkHandler{
		client: setClient(moduleClient.Network, testClient).(networkintents.NetworkManager),
	}
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks", networkHandler.createNetworkHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks", networkHandler.getNetworkHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks/{name}", networkHandler.putNetworkHandler).Methods("PUT")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks/{name}", networkHandler.getNetworkHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/networks/{name}", networkHandler.deleteNetworkHandler).Methods("DELETE")

	providernetHandler := providernetHandler{
		client: setClient(moduleClient.ProviderNet, testClient).(networkintents.ProviderNetManager),
	}
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks", providernetHandler.createProviderNetHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks", providernetHandler.getProviderNetHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks/{name}", providernetHandler.putProviderNetHandler).Methods("PUT")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks/{name}", providernetHandler.getProviderNetHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{provider-name}/clusters/{cluster-name}/provider-networks/{name}", providernetHandler.deleteProviderNetHandler).Methods("DELETE")

	schedulerHandler := schedulerHandler{
		client: setClient(moduleClient.Scheduler, testClient).(scheduler.SchedulerManager),
	}
	router.HandleFunc("/cluster-providers/{cluster-provider}/clusters/{cluster}/apply", schedulerHandler.applySchedulerHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{cluster-provider}/clusters/{cluster}/terminate", schedulerHandler.terminateSchedulerHandler).Methods("POST")
	router.HandleFunc("/cluster-providers/{cluster-provider}/clusters/{cluster}/status", schedulerHandler.statusSchedulerHandler).Methods("GET")
	router.HandleFunc("/cluster-providers/{cluster-provider}/clusters/{cluster}/status",
		schedulerHandler.statusSchedulerHandler).Queries("instance", "{instance}", "type", "{type}", "output", "{output}", "app", "{app}", "cluster", "{cluster}", "resource", "{resource}")

	return router
}
