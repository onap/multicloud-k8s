/*
 * Copyright 2018 Intel Corporation, Inc
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

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"

	"github.com/gorilla/mux"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type rbConfigHandler struct {
	// Interface that implements bundle Definition operations
	// We will set this variable with a mock interface for testing
	client app.ConfigManager
}

// createHandler handles creation of the definition entry in the database
func (h rbConfigHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var p app.Config
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]

	if r.Body == nil {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if p.ConfigName == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.Create(rbName, rbVersion, prName, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles GET operations on a particular config
// Returns a app.Definition
func (h rbConfigHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]
	cfgName := vars["cfgname"]

	ret, err := h.client.Get(rbName, rbVersion, prName, cfgName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// deleteHandler handles DELETE operations on a config
func (h rbConfigHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]
	cfgName := vars["cfgname"]

	ret, err := h.client.Delete(rbName, rbVersion, prName, cfgName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

// UpdateHandler handles Update operations on a particular configuration
func (h rbConfigHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]
	cfgName := vars["cfgname"]

	var p app.Config

	if r.Body == nil {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	ret, err := h.client.Update(rbName, rbVersion, prName, cfgName, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// rollbackHandler handles Rollback operations to a specific version
func (h rbConfigHandler) rollbackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]

	if r.Body == nil {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	var p app.ConfigRollback
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	err = h.client.Rollback(rbName, rbVersion, prName, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// tagitHandler handles TAGIT operation
func (h rbConfigHandler) tagitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]

	if r.Body == nil {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	var p app.ConfigTagit
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = h.client.Tagit(rbName, rbVersion, prName, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
