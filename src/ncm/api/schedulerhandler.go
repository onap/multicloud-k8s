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
	"net/http"

	"github.com/onap/multicloud-k8s/src/ncm/pkg/scheduler"

	"github.com/gorilla/mux"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type schedulerHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client scheduler.SchedulerManager
}

//  applyClusterHandler handles requests to apply network intents for a cluster
func (h schedulerHandler) applySchedulerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["cluster-provider"]
	cluster := vars["cluster"]

	err := h.client.ApplyNetworkIntents(provider, cluster)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

//  terminateSchedulerHandler handles requests to apply network intents for a cluster
func (h schedulerHandler) terminateSchedulerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["cluster-provider"]
	cluster := vars["cluster"]

	err := h.client.TerminateNetworkIntents(provider, cluster)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
