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
	"fmt"
	"io"
	"net/http"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/onap/multicloud-k8s/src/ovnaction/pkg/module"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

var netCntIntJSONFile string = "json-schemas/metadata.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type netcontrolintentHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client moduleLib.NetControlIntentManager
}

// Check for valid format of input parameters
func validateNetControlIntentInputs(nci moduleLib.NetControlIntent) error {
	// validate metadata
	err := moduleLib.IsValidMetadata(nci.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid network controller intent metadata")
	}
	return nil
}

// Create handles creation of the NetControlIntent entry in the database
func (h netcontrolintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var nci moduleLib.NetControlIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := json.NewDecoder(r.Body).Decode(&nci)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(netCntIntJSONFile, nci)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if nci.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateNetControlIntentInputs(nci)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetControlIntent(nci, project, compositeApp, compositeAppVersion, deployIntentGroup, false)
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

// Put handles creation/update of the NetControlIntent entry in the database
func (h netcontrolintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var nci moduleLib.NetControlIntent
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := json.NewDecoder(r.Body).Decode(&nci)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if nci.Metadata.Name == "" {
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if nci.Metadata.Name != name {
		fmt.Printf("bodyname = %v, name= %v\n", nci.Metadata.Name, name)
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateNetControlIntentInputs(nci)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetControlIntent(nci, project, compositeApp, compositeAppVersion, deployIntentGroup, true)
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

// Get handles GET operations on a particular NetControlIntent Name
// Returns a NetControlIntent
func (h netcontrolintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetNetControlIntents(project, compositeApp, compositeAppVersion, deployIntentGroup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetNetControlIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete handles DELETE operations on a particular NetControlIntent  Name
func (h netcontrolintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]

	err := h.client.DeleteNetControlIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
