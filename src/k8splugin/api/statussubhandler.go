/*
Copyright © 2022 Orange
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
	"io"
	"net/http"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"

	"github.com/gorilla/mux"
)

// Used to store the backend implementation objects
// Also simplifies the mocking needed for unit testing
type instanceStatusSubHandler struct {
	// Interface that implements Status Subscription operations
	client app.InstanceStatusSubManager
}

func (iss instanceStatusSubHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]

	var subRequest app.SubscriptionRequest

	err := json.NewDecoder(r.Body).Decode(&subRequest)
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

	// Name is required.
	if subRequest.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	// MinNotifyInterval cannot be less than 0
	if subRequest.MinNotifyInterval < 0 {
		http.Error(w, "Min Notify Interval has invalid value", http.StatusBadRequest)
		return
	}

	// CallbackUrl is required
	if subRequest.CallbackUrl == "" {
		http.Error(w, "CallbackUrl has invalid value", http.StatusBadRequest)
		return
	}

	resp, err := iss.client.Create(id, subRequest)
	if err != nil {
		log.Error("Error creating subscription", log.Fields{
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
		log.Error("Error Marshaling Response", log.Fields{
			"error":    err,
			"response": resp,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (iss instanceStatusSubHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instID"]
	subId := vars["subID"]

	resp, err := iss.client.Get(instanceId, subId)
	if err != nil {
		log.Error("Error getting instance's Status Subscription", log.Fields{
			"error":          err,
			"instanceID":     instanceId,
			"subscriptionID": subId,
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

func (iss instanceStatusSubHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instID"]
	subId := vars["subID"]

	var subRequest app.SubscriptionRequest

	err := json.NewDecoder(r.Body).Decode(&subRequest)
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

	// MinNotifyInterval cannot be less than 0
	if subRequest.MinNotifyInterval < 0 {
		http.Error(w, "Min Notify Interval has invalid value", http.StatusBadRequest)
		return
	}

	// CallbackUrl is required
	if subRequest.CallbackUrl == "" {
		http.Error(w, "CallbackUrl has invalid value", http.StatusBadRequest)
		return
	}

	resp, err := iss.client.Update(instanceId, subId, subRequest)
	if err != nil {
		log.Error("Error updating instance's Status Subscription", log.Fields{
			"error":          err,
			"instanceID":     instanceId,
			"subscriptionID": subId,
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

func (iss instanceStatusSubHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceId := vars["instID"]
	subId := vars["subID"]

	err := iss.client.Delete(instanceId, subId)
	if err != nil {
		log.Error("Error deleting instance's Status Subscription", log.Fields{
			"error":          err,
			"instanceID":     instanceId,
			"subscriptionID": subId,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (iss instanceStatusSubHandler) listHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]

	resp, err := iss.client.List(id)
	if err != nil {
		log.Error("Error listing instance Status Subscriptions", log.Fields{
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
