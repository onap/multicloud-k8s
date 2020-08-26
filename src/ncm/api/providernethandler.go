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
	"strings"

	netintents "github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents"
	nettypes "github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents/types"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

var pnetJSONFile string = "json-schemas/provider-network.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type providernetHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client netintents.ProviderNetManager
}

// Check for valid format of input parameters
func validateProviderNetInputs(p netintents.ProviderNet) error {
	// validate name
	errs := validation.IsValidName(p.Metadata.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid provider network name=[%v], errors: %v", p.Metadata.Name, errs)
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

	// validate the provider network type
	found = false
	for _, val := range nettypes.PROVIDER_NET_TYPES {
		if strings.ToUpper(p.Spec.ProviderNetType) == val {
			found = true
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid provider network type: %v", p.Spec.ProviderNetType)
	}

	// validate the subnets
	subnets := p.Spec.Ipv4Subnets
	for _, subnet := range subnets {
		err := nettypes.ValidateSubnet(subnet)
		if err != nil {
			return pkgerrors.Wrap(err, "invalid subnet")
		}
	}

	// validate the VLAN ID
	errs = validation.IsValidNumberStr(p.Spec.Vlan.VlanId, 0, 4095)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid VlAN ID %v - error: %v", p.Spec.Vlan.VlanId, errs)
	}

	// validate the VLAN Node Selector value
	expectLabels := false
	found = false
	for _, val := range nettypes.VLAN_NODE_SELECTORS {
		if strings.ToLower(p.Spec.Vlan.VlanNodeSelector) == val {
			found = true
			if val == nettypes.VLAN_NODE_SPECIFIC {
				expectLabels = true
			}
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid VlAN Node Selector %v", p.Spec.Vlan.VlanNodeSelector)
	}

	// validate the node label list
	gotLabels := false
	for _, label := range p.Spec.Vlan.NodeLabelList {
		errs = validation.IsValidLabel(label)
		if len(errs) > 0 {
			return pkgerrors.Errorf("Invalid Label=%v - errors: %v", label, errs)
		}
		gotLabels = true
	}

	// Need at least one label if node selector value was "specific"
	// (if selector is "any" - don't care if labels were supplied or not
	if expectLabels && !gotLabels {
		return pkgerrors.Errorf("Node Labels required for VlAN node selector \"%v\"", nettypes.VLAN_NODE_SPECIFIC)
	}

	return nil
}

// Create handles creation of the ProviderNet entry in the database
func (h providernetHandler) createProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	var p netintents.ProviderNet
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

	err, httpError := validation.ValidateJsonSchemaData(pnetJSONFile, p)
if err != nil {
	http.Error(w, err.Error(), httpError)
	return
}

	// Name is required.
	if p.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateProviderNetInputs(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateProviderNet(p, clusterProvider, cluster, false)
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

// Put handles creation/update of the ProviderNet entry in the database
func (h providernetHandler) putProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	var p netintents.ProviderNet
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

	err = validateProviderNetInputs(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateProviderNet(p, clusterProvider, cluster, true)
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

// Get handles GET operations on a particular ProviderNet Name
// Returns a ProviderNet
func (h providernetHandler) getProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetProviderNets(clusterProvider, cluster)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetProviderNet(name, clusterProvider, cluster)
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

// Delete handles DELETE operations on a particular ProviderNet  Name
func (h providernetHandler) deleteProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterProvider := vars["provider-name"]
	cluster := vars["cluster-name"]
	name := vars["name"]

	err := h.client.DeleteProviderNet(name, clusterProvider, cluster)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
