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
	"io"
	"io/ioutil"
	"net/http"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	"github.com/gorilla/mux"
	logr "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
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
	var p rb.Profile

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		logr.WithFields("http.StatusBadRequest", "Error", "Empty body")
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		logr.WithFields("http.StatusUnprocessableEntity", "Error", "StatusUnprocessableEntity")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if p.ProfileName == "" {
		logr.WithFields("http.StatusBadRequest", "ProfileName", "Missing name in POST request")
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.Create(p)
	if err != nil {
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
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

	inpBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logr.WithFields("http.StatusBadRequest", "Error", "Unable to read body")
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}

	if len(inpBytes) == 0 {
		logr.WithFields("http.StatusBadRequest", "Error", "Empty body")
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	err = h.client.Upload(rbName, rbVersion, prName, inpBytes)
	if err != nil {
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
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
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles GET operations on a particular ids
// Returns a rb.Definition
func (h rbProfileHandler) listHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]

	ret, err := h.client.List(rbName, rbVersion)
	if err != nil {
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
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
		logr.WithFields("http.StatusInternalServerError", "Error", "StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
