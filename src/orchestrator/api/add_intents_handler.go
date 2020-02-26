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
	"io"
	"net/http"

	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)

type intentHandler struct {
	client moduleLib.IntentManager
}

func (h intentHandler) addIntentHandler(w http.ResponseWriter, r *http.Request) {
	var i moduleLib.Intent

	err := json.NewDecoder(r.Body).Decode(&i)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return

	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if i.MetaData.Name == "" {
		http.Error(w, "Missing Intent in POST request", http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	d := vars["deployment-intent-group-name"]

	intent, addError := h.client.AddIntent(i, p, ca, v, d)
	if addError != nil {
		http.Error(w, addError.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h intentHandler) getIntentHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	i := vars["intent-name"]
	if i == "" {
		http.Error(w, "Missing intentName in GET request", http.StatusBadRequest)
		return
	}

	p := vars["project-name"]
	if p == "" {
		http.Error(w, "Missing projectName in GET request", http.StatusBadRequest)
		return
	}
	ca := vars["composite-app-name"]
	if ca == "" {
		http.Error(w, "Missing compositeAppName in GET request", http.StatusBadRequest)
		return
	}

	v := vars["composite-app-version"]
	if v == "" {
		http.Error(w, "Missing version of compositeApp in GET request", http.StatusBadRequest)
		return
	}

	di := vars["deployment-intent-group-name"]
	if di == "" {
		http.Error(w, "Missing name of DeploymentIntentGroup in GET request", http.StatusBadRequest)
		return
	}

	intent, err := h.client.GetIntent(i, p, ca, v, di)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(intent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h intentHandler) deleteIntentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	i := vars["intent-name"]
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	err := h.client.DeleteIntent(i, p, ca, v, di)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
