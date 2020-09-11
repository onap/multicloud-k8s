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

var dpiJSONFile string = "json-schemas/deployment-group-intent.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type deploymentIntentGroupHandler struct {
	client moduleLib.DeploymentIntentGroupManager
}

// createDeploymentIntentGroupHandler handles the create operation of DeploymentIntentGroup
func (h deploymentIntentGroupHandler) createDeploymentIntentGroupHandler(w http.ResponseWriter, r *http.Request) {

	var d moduleLib.DeploymentIntentGroup

	err := json.NewDecoder(r.Body).Decode(&d)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(dpiJSONFile, d)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	version := vars["composite-app-version"]

	dIntent, createErr := h.client.CreateDeploymentIntentGroup(d, projectName, compositeAppName, version)
	if createErr != nil {
		http.Error(w, createErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(dIntent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h deploymentIntentGroupHandler) getDeploymentIntentGroupHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

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
	if v == "" {
		http.Error(w, "Missing name of DeploymentIntentGroup in GET request", http.StatusBadRequest)
		return
	}

	dIntentGrp, err := h.client.GetDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(dIntentGrp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h deploymentIntentGroupHandler) getAllDeploymentIntentGroupsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pList := []string{"project-name", "composite-app-name", "composite-app-version"}
	err := validation.IsValidParameterPresent(vars, pList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]

	diList, err := h.client.GetAllDeploymentIntentGroups(p, ca, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(diList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
}

func (h deploymentIntentGroupHandler) deleteDeploymentIntentGroupHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	err := h.client.DeleteDeploymentIntentGroup(di, p, ca, v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
