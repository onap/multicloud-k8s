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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)

// appHandler to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type appHandler struct {
	// Interface that implements App operations
	// We will set this variable with a mock interface for testing
	client moduleLib.AppManager
}

// createHandler handles creation of the App entry in the database
// This is a multipart handler. See following example curl request
// curl -X POST http://localhost:9015/v2/projects/sampleProject/composite-apps/sampleCompositeApp/v1/apps \
// -F "meta={\"metadata\":{\"name\":\"app\",\"description\":\"sample app\",\"UserData1\":\"data1\",\"UserData2\":\"data2\"}};type=application/json" \
// -F file=@/pathToFile

func (h appHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	var a moduleLib.App

	// Implemenation using multipart form
	// Set Max size to 16mb here
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("meta")))
	err = json.NewDecoder(jsn).Decode(&a)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Name is required.
	if a.Metadata.Name == "" {
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

	//Convert the file content to base64 for storage
	content, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusUnprocessableEntity)
		return
	}

	a.Metadata.File = base64.StdEncoding.EncodeToString(content)

	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	compositeAppVersion := vars["version"]

	ret, err := h.client.CreateApp(a, projectName, compositeAppName, compositeAppVersion)
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

// getHandler handles GET operations on a particular App Name
// Returns an app
func (h appHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	name := vars["app-name"]

	ret, err := h.client.GetApp(name, projectName, compositeAppName, compositeAppVersion)
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

// deleteHandler handles DELETE operations on a particular App Name
func (h appHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	name := vars["app-name"]

	err := h.client.DeleteApp(name, projectName, compositeAppName, compositeAppVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
