/*
Copyright 2018 Intel Corporation.
Copyright © 2021 Samsung Electronics
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
	"net/http"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"

	pkgerrors "github.com/pkg/errors"
)

// Used to store the backend implementation objects
// Also simplifies the mocking needed for unit testing
type queryHandler struct {
	// Interface that implements the Instance operations
	client app.QueryManager
}

func (i queryHandler) validateBody(body interface{}) error {
	switch b := body.(type) {
	case app.InstanceRequest:
		if b.CloudRegion == "" {
			log.Error("CreateVnfRequest Bad Request", log.Fields{
				"cloudRegion": "Missing CloudRegion in POST request",
			})
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing CloudRegion in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.RBName == "" || b.RBVersion == "" {
			log.Error("CreateVnfRequest Bad Request", log.Fields{
				"message":   "One of RBName, RBVersion is missing",
				"RBName":    b.RBName,
				"RBVersion": b.RBVersion,
			})
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing resource bundle parameters in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.ProfileName == "" {
			log.Error("CreateVnfRequest bad request", log.Fields{
				"ProfileName": "Missing profile name in POST request",
			})
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing profile name in POST request"), "CreateVnfRequest bad request")
			return werr
		}
	}
	return nil
}

// queryHandler retrieves information about specified resources for instance
func (i queryHandler) queryHandler(w http.ResponseWriter, r *http.Request) {
	namespace := r.FormValue("Namespace")
	cloudRegion := r.FormValue("CloudRegion")
	apiVersion := r.FormValue("ApiVersion")
	kind := r.FormValue("Kind")
	name := r.FormValue("Name")
	labels := r.FormValue("Labels")
	if cloudRegion == "" {
		http.Error(w, "Missing CloudRegion mandatory parameter", http.StatusBadRequest)
		return
	}
	if apiVersion == "" {
		http.Error(w, "Missing ApiVersion mandatory parameter", http.StatusBadRequest)
		return
	}
	if kind == "" {
		http.Error(w, "Missing Kind mandatory parameter", http.StatusBadRequest)
		return
	}
	// instance id is irrelevant here
	resp, err := i.client.Query(namespace, cloudRegion, apiVersion, kind, name, labels, "query")
	if err != nil {
		log.Error("Error getting Query results", log.Fields{
			"error":       err,
			"cloudRegion": cloudRegion,
			"namespace":   namespace,
			"apiVersion":  apiVersion,
			"kind":        kind,
			"name":        name,
			"labels":      labels,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
