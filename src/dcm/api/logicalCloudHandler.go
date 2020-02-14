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
    "net/http"
    "io"
    "github.com/onap/multicloud-k8s/src/dcm/internal/logicalcloud"
    "github.com/gorilla/mux"
)


// logicalCloudHandler is used to store backend implementations objects
type logicalCloudHandler struct {
    client logicalcloud.LogicalCloudManager
}

// CreateHandler handles creation of the logical cloud entry in the database

func (h logicalCloudHandler) createHandler(w http.ResponseWriter, r *http.Request) {

    vars := mux.Vars(r)
    project := vars["project-name"]
    var v logicalcloud.LogicalCloud

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
    var v logicalcloud.LogicalCloud
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
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}

func (h logicalCloudHandler) applyHandler(w http.ResponseWriter, r *http.Request) {
    /*vars := mux.Vars(r)
    project := vars["project-name"]
    name := vars["logical-cloud-name"]*/
    /*ret, err := h.client.Get(project, name)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }*/

    // Do Some Work
    // someApplyFunction(project, name, ret)
}
