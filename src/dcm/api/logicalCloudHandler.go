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

	"github.com/gorilla/mux"
	"github.com/onap/multicloud-k8s/src/dcm/pkg/module"
	pkgerrors "github.com/pkg/errors"
)

// logicalCloudHandler is used to store backend implementations objects
type logicalCloudHandler struct {
	client        module.LogicalCloudManager
	clusterClient module.ClusterManager
	quotaClient   module.QuotaManager
}

// CreateHandler handles creation of the logical cloud entry in the database

func (h logicalCloudHandler) createHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	project := vars["project-name"]
	var v module.LogicalCloud

	err := json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Logical Cloud Name is required.
	if v.MetaData.LogicalCloudName == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.Create(project, v)
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
// Returns a Logical Cloud
func (h logicalCloudHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetAll(project)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.Get(project, name)
		if err != nil {
			if err.Error() == "Logical Cloud does not exist" {
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

// UpdateHandler handles Update operations on a particular logical cloud
func (h logicalCloudHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
	var v module.LogicalCloud
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

	err := json.NewDecoder(r.Body).Decode(&v)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if v.MetaData.LogicalCloudName == "" {
		http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.Update(project, name, v)
	if err != nil {
		if err.Error() == "Logical Cloud does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(),
			http.StatusInternalServerError)
		return
	}
}

func (h logicalCloudHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

	err := h.client.Delete(project, name)
	if err != nil {
		if err.Error() == "Logical Cloud does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h logicalCloudHandler) applyHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

	// Get logical cloud
	lc, err := h.client.Get(project, name)
	if err != nil {
		if err.Error() == "Logical Cloud does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, ctxVal, err := h.client.GetLogicalCloudContext(name)
	if ctxVal != "" {
		err = pkgerrors.New("Logical Cloud already applied")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Get Clusters
	clusters, err := h.clusterClient.GetAllClusters(project, name)

	if err != nil {
		if err.Error() == "No Cluster References associated" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Get Quotas
	quotas, err := h.quotaClient.GetAllQuotas(project, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = module.CreateEtcdContext(lc, clusters, quotas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

func (h logicalCloudHandler) terminateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	name := vars["logical-cloud-name"]

	// Get logical cloud
	lc, err := h.client.Get(project, name)
	if err != nil {
		if err.Error() == "Logical Cloud does not exist" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, ctxVal, err := h.client.GetLogicalCloudContext(name)
	if ctxVal == "" {
		err = pkgerrors.New("Logical Cloud hasn't been applied yet")
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	// Get Clusters
	clusters, err := h.clusterClient.GetAllClusters(project, name)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//Get Quotas
	quotas, err := h.quotaClient.GetAllQuotas(project, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = module.DestroyEtcdContext(lc, clusters, quotas)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
