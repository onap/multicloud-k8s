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

var workloadIntJSONFile string = "json-schemas/network-workload.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type workloadintentHandler struct {
	// Interface that implements workload intent operations
	// We will set this variable with a mock interface for testing
	client moduleLib.WorkloadIntentManager
}

// Check for valid format of input parameters
func validateWorkloadIntentInputs(wi moduleLib.WorkloadIntent) error {
	// validate metadata
	err := moduleLib.IsValidMetadata(wi.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid network controller intent metadata")
	}

	errs := validation.IsValidName(wi.Spec.AppName)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid application name = [%v], errors: %v", wi.Spec.AppName, errs)
	}

	errs = validation.IsValidName(wi.Spec.WorkloadResource)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid workload resource = [%v], errors: %v", wi.Spec.WorkloadResource, errs)
	}

	errs = validation.IsValidName(wi.Spec.Type)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid workload type = [%v], errors: %v", wi.Spec.Type, errs)
	}
	return nil
}

// Create handles creation of the Network entry in the database
func (h workloadintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var wi moduleLib.WorkloadIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&wi)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(workloadIntJSONFile, wi)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if wi.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateWorkloadIntentInputs(wi)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateWorkloadIntent(wi, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, false)
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

// Put handles creation/update of the Network entry in the database
func (h workloadintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var wi moduleLib.WorkloadIntent
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&wi)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if wi.Metadata.Name == "" {
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if wi.Metadata.Name != name {
		fmt.Printf("bodyname = %v, name= %v\n", wi.Metadata.Name, name)
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateWorkloadIntentInputs(wi)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateWorkloadIntent(wi, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, true)
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

// Get handles GET operations on a particular Network Name
// Returns a Network
func (h workloadintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetWorkloadIntents(project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetWorkloadIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
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

// Delete handles DELETE operations on a particular Network  Name
func (h workloadintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := h.client.DeleteWorkloadIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
