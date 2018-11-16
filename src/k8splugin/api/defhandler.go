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

	"k8splugin/resource"

	"github.com/gorilla/mux"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type bundleDefHandler struct {
	// Interface that implements bundle Definition operations
	// We will set this variable with a mock interface for testing
	client resource.BundleDefInterface
}

// createHandler handles creation of the definition entry in the database
func (h bundleDefHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var v resource.BundleDefinition

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
	if v.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.Create(v)
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
// Note: This will be implemented in a different patch
func (h bundleDefHandler) uploadHandler(w http.ResponseWriter, r *http.Request) {
}

// listHandler handles GET (list) operations on the endpoint
// Returns a list of resource.BundleDefinitions
func (h bundleDefHandler) listHandler(w http.ResponseWriter, r *http.Request) {
	ret, err := h.client.List()
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

// getHandler handles GET operations on a particular ids
// Returns a resource.BundleDefinition
func (h bundleDefHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["resID"]

	ret, err := h.client.Get(id)
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

// deleteHandler handles DELETE operations on a particular bundle id
func (h bundleDefHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["resID"]

	err := h.client.Delete(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
