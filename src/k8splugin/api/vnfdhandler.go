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

	"github.com/gorilla/mux"
	"k8splugin/vnfd"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type vnfdHandler struct {
	// Interface that implements vnfDefinition operations
	// We will set this variable with a mock interface for testing
	vnfdImpl vnfd.VNFDefinitionInterface
}

// vnfdCreateHandler handles creation of the vnfd entry in the database
func (h vnfdHandler) vnfdCreateHandler(w http.ResponseWriter, r *http.Request) {
	var v vnfd.VNFDefinition

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

	ret, err := h.vnfdImpl.Create(v)
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

// vnfdUploadHandler handles upload of the vnf tar file into the database
// Note: This will be implemented in a different patch
func (h vnfdHandler) vnfdUploadHandler(w http.ResponseWriter, r *http.Request) {
}

// vnfdListHandler handles GET (list) operations on the /v1/vnfd endpoint
// Returns a list of vnfd.VNFDefinitions
func (h vnfdHandler) vnfdListHandler(w http.ResponseWriter, r *http.Request) {

	ret, err := h.vnfdImpl.List()
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

// vnfdGetHandler handles GET operations on a particular VNFID
// Returns a vnfd.VNFDefinition
func (h vnfdHandler) vnfdGetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	vnfdID := vars["vnfdID"]

	ret, err := h.vnfdImpl.Get(vnfdID)
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

// vnfdDeleteHandler handles DELETE operations on a particular VNFID
func (h vnfdHandler) vnfdDeleteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	vnfdID := vars["vnfdID"]

	err := h.vnfdImpl.Delete(vnfdID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
