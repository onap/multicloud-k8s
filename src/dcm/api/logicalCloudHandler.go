/*
* Copyright 2018 Intel Corporation, Inc
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
        "dcm/internal/logicalcloud"
        "net/http"
        "io"
        "github.com/gorilla/mux"
)


// logicalCloudHandler is used to store backend implementations objects
type logicalCloudHandler struct {
    client logicalcloud.LogicalCloudManager
}

// CreateHandler handles creation of the logical cloud entry in the database

func (h logicalCloudHandler) createHandler(w http.ResponseWriter, r *http.Request) {
    var v logicalcloud.LogicalCloud

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

    // Logical Cloud Name is required.
    if v.LogicalCloudName == "" {
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

    ret, err := h.client.Create(v)
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
        name := vars["logicalcloudname"]
        ret, err := h.client.Get(name)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        err = json.NewEncoder(w).Encode(ret)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
}

//deleteHandler handles DELETE operations on a particular record
func (h logicalCloudHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        name := vars["logicalcloudname"]
        
        err := h.client.Delete(name)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }

        w.WriteHeader(http.StatusNoContent)
}
