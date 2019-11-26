/*
Copyright 2019 Intel Corporation.
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

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"

	"github.com/gorilla/mux"
)

// Used to store the backend implementation objects
// Also simplifies the mocking needed for unit testing
type registryHandler struct {
	// Interface that implements the Instance operations
	client app.RegistryManager
}

type registryPOSTResponse struct {
	ID                 string                   `json:"id"`
	CloudOwner         string                   `json:"cloud_owner"`
	CloudRegionID      string                   `json:"cloud_region_id"`
}

type registryDELETEResponse struct {
	ID                 string                   `json:"id"`
	CloudOwner         string                   `json:"cloud_owner"`
	CloudRegionID      string                   `json:"cloud_region_id"`
}

// createHandler send registry request and save registry status
func (b registryHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cloudOwner := vars["cloud-owner"]
	cloudRegion := vars["cloud-region"]

	// Setup the resource parameters for making the request
	var RegistryReq app.RegistryRequest
	RegistryReq.CloudOwner  = cloudOwner
	RegistryReq.CloudRegion = cloudRegion

	resp, err := b.client.Create(RegistryReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	registryResp := registryPOSTResponse{
		ID: resp.ID,
		CloudOwner: cloudOwner,
		CloudRegionID: cloudRegion,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(registryResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler retrieves information about an instance via the ID
func (b registryHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	cloudOwner := vars["cloud-owner"]
	cloudRegion := vars["cloud-region"]

	resp, err := b.client.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	id = resp.ID

	registryResp := app.RegistryResponse{
		ID: id,
		Request: app.RegistryRequest{
			CloudOwner: cloudOwner,
			CloudRegion: cloudRegion,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(registryResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// deleteHandler method terminates an instance via the ID
func (b registryHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ID := vars["id"]
	cloudOwner := vars["cloud-owner"]
	cloudRegion := vars["cloud-region"]

	err := b.client.Delete(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	registryResp := registryDELETEResponse{
		ID: ID,
		CloudOwner:   cloudOwner,
		CloudRegionID:   cloudRegion,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(registryResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
