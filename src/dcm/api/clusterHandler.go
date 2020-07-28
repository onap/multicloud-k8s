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
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
* implied.
* See the License for the specific language governing permissions
* and
* limitations under the License.
 */

package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/onap/multicloud-k8s/src/dcm/pkg/module"
	pkgerrors "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

// clusterHandler is used to store backend implementations objects
type clusterHandler struct {
	client module.ClusterManager
}

// CreateHandler handles creation of the cluster reference entry in the database

func (h clusterHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	var v module.Cluster

	err := json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Cluster Reference Name is required.
	if v.MetaData.ClusterReference == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateCluster(project, logicalCloud, v)
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

// getHandler handle GET operations on a particular name
// Returns a Cluster Reference
func (h clusterHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	name := vars["cluster-reference"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAllClusters(project, logicalCloud)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetCluster(project, logicalCloud, name)
		if err != nil {
			if err.Error() == "Cluster Reference does not exist" {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
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

// UpdateHandler handles Update operations on a particular cluster reference
func (h clusterHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	var v module.Cluster
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	name := vars["cluster-reference"]

	err := json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if v.MetaData.ClusterReference == "" {
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.UpdateCluster(project, logicalCloud, name, v)
	if err != nil {
		if err.Error() == "Cluster Reference does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
}

//deleteHandler handles DELETE operations on a particular record
func (h clusterHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	name := vars["cluster-reference"]

	err := h.client.DeleteCluster(project, logicalCloud, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getConfigHandler handles GET operations on kubeconfigs
// Returns a kubeconfig file
func (h clusterHandler) getConfigHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	logicalCloud := vars["logical-cloud-name"]
	name := vars["cluster-reference"]
	var ret interface{}
	var err error

	ret, err = h.client.GetCluster(project, logicalCloud, name)
	if err != nil {
		if err.Error() == "Cluster Reference does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lcClient := module.NewLogicalCloudClient()
	_, ctxVal, err := lcClient.GetLogicalCloudContext(logicalCloud)
	if ctxVal == "" {
		err = pkgerrors.New("Logical Cloud hasn't been applied yet")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ret, err = h.client.GetClusterConfig(project, logicalCloud, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
