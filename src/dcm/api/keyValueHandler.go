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
    "bytes"
    "encoding/json"
    "net/http"
    "io"
    "github.com/onap/multicloud-k8s/src/dcm/internal/logicalcloud"
    "github.com/gorilla/mux"
)


// keyValueHandler is used to store backend implementations objects
type keyValueHandler struct {
    client logicalcloud.KeyValueManager
}

// CreateHandler handles creation of the key value entry in the database

func (h keyValueHandler) createHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    project := vars["project-name"]
    logicalCloud := vars["logical-cloud-name"]
    var v logicalcloud.KeyValue

    // Implemenation using multipart form
    // Review and enable/remove at a later date
    // Set Max size to 16mb here
    err := r.ParseMultipartForm(16777216)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnprocessableEntity)
        return
    }

    jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
    err = json.NewDecoder(jsn).Decode(&v)
    switch {
    case err == io.EOF:
        http.Error(w, "Empty body", http.StatusBadRequest)
        return
    case err != nil:
        http.Error(w, err.Error(), http.StatusUnprocessableEntity)
        return
    }

    // Key Value Name is required.
    if v.MetaData.KeyValueName == "" {
        http.Error(w, "Missing name in POST request", http.StatusBadRequest)
        return
    }

    //Read the file section and ignore the header
    file, _, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Unable to process file", http.StatusUnprocessableEntity)
        return
    }

    defer file.Close()

    ret, err := h.client.Create(project, logicalCloud, v)
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
// Returns a Key Value
func (h keyValueHandler) getHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    project := vars["project-name"]
    logicalCloud := vars["logical-cloud-name"]
    name := vars["kv-pair-name"]
    var ret interface{}
    var err error

    if len(name) == 0 {
        ret, err = h.client.GetAll(project, logicalCloud)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    } else {
        ret, err = h.client.Get(project, logicalCloud, name)
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

// UpdateHandler handles Update operations on a particular Key Value
func (h keyValueHandler) updateHandler(w http.ResponseWriter, r *http.Request) {
    var v logicalcloud.KeyValue
    vars := mux.Vars(r)
    project := vars["project-name"]
    logicalCloud := vars["logical-cloud-name"]
    name := vars["kv-pair-name"]

    err := r.ParseMultipartForm(16777216)
    if err != nil {
        http.Error(w, err.Error(),
        http.StatusUnprocessableEntity)
        return
    }

    jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
    err = json.NewDecoder(jsn).Decode(&v)
    switch {
        case err == io.EOF:
            http.Error(w, "Empty body", http.StatusBadRequest)
            return
        case err != nil:
            http.Error(w, err.Error(), http.StatusUnprocessableEntity)
            return
    }

    // Name is required.
    if v.MetaData.KeyValueName == "" {
        http.Error(w, "Missing name in PUT request", http.StatusBadRequest)
        return
    }

    //Read the file section and ignore the header
    file, _, err := r.FormFile("file")
    if err != nil {
        http.Error(w, "Unable to process file",
        http.StatusUnprocessableEntity)
        return
    }

    defer file.Close()

    ret, err := h.client.Update(project, logicalCloud, name, v)
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

//deleteHandler handles DELETE operations on a particular record
func (h keyValueHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    project := vars["project-name"]
    logicalCloud := vars["logical-cloud-name"]
    name := vars["kv-pair-name"]

        err := h.client.Delete(project, logicalCloud, name)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.WriteHeader(http.StatusNoContent)
}
