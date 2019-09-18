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

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	logr "github.com/sirupsen/logrus"
)

// Used to store the backend implementation objects
// Also simplifies the mocking needed for unit testing
type instanceHandler struct {
	// Interface that implements the Instance operations
	client app.InstanceManager
}

func (i instanceHandler) validateBody(body interface{}) error {
	logr.SetFormatter(&logr.JSONFormatter{})
	switch b := body.(type) {
	case app.InstanceRequest:
		if b.CloudRegion == "" {
			logr.WithFields(logr.Fields{"CloudRegion": "Invalid/Missing CloudRegion in POST request"}).Error("CreateVnfRequest bad request")
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing CloudRegion in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.RBName == "" || b.RBVersion == "" {
			logr.WithFields(logr.Fields{"RBName": "Invalid/Missing resource bundle parameters in POST request", "RBVersion": "Invalid/Missing resource bundle parameters in POST request"}).Error("CreateVnfRequest bad request")
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing resource bundle parameters in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.ProfileName == "" {
			logr.WithFields(logr.Fields{"ProfileName": "Invalid/Missing profile name in POST request"}).Error("CreateVnfRequest bad request")
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
		logr.WithFields(logr.Fields{"Error": "Body empty"}).Error("http.StatusBadRequest")
		http.Error(w, "Body empty", http.StatusBadRequest)
		return
	case err != nil:
		logr.WithFields(logr.Fields{"Error": "http.StatusUnprocessableEntity"}).Error("http.StatusUnprocessableEntity")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Check body for expected parameters
	err = i.validateBody(resource)
	if err != nil {
		logr.WithFields(logr.Fields{"Error": "http.StatusUnprocessableEntity"}).Error("StatusUnprocessableEntity")
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	resp, err := i.client.Create(resource)
	if err != nil {
		logr.WithFields(logr.Fields{"Error": "http.StatusInternalServerError"}).Error("StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logr.WithFields(logr.Fields{"Error": "http.StatusInternalServerError"}).Error("StatusInternalServerError")
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
		logr.WithFields(logr.Fields{"Error": "http.StatusInternalServerError"}).Error("StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logr.WithFields(logr.Fields{"Error": "http.StatusInternalServerError"}).Error("StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// listHandler retrieves information about an instance via the ID
func (i instanceHandler) listHandler(w http.ResponseWriter, r *http.Request) {

	//If parameters are not provided, they are sent as empty strings
	//Which will list all instances
	rbName := r.FormValue("rb-name")
	rbVersion := r.FormValue("rb-version")
	ProfileName := r.FormValue("profile-name")

	resp, err := i.client.List(rbName, rbVersion, ProfileName)
	if err != nil {
		logr.WithFields(logr.Fields{"Error": "http.StatusInternalServerError"}).Error("StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		logr.WithFields(logr.Fields{"Error": "http.StatusInternalServerError"}).Error("StatusInternalServerError")
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
		logr.WithFields(logr.Fields{"Error": "http.StatusInternalServerError"}).Error("StatusInternalServerError")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}

// // UpdateHandler method to update a VNF instance.
// func UpdateHandler(w http.ResponseWriter, r *http.Request) {
//      vars := mux.Vars(r)
//      id := vars["vnfInstanceId"]

//      var resource UpdateVnfRequest

//      if r.Body == nil {
//              http.Error(w, "Body empty", http.StatusBadRequest)
//              return
//      }

//      err := json.NewDecoder(r.Body).Decode(&resource)
//      if err != nil {
//              http.Error(w, err.Error(), http.StatusUnprocessableEntity)
//              return
//      }

//      err = validateBody(resource)
//      if err != nil {
//              http.Error(w, err.Error(), http.StatusUnprocessableEntity)
//              return
//      }

//      kubeData, err := utils.ReadCSARFromFileSystem(resource.CsarID)

//      if kubeData.Deployment == nil {
//              werr := pkgerrors.Wrap(err, "Update VNF deployment error")
//              http.Error(w, werr.Error(), http.StatusInternalServerError)
//              return
//      }
//      kubeData.Deployment.SetUID(types.UID(id))

//      if err != nil {
//              werr := pkgerrors.Wrap(err, "Update VNF deployment information error")
//              http.Error(w, werr.Error(), http.StatusInternalServerError)
//              return
//      }

//      // (TODO): Read kubeconfig for specific Cloud Region from local file system
//      // if present or download it from AAI
//      s, err := NewVNFInstanceService("../kubeconfig/config")
//      if err != nil {
//              http.Error(w, err.Error(), http.StatusInternalServerError)
//              return
//      }

//      err = s.Client.UpdateDeployment(kubeData.Deployment, resource.Namespace)
//      if err != nil {
//              werr := pkgerrors.Wrap(err, "Update VNF error")

//              http.Error(w, werr.Error(), http.StatusInternalServerError)
//              return
//      }

//      resp := UpdateVnfResponse{
//              DeploymentID: id,
//      }

//      w.Header().Set("Content-Type", "application/json")
//      w.WriteHeader(http.StatusCreated)

//      err = json.NewEncoder(w).Encode(resp)
//      if err != nil {
//              werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
//              http.Error(w, werr.Error(), http.StatusInternalServerError)
//      }
// }
