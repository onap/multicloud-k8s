/*
 * Copyright 2019 Intel Corporation, Inc
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

var projectJSONFile string = "json-schemas/metadata.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type projectHandler struct {
	// Interface that implements Project operations
	// We will set this variable with a mock interface for testing
	client moduleLib.ProjectManager
}

// Create handles creation of the Project entry in the database
func (h projectHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var p moduleLib.Project

	err := json.NewDecoder(r.Body).Decode(&p)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(projectJSONFile, p)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateProject(p, false)
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

// Update handles updating the Project entry in the database
func (h projectHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["project-name"]

	var p moduleLib.Project

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
	if p.MetaData.Name == "" {
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if p.MetaData.Name != name {
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateProject(p, true)
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

// Get handles GET operations on a particular Project Name
// Returns a Project
func (h projectHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["project-name"]

	// handle for get all projects
	if len(name) == 0 {
		var pList []moduleLib.Project

		projects, err := h.client.GetAllProjects()
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		for _, p := range projects {
			pList = append(pList, moduleLib.Project{MetaData: p.MetaData})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(pList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return

	}

	ret, err := h.client.GetProject(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
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

// Delete handles DELETE operations on a particular Project Name
func (h projectHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["project-name"]

	_, err := h.client.GetProject(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	err = h.client.DeleteProject(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
