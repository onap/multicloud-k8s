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

	netintents "github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents"
	nettypes "github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents/types"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

var vnJSONFile string = "json-schemas/virtual-network.json"


// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type networkHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client netintents.NetworkManager
}

// Check for valid format of input parameters
func validateNetworkInputs(p netintents.Network) error {
	// validate name
	errs := validation.IsValidName(p.Metadata.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid network name - name=[%v], errors: %v", p.Metadata.Name, errs)
	}

	// validate cni type
	found := false
	for _, val := range nettypes.CNI_TYPES {
		if p.Spec.CniType == val {
			found = true
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid cni type: %v", p.Spec.CniType)
	}

	subnets := p.Spec.Ipv4Subnets
	for _, subnet := range subnets {
		err := nettypes.ValidateSubnet(subnet)
		if err != nil {
			return pkgerrors.Wrap(err, "invalid subnet")
		}
	}
	return nil
}

// Create handles creation of the Network entry in the database
func (h networkHandler) createNetworkHandler(w http.ResponseWriter, r *http.Request) {
	var p netintents.Network
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(vnJSONFile, p)
if err != nil {
	http.Error(w, err.Error(), httpError)
	return
}

	// Name is required.
	if p.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateNetworkInputs(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetwork(p, clusterProvider, cluster, false)
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
func (h networkHandler) putNetworkHandler(w http.ResponseWriter, r *http.Request) {
	var p netintents.Network
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]

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
	if p.Metadata.Name == "" {
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if p.Metadata.Name != name {
		fmt.Printf("bodyname = %v, name= %v\n", p.Metadata.Name, name)
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateNetworkInputs(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateNetwork(p, clusterProvider, cluster, true)
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
func (h networkHandler) getNetworkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetNetworks(clusterProvider, cluster)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetNetwork(name, clusterProvider, cluster)
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
func (h networkHandler) deleteNetworkHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]

	err := h.client.DeleteNetwork(name, clusterProvider, cluster)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
