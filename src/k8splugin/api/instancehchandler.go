/*
Copyright © 2021 Samsung Electronics
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
	"net/http"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/healthcheck"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"

	"github.com/gorilla/mux"
)

// Used to store the backend implementation objects
// Also simplifies the mocking needed for unit testing
type instanceHCHandler struct {
	// Interface that implements Healthcheck operations
	client healthcheck.InstanceHCManager
}

func (ih instanceHCHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]

	resp, err := ih.client.Create(id)
	if err != nil {
		log.Error("Error scheduling healhtcheck", log.Fields{
			"error":    err,
			"instance": id,
			"response": resp,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error("Error Marshaling Reponse", log.Fields{
			"error":    err,
			"response": resp,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (ih instanceHCHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instID"]
	healthcheckId := vars["hcID"]

	resp, err := ih.client.Get(instanceId, healthcheckId)
	if err != nil {
		log.Error("Error getting Instance's healthcheck", log.Fields{
			"error":         err,
			"instanceID":    instanceId,
			"healthcheckID": healthcheckId,
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

func (ih instanceHCHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instID"]
	healthcheckId := vars["hcID"]

	err := ih.client.Delete(instanceId, healthcheckId)
	if err != nil {
		log.Error("Error deleting Instance's healthcheck", log.Fields{
			"error":         err,
			"instanceID":    instanceId,
			"healthcheckID": healthcheckId,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (ih instanceHCHandler) listHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]

	resp, err := ih.client.List(id)
	if err != nil {
		log.Error("Error getting instance healthcheck overview", log.Fields{
			"error":       err,
			"instance-id": id,
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
