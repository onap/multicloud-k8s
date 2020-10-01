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

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/onap/multicloud-k8s/src/ovnaction/pkg/module"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type chainHandler struct {
	// Interface that implements workload intent operations
	// We will set this variable with a mock interface for testing
	client moduleLib.ChainManager
}

func validateRoutingNetwork(r moduleLib.RoutingNetwork) error {
	errs := validation.IsValidName(r.NetworkName)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid routing network name: %v", errs)
	}

	err := validation.IsIpv4Cidr(r.Subnet)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid routing network subnet")
	}

	err = validation.IsIpv4(r.GatewayIP)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid routing network gateway IP")
	}

	return nil
}

// validateNetworkChain checks that the network chain string input follows the
// generic format:  "app=app1,net1,app=app2,net2, ..... ,netN-1,app=appN"
// assume "app=app1" can conform to validation.IsValidLabel() with an "="
func validateNetworkChain(chain string) error {
	elems := strings.Split(chain, ",")

	// chain needs at least two apps and a network
	if len(elems) < 3 {
		return pkgerrors.Errorf("Network chain is too short")
	}

	// chain needs to have an odd number of elements
	if len(elems)%2 == 0 {
		return pkgerrors.Errorf("Invalid network chain - even number of elements")
	}

	for i, s := range elems {
		// allows whitespace in comma separated elements
		t := strings.TrimSpace(s)
		// if even element, verify a=b format
		if i%2 == 0 {
			if strings.Index(t, "=") < 1 {
				return pkgerrors.Errorf("Invalid deployment label element of network chain")
			}
			errs := validation.IsValidLabel(t)
			if len(errs) > 0 {
				return pkgerrors.Errorf("Invalid deployment label element: %v", errs)
			}
		} else {
			errs := validation.IsValidName(t)
			if len(errs) > 0 {
				return pkgerrors.Errorf("Invalid network element of network chain: %v", errs)
			}
		}
	}
	return nil
}

// Check for valid format of input parameters
func validateChainInputs(ch moduleLib.Chain) error {
	// validate metadata
	err := moduleLib.IsValidMetadata(ch.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid network chain metadata")
	}

	if strings.ToLower(ch.Spec.ChainType) != moduleLib.RoutingChainType {
		return pkgerrors.Wrap(err, "Invalid network chain type")
	}

	for _, r := range ch.Spec.RoutingSpec.LeftNetwork {
		err = validateRoutingNetwork(r)
		if err != nil {
			return err
		}
	}

	for _, r := range ch.Spec.RoutingSpec.RightNetwork {
		err = validateRoutingNetwork(r)
		if err != nil {
			return err
		}
	}

	err = validateNetworkChain(ch.Spec.RoutingSpec.NetworkChain)
	if err != nil {
		return err
	}

	errs := validation.IsValidName(ch.Spec.RoutingSpec.Namespace)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid network chain route spec namespace: %v", errs)
	}

	return nil
}

// Create handles creation of the Chain entry in the database
func (h chainHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var ch moduleLib.Chain
	vars := mux.Vars(r)
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&ch)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if ch.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	err = validateChainInputs(ch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateChain(ch, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, false)
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

// Put handles creation/update of the Chain entry in the database
func (h chainHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var ch moduleLib.Chain
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := json.NewDecoder(r.Body).Decode(&ch)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if ch.Metadata.Name == "" {
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	// Name in URL should match name in body
	if ch.Metadata.Name != name {
		fmt.Printf("bodyname = %v, name= %v\n", ch.Metadata.Name, name)
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	err = validateChainInputs(ch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateChain(ch, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent, true)
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

// Get handles GET operations on a particular Chain Name
// Returns a Chain
func (h chainHandler) getHandler(w http.ResponseWriter, r *http.Request) {
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
		ret, err = h.client.GetChains(project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetChain(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
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

// Delete handles DELETE operations on a particular Chain
func (h chainHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	project := vars["project"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	deployIntentGroup := vars["deployment-intent-group-name"]
	netControlIntent := vars["net-control-intent"]

	err := h.client.DeleteChain(name, project, compositeApp, compositeAppVersion, deployIntentGroup, netControlIntent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
