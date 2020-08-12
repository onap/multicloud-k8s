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
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/mux"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"
)

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type instantiationHandler struct {
	client moduleLib.InstantiationManager
}

func (h instantiationHandler) approveHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	iErr := h.client.Approve(p, ca, v, di)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (h instantiationHandler) instantiateHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	iErr := h.client.Instantiate(p, ca, v, di)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (h instantiationHandler) terminateHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	iErr := h.client.Terminate(p, ca, v, di)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)

}

func (h instantiationHandler) statusHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	p := vars["project-name"]
	ca := vars["composite-app-name"]
	v := vars["composite-app-version"]
	di := vars["deployment-intent-group-name"]

	qParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var queryInstance string
	if o, found := qParams["instance"]; found {
		queryInstance = o[0]
		if queryInstance == "" {
			http.Error(w, "Invalid query instance", http.StatusBadRequest)
			return
		}
	} else {
		queryInstance = "" // default instance value
	}

	var queryType string
	if t, found := qParams["type"]; found {
		queryType = t[0]
		if queryType != "cluster" && queryType != "rsync" {
			http.Error(w, "Invalid query type", http.StatusBadRequest)
			return
		}
	} else {
		queryType = "rsync" // default type
	}

	var queryOutput string
	if o, found := qParams["output"]; found {
		queryOutput = o[0]
		if queryOutput != "summary" && queryOutput != "all" && queryOutput != "detail" {
			http.Error(w, "Invalid query output", http.StatusBadRequest)
			return
		}
	} else {
		queryOutput = "all" // default output format
	}

	var queryApps []string
	if a, found := qParams["app"]; found {
		queryApps = a
		for _, app := range queryApps {
			errs := validation.IsValidName(app)
			if len(errs) > 0 {
				http.Error(w, "Invalid app query", http.StatusBadRequest)
				return
			}
		}
	} else {
		queryApps = make([]string, 0)
	}

	var queryClusters []string
	if c, found := qParams["cluster"]; found {
		queryClusters = c
		for _, cl := range queryClusters {
			parts := strings.Split(cl, "+")
			if len(parts) != 2 {
				http.Error(w, "Invalid cluster query", http.StatusBadRequest)
				return
			}
			for _, p := range parts {
				errs := validation.IsValidName(p)
				if len(errs) > 0 {
					http.Error(w, "Invalid cluster query", http.StatusBadRequest)
					return
				}
			}
		}
	} else {
		queryClusters = make([]string, 0)
	}

	var queryResources []string
	if r, found := qParams["resource"]; found {
		queryResources = r
		for _, res := range queryResources {
			errs := validation.IsValidName(res)
			if len(errs) > 0 {
				http.Error(w, "Invalid resources query", http.StatusBadRequest)
				return
			}
		}
	} else {
		queryResources = make([]string, 0)
	}

	status, iErr := h.client.Status(p, ca, v, di, queryInstance, queryType, queryOutput, queryApps, queryClusters, queryResources)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	iErr = json.NewEncoder(w).Encode(status)
	if iErr != nil {
		http.Error(w, iErr.Error(), http.StatusInternalServerError)
		return
	}
}
