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

	"github.com/gorilla/mux"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	controller "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
	mtypes "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/types"
	pkgerrors "github.com/pkg/errors"
)

var controllerJSONFile string = "json-schemas/controller.json"

// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type controllerHandler struct {
	// Interface that implements controller operations
	// We will set this variable with a mock interface for testing
	client controller.ControllerManager
}

// Check for valid format of input parameters
func validateControllerInputs(c controller.Controller) error {
	// validate metadata
	err := mtypes.IsValidMetadata(c.Metadata)
	if err != nil {
		return pkgerrors.Wrap(err, "Invalid controller metadata")
	}

	errs := validation.IsValidName(c.Spec.Host)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid host name for controller %v, errors: %v", c.Spec.Host, errs)
	}

	errs = validation.IsValidNumber(c.Spec.Port, 0, 65535)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid controller port [%v], errors: %v", c.Spec.Port, errs)
	}

	found := false
	for _, val := range controller.CONTROLLER_TYPES {
		if c.Spec.Type == val {
			found = true
			break
		}
	}
	if !found {
		return pkgerrors.Errorf("Invalid controller type: %v", c.Spec.Type)
	}

	errs = validation.IsValidNumber(c.Spec.Priority, controller.MinControllerPriority, controller.MaxControllerPriority)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid controller priority = [%v], errors: %v", c.Spec.Priority, errs)
	}

	return nil
}

// Create handles creation of the controller entry in the database
func (h controllerHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var m controller.Controller

	err := json.NewDecoder(r.Body).Decode(&m)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(controllerJSONFile, m)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	ret, err := h.client.CreateController(m, false)
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

// Put handles creation or update of the controller entry in the database
func (h controllerHandler) putHandler(w http.ResponseWriter, r *http.Request) {
	var m controller.Controller
	vars := mux.Vars(r)
	name := vars["controller-name"]

	err := json.NewDecoder(r.Body).Decode(&m)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if m.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	// name in URL should match name in body
	if m.Metadata.Name != name {
		http.Error(w, "Mismatched name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateController(m, true)
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

// Get handles GET operations on a particular controller Name
// Returns a controller
func (h controllerHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["controller-name"]
	var ret interface{}
	var err error

	// handle the get all controllers case
	if len(name) == 0 {
		ret, err = h.client.GetControllers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {

		ret, err = h.client.GetController(name)
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

// Delete handles DELETE operations on a particular controller Name
func (h controllerHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["controller-name"]

	err := h.client.DeleteController(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
