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
	"io/ioutil"
	"k8splugin/internal/rb"
	"net/http"

	"github.com/gorilla/mux"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type rbProfileHandler struct {
	// Interface that implements bundle Definition operations
	// We will set this variable with a mock interface for testing
	client rb.ProfileManager
}

// createHandler handles creation of the definition entry in the database
func (h rbProfileHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]

	var p rb.Profile

	if r.Body == nil {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if p.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	if p.RBName != rbName || p.RBVersion != rbVersion {
		http.Error(w, "Mismatched parameters", http.StatusConflict)
		return
	}

	ret, err := h.client.Create(p)
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

// uploadHandler handles upload of the bundle tar file into the database
func (h rbProfileHandler) uploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]

	if r.Body == nil {
		http.Error(w, "Empty Body", http.StatusBadRequest)
		return
	}

	inpBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}

	err = h.client.Upload(rbName, rbVersion, prName, inpBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// getHandler handles GET operations on a particular ids
// Returns a rb.Definition
func (h rbProfileHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]

	ret, err := h.client.Get(rbName, rbVersion, prName)
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

// deleteHandler handles DELETE operations on a particular bundle definition id
func (h rbProfileHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	prName := vars["prname"]

	err := h.client.Delete(rbName, rbVersion, prName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
