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

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)

var gpiJSONFile string = "json-schemas/generic-placement-intent.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type genericPlacementIntentHandler struct {
	client moduleLib.GenericPlacementIntentManager
}

// createGenericPlacementIntentHandler handles the create operation of intent
func (h genericPlacementIntentHandler) createGenericPlacementIntentHandler(w http.ResponseWriter, r *http.Request) {

	var g moduleLib.GenericPlacementIntent

	err := json.NewDecoder(r.Body).Decode(&g)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(gpiJSONFile, g)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	version := vars["composite-app-version"]
	digName := vars["deployment-intent-group-name"]

	gPIntent, createErr := h.client.CreateGenericPlacementIntent(g, projectName, compositeAppName, version, digName)
	if createErr != nil {
		http.Error(w, createErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(gPIntent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getGenericPlacementHandler handles the GET operations on intent
func (h genericPlacementIntentHandler) getGenericPlacementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	intentName := vars["intent-name"]
	if intentName == "" {
		http.Error(w, "Missing genericPlacementIntentName in GET request", http.StatusBadRequest)
		return
	}
	projectName := vars["project-name"]
	if projectName == "" {
		http.Error(w, "Missing projectName in GET request", http.StatusBadRequest)
		return
	}
	compositeAppName := vars["composite-app-name"]
	if compositeAppName == "" {
		http.Error(w, "Missing compositeAppName in GET request", http.StatusBadRequest)
		return
	}

	version := vars["composite-app-version"]
	if version == "" {
		http.Error(w, "Missing version in GET request", http.StatusBadRequest)
		return
	}

	dig := vars["deployment-intent-group-name"]
	if dig == "" {
		http.Error(w, "Missing deploymentIntentGroupName in GET request", http.StatusBadRequest)
		return
	}

	gPIntent, err := h.client.GetGenericPlacementIntent(intentName, projectName, compositeAppName, version, dig)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(gPIntent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h genericPlacementIntentHandler) getAllGenericPlacementIntentsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pList := []string{"project-name", "composite-app-name", "composite-app-version"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	digName := vars["deployment-intent-group-name"]

	gpList, err := h.client.GetAllGenericPlacementIntents(p, ca, v, digName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(gpList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

// deleteGenericPlacementHandler handles the delete operations on intent
func (h genericPlacementIntentHandler) deleteGenericPlacementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	i := vars["intent-name"]
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	digName := vars["deployment-intent-group-name"]

	err := h.client.DeleteGenericPlacementIntent(i, p, ca, v, digName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
