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
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"
)

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type instantiationHandler struct {
	client moduleLib.InstantiationManager
}

func (h instantiationHandler) instantiateHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	iErr := h.client.Instantiate(p, ca, v, di)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (h instantiationHandler) terminateHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	iErr := h.client.Terminate(p, ca, v, di)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (h instantiationHandler) statusHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	status, iErr := h.client.Status(p, ca, v, di)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	iErr = json.NewEncoder(w).Encode(status)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}
