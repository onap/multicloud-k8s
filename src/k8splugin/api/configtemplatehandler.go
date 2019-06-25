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
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"
	"net/http"

	"github.com/gorilla/mux"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type rbTemplateHandler struct {
	// Interface that implements bundle Definition operations
	// We will set this variable with a mock interface for testing
	client rb.ConfigTemplateManager
}

// createHandler handles creation of the template entry in the database
func (h rbTemplateHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var p rb.ConfigTemplate

	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if p.TemplateName == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = h.client.Create(rbName, rbVersion, p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}

// uploadHandler handles upload of the template tar file into the database
func (h rbTemplateHandler) uploadHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	templateName := vars["tname"]

	inpBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}

	if len(inpBytes) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	err = h.client.Upload(rbName, rbVersion, templateName, inpBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// getHandler handles GET operations on a particular template
func (h rbTemplateHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	templateName := vars["tname"]

	ret, err := h.client.Get(rbName, rbVersion, templateName)
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

// deleteHandler handles DELETE operations on a template
func (h rbTemplateHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rbName := vars["rbname"]
	rbVersion := vars["rbversion"]
	templateName := vars["tname"]

	err := h.client.Delete(rbName, rbVersion, templateName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
