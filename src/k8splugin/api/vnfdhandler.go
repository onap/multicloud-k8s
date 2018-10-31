/*
Copyright 2018 Intel Corporation.
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

	"github.com/gorilla/mux"
	"k8splugin/vnfd"
)

// vnfdCreateHandler handles creation of the vnfd entry in the database
func vnfdCreateHandler(w http.ResponseWriter, r *http.Request) {
	var v vnfd.VNFDefinition

	if r.Body == nil {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if v.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := vnfd.CreateVNFDefinition(v)
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

// vnfdUploadHandler handles creation of the vnfd entry in the database
func vnfdUploadHandler(w http.ResponseWriter, r *http.Request) {
}

// vnfdListHandler handles GET (list) operations on the /v1/vnfd endpoint
// Returns a list of vnfd.VNFDefinitions
func vnfdListHandler(w http.ResponseWriter, r *http.Request) {

	ret, err := vnfd.ListVNFDefinitions()
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

// vnfdGetHandler handles GET operations on a particular VNFID
// Returns a vnfd.VNFDefinition
func vnfdGetHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	vnfdID := vars["vnfdID"]

	ret, err := vnfd.GetVNFDefinition(vnfdID)
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

// vnfdDeleteHandler handles DELETE operations on a particular VNFID
func vnfdDeleteHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	vnfdID := vars["vnfdID"]

	err := vnfd.DeleteVNFDefinition(vnfdID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
