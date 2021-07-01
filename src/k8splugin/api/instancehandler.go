/*
Copyright 2018 Intel Corporation.
Copyright © 2021 Samsung Electronics
Copyright © 2021 Orange

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
)

// Used to store the backend implementation objects
// Also simplifies the mocking needed for unit testing
type instanceHandler struct {
	// Interface that implements the Instance operations
	client app.InstanceManager
}

func (i instanceHandler) validateBody(body interface{}) error {
	switch b := body.(type) {
	case app.InstanceRequest:
		if b.CloudRegion == "" {
			log.Error("CreateVnfRequest Bad Request", log.Fields{
				"cloudRegion": "Missing CloudRegion in POST request",
			})
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing CloudRegion in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.RBName == "" || b.RBVersion == "" {
			log.Error("CreateVnfRequest Bad Request", log.Fields{
				"message":   "One of RBName, RBVersion is missing",
				"RBName":    b.RBName,
				"RBVersion": b.RBVersion,
			})
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing resource bundle parameters in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.ProfileName == "" {
			log.Error("CreateVnfRequest bad request", log.Fields{
				"ProfileName": "Missing profile name in POST request",
			})
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing profile name in POST request"), "CreateVnfRequest bad request")
			return werr
		}
	}
	return nil
}

func (i instanceHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var resource app.InstanceRequest

	err := json.NewDecoder(r.Body).Decode(&resource)
	switch {
	case err == io.EOF:
		log.Error("Body Empty", log.Fields{
			"error": io.EOF,
		})
		http.Error(w, "Body empty", http.StatusBadRequest)
		return
	case err != nil:
		log.Error("Error unmarshaling Body", log.Fields{
			"error": err,
		})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Check body for expected parameters
	err = i.validateBody(resource)
	if err != nil {
		log.Error("Invalid Parameters in Body", log.Fields{
			"error": err,
		})
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	resp, err := i.client.Create(resource)
	if err != nil {
		log.Error("Error Creating Resource", log.Fields{
			"error":    err,
			"resource": resource,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("Error Marshaling Response", log.Fields{
			"error":    err,
			"response": resp,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler retrieves information about an instance via the ID
func (i instanceHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]
	var resp interface{}
	var err error

	if r.URL.Query().Get("full") == "true" {
		resp, err = i.client.GetFull(id)
	} else {
		resp, err = i.client.Get(id)
	}

	if err != nil {
		log.Error("Error getting Instance", log.Fields{
			"error": err,
			"id":    id,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("Error Marshaling Response", log.Fields{
			"error":    err,
			"response": resp,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// statusHandler retrieves status about an instance via the ID
func (i instanceHandler) statusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]

	resp, err := i.client.Status(id)
	if err != nil {
		log.Error("Error getting Status", log.Fields{
			"error": err,
			"id":    id,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("Error Marshaling Response", log.Fields{
			"error":    err,
			"response": resp,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// queryHandler retrieves information about specified resources for instance
func (i instanceHandler) queryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]
	apiVersion := r.FormValue("ApiVersion")
	kind := r.FormValue("Kind")
	name := r.FormValue("Name")
	labels := r.FormValue("Labels")
	if apiVersion == "" {
		http.Error(w, "Missing ApiVersion mandatory parameter", http.StatusBadRequest)
		return
	}
	if kind == "" {
		http.Error(w, "Missing Kind mandatory parameter", http.StatusBadRequest)
		return
	}
	resp, err := i.client.Query(id, apiVersion, kind, name, labels)
	if err != nil {
		log.Error("Error getting Query results", log.Fields{
			"error":      err,
			"id":         id,
			"apiVersion": apiVersion,
			"kind":       kind,
			"name":       name,
			"labels":     labels,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("Error Marshaling Response", log.Fields{
			"error":    err,
			"response": resp,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// listHandler retrieves information about an instance via the ID
func (i instanceHandler) listHandler(w http.ResponseWriter, r *http.Request) {

	//If parameters are not provided, they are sent as empty strings
	//Which will list all instances
	rbName := r.FormValue("rb-name")
	rbVersion := r.FormValue("rb-version")
	profileName := r.FormValue("profile-name")

	resp, err := i.client.List(rbName, rbVersion, profileName)
	if err != nil {
		log.Error("Error listing instances", log.Fields{
			"error":        err,
			"rb-name":      rbName,
			"rb-version":   rbVersion,
			"profile-name": profileName,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("Error Marshaling Response", log.Fields{
			"error":    err,
			"response": resp,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// deleteHandler method terminates an instance via the ID
func (i instanceHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]

	err := i.client.Delete(id)
	if err != nil {
		log.Error("Error Deleting Instance", log.Fields{
			"error": err,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}
