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

	"github.com/gorilla/mux"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"
)

var caJSONFile string = "json-schemas/composite-app.json"

// compositeAppHandler to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type compositeAppHandler struct {
	// Interface that implements CompositeApp operations
	// We will set this variable with a mock interface for testing
	client moduleLib.CompositeAppManager
}

// createHandler handles creation of the CompositeApp entry in the database
func (h compositeAppHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var c moduleLib.CompositeApp

	err := json.NewDecoder(r.Body).Decode(&c)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}
	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(caJSONFile, c)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["project-name"]

	ret, err := h.client.CreateCompositeApp(c, projectName)
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

// getHandler handles GET operations on a particular CompositeApp Name
// Returns a compositeApp
func (h compositeAppHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["composite-app-name"]
	version := vars["version"]
	projectName := vars["project-name"]

	ret, err := h.client.GetCompositeApp(name, version, projectName)
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

// getAllCompositeAppsHandler handles the GetAllComppositeApps, returns a list of compositeApps under a project
func (h compositeAppHandler) getAllCompositeAppsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pName := vars["project-name"]

	var caList []moduleLib.CompositeApp

	cApps, err := h.client.GetAllCompositeApps(pName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	for _, cApp := range cApps {
		caList = append(caList, moduleLib.CompositeApp{Metadata: cApp.Metadata, Spec: cApp.Spec})
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(caList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	return
}

// deleteHandler handles DELETE operations on a particular CompositeApp Name
func (h compositeAppHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["composite-app-name"]
	version := vars["version"]
	projectName := vars["project-name"]

	_, err := h.client.GetCompositeApp(name, version, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = h.client.DeleteCompositeApp(name, version, projectName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
