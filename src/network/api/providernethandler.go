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

	moduleLib "github.com/onap/multicloud-k8s/src/network/pkg/module"
	"github.com/onap/multicloud-k8s/src/orchestrator/utils"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/validation"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type providernetHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client moduleLib.ProviderNetManager
}

// Check for valid format of input parameters
func validateProviderNetInputs(p moduleLib.ProviderNet) error {
	// validate name
	errs := utils.IsValidName(p.Metadata.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid provider network name - name=[%v], errors: %v", p.Metadata.Name, errs)
	}

	// validate the subnets
	subnets := p.Spec.Ipv4Subnets
	for _, subnet := range subnets {
		err := moduleLib.ValidateSubnet(subnet)
		if err != nil {
			return pkgerrors.Wrap(err, "invalid subnet")
		}
	}

	// validate the VLAN ID
	if p.Spec.Vlan.VlanId < 0 || p.Spec.Vlan.VlanId > 4095 {
		return pkgerrors.Errorf("Invalid VlAN ID %v", p.Spec.Vlan.VlanId)
	}

	// validate the vlan Node Selector value
	expectLabels := false
	switch strings.ToLower(p.Spec.Vlan.VlanNodeSelector) {
	case moduleLib.VLAN_NODE_ANY:
	case moduleLib.VLAN_NODE_SPECIFIC:
		expectLabels = true
	default:
		return pkgerrors.Errorf("Invalid VlAN Node Selector %v", p.Spec.Vlan.VlanNodeSelector)
	}

	// validate the node label list
	// Looking for labels to match following formats:
	//  "<DNS1123Subdomain>/<DNS1123Label>=<LabelValue>"
	//  "<DNS1123Label>=<LabelValue>"
	//  "<LabelValue>"
	gotLabel := false
	for _, label := range p.Spec.Vlan.NodeLabelList {
		expectLabelName := false
		expectLabelPrefix := false

		// split label up into prefix, name and value
		// format:  prefix/label=value
		var labelprefix, labelname, labelvalue string
		kv := strings.SplitN(label, "=", 2)
		if len(kv) == 1 {
			labelprefix = ""
			labelname = ""
			labelvalue = kv[0]
		} else {
			pn := strings.SplitN(kv[0], "/", 2)
			if len(pn) == 1 {
				labelprefix = ""
				labelname = pn[0]
			} else {
				labelprefix = pn[0]
				labelname = pn[1]
				expectLabelPrefix = true
			}
			labelvalue = kv[1]
			// if "=" was in the label input, then expect a non-zero length name
			expectLabelName = true
		}

		// check label prefix validity - prefix is optional
		if len(labelprefix) > 0 {
			errs := validation.IsDNS1123Subdomain(labelprefix)
			if len(errs) > 0 {
				return pkgerrors.Errorf("Invalid VlAN Node Label prefix - label=[%v], labelprefix=[%v], errors: %v", label, labelprefix, errs)
			}
		} else if expectLabelPrefix {
			return pkgerrors.Errorf("Invalid VlAN Node Label prefix - label=[%v], labelprefix=[%v]", label, labelprefix)
		}
		if expectLabelName {
			errs := validation.IsDNS1123Label(labelname)
			if len(errs) > 0 {
				return pkgerrors.Errorf("Invalid VlAN Node Label name - label=[%v], labelname=[%v], errors: %v", label, labelname, errs)
			}
		}
		if len(labelvalue) > 0 {
			errs := validation.IsValidLabelValue(labelvalue)
			if len(errs) > 0 {
				return pkgerrors.Errorf("Invalid VlAN Node Label value - label=[%v], labelvalue=[%v], errors: %v", label, labelvalue, errs)
			}
		} else {
			// expect a non-zero value
			return pkgerrors.Errorf("Invalid VlAN Node Label value - label=[%v], labelvalue=[%v]", label, labelvalue)
		}

		gotLabel = true
	}

	// Need at least one label if node selector value was "specific"
	// (if selector is "any" - don't care if labels were supplied or not
	if expectLabels && !gotLabel {
		return pkgerrors.Errorf("VlAN Node Labels required for node selector \"%v\"", moduleLib.VLAN_NODE_SPECIFIC)
	}

	return nil
}

// Create handles creation of the ProviderNet entry in the database
func (h providernetHandler) createProviderNetHandler(w http.ResponseWriter, r *http.Request) {
	var p moduleLib.ProviderNet
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
	var p moduleLib.ProviderNet
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
