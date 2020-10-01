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

var netIfJSONFile string = "json-schemas/network-load-interface.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type workloadifintentHandler struct {
	// Interface that implements workload intent operations
	// We will set this variable with a mock interface for testing
	client moduleLib.WorkloadIfIntentManager
}

// Check for valid format of input parameters
func validateWorkloadIfIntentInputs(wif moduleLib.WorkloadIfIntent) error {
	// validate metadata
	err := moduleLib.IsValidMetadata(wif.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid network controller intent metadata")
	}

	errs := validation.IsValidName(wif.Spec.IfName)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid interface name = [%v], errors: %v", wif.Spec.IfName, errs)
	}

	errs = validation.IsValidName(wif.Spec.NetworkName)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid network name = [%v], errors: %v", wif.Spec.NetworkName, errs)
	}

	// optional - only validate if supplied
	if len(wif.Spec.DefaultGateway) > 0 {
		errs = validation.IsValidName(wif.Spec.DefaultGateway)
		if len(errs) > 0 {
			return pkgerrors.Errorf("Invalid default interface = [%v], errors: %v", wif.Spec.DefaultGateway, errs)
		}
	}

	// optional - only validate if supplied
	if len(wif.Spec.IpAddr) > 0 {
		err = validation.IsIp(wif.Spec.IpAddr)
		if err != nil {
			return pkgerrors.Errorf("Invalid IP address = [%v], errors: %v", wif.Spec.IpAddr, err)
		}
	}

	// optional - only validate if supplied
	if len(wif.Spec.MacAddr) > 0 {
		err = validation.IsMac(wif.Spec.MacAddr)
		if err != nil {
			return pkgerrors.Errorf("Invalid MAC address = [%v], errors: %v", wif.Spec.MacAddr, err)
		}
	}
	return nil
}

// Create handles creation of the Network entry in the database
func (h workloadifintentHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var wif moduleLib.WorkloadIfIntent
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	workloadIntent := vars["workload-intent"]

	err := json.NewDecoder(r.Body).Decode(&wif)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(netIfJSONFile, wif)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if wif.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	// set default value
	if len(wif.Spec.DefaultGateway) == 0 {
		wif.Spec.DefaultGateway = "false" // set default value
	}

	err = validateWorkloadIfIntentInputs(wif)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateWorkloadIfIntent(wif, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent, false)
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
func (h workloadifintentHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var wif moduleLib.WorkloadIfIntent
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	workloadIntent := vars["workload-intent"]

	err := json.NewDecoder(r.Body).Decode(&wif)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if wif.Metadata.Name == "" {
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if wif.Metadata.Name != name {
		fmt.Printf("bodyname = %v, name= %v\n", wif.Metadata.Name, name)
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	// set default value
	if len(wif.Spec.DefaultGateway) == 0 {
		wif.Spec.DefaultGateway = "false" // set default value
	}

	err = validateWorkloadIfIntentInputs(wif)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateWorkloadIfIntent(wif, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent, true)
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
func (h workloadifintentHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	workloadIntent := vars["workload-intent"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetWorkloadIfIntents(project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetWorkloadIfIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent)
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
func (h workloadifintentHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]
	workloadIntent := vars["workload-intent"]

	err := h.client.DeleteWorkloadIfIntent(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, workloadIntent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
