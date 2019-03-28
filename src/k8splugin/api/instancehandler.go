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
	"errors"
	"io"
	"net/http"

	"k8splugin/internal/app"

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
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing CloudRegion in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.RBName == "" || b.RBVersion == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing resource bundle parameters in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.ProfileName == "" {
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
		http.Error(w, "Body empty", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Check body for expected parameters
	err = i.validateBody(resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	resp, err := i.client.Create(resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler retrieves information about an instance via the ID
func (i instanceHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["instID"]

	resp, err := i.client.Get(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

// // UpdateHandler method to update a VNF instance.
// func UpdateHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	id := vars["vnfInstanceId"]

// 	var resource UpdateVnfRequest

// 	if r.Body == nil {
// 		http.Error(w, "Body empty", http.StatusBadRequest)
// 		return
// 	}

// 	err := json.NewDecoder(r.Body).Decode(&resource)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
// 		return
// 	}

// 	err = validateBody(resource)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
// 		return
// 	}

// 	kubeData, err := utils.ReadCSARFromFileSystem(resource.CsarID)

// 	if kubeData.Deployment == nil {
// 		werr := pkgerrors.Wrap(err, "Update VNF deployment error")
// 		http.Error(w, werr.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	kubeData.Deployment.SetUID(types.UID(id))

// 	if err != nil {
// 		werr := pkgerrors.Wrap(err, "Update VNF deployment information error")
// 		http.Error(w, werr.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	// (TODO): Read kubeconfig for specific Cloud Region from local file system
// 	// if present or download it from AAI
// 	s, err := NewVNFInstanceService("../kubeconfig/config")
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	err = s.Client.UpdateDeployment(kubeData.Deployment, resource.Namespace)
// 	if err != nil {
// 		werr := pkgerrors.Wrap(err, "Update VNF error")

// 		http.Error(w, werr.Error(), http.StatusInternalServerError)
// 		return
// 	}

// 	resp := UpdateVnfResponse{
// 		DeploymentID: id,
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)

// 	err = json.NewEncoder(w).Encode(resp)
// 	if err != nil {
// 		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
// 		http.Error(w, werr.Error(), http.StatusInternalServerError)
// 	}
// }
