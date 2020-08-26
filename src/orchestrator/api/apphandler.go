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
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	"github.com/gorilla/mux"
)

var appJSONFile string = "json-schemas/metadata.json"

// appHandler to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type appHandler struct {
	// Interface that implements App operations
	// We will set this variable with a mock interface for testing
	client moduleLib.AppManager
}

// createAppHandler handles creation of the App entry in the database
// This is a multipart handler. See following example curl request
// curl -X POST http://localhost:9015/v2/projects/sampleProject/composite-apps/sampleCompositeApp/v1/apps \
// -F "metadata={\"metadata\":{\"name\":\"app\",\"description\":\"sample app\",\"UserData1\":\"data1\",\"UserData2\":\"data2\"}};type=application/json" \
// -F file=@/pathToFile

func (h appHandler) createAppHandler(w http.ResponseWriter, r *http.Request) {
	var a moduleLib.App
	var ac moduleLib.AppContent

	// Implemenation using multipart form
	// Set Max size to 16mb here
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&a)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(appJSONFile, a)
	if err != nil {
		http.Error(w, err.Error(), httpError)
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
	// Limit file Size to 1 GB
	if len(content) > 1073741824 {
		http.Error(w, "File Size Exceeds 1 GB", http.StatusUnprocessableEntity)
		return
	}
	err = validation.IsTarGz(bytes.NewBuffer(content))
	if err != nil {
		http.Error(w, "Error in file format", http.StatusUnprocessableEntity)
		return
	}

	ac.FileContent = base64.StdEncoding.EncodeToString(content)

	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	compositeAppVersion := vars["version"]

	ret, err := h.client.CreateApp(a, ac, projectName, compositeAppName, compositeAppVersion)
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

// getAppHandler handles GET operations on a particular App Name
// Returns an app
func (h appHandler) getAppHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	projectName := vars["project-name"]
	compositeAppName := vars["composite-app-name"]
	compositeAppVersion := vars["version"]
	name := vars["app-name"]

	// handle the get all apps case - return a list of only the json parts
	if len(name) == 0 {
		var retList []moduleLib.App

		ret, err := h.client.GetApps(projectName, compositeAppName, compositeAppVersion)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, app := range ret {
			retList = append(retList, moduleLib.App{Metadata: app.Metadata})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retList)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		return
	}

	accepted, _, err := mime.ParseMediaType(r.Header.Get("Accept"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	var retApp moduleLib.App
	var retAppContent moduleLib.AppContent

	retApp, err = h.client.GetApp(name, projectName, compositeAppName, compositeAppVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	retAppContent, err = h.client.GetAppContent(name, projectName, compositeAppName, compositeAppVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	switch accepted {
	case "multipart/form-data":
		mpw := multipart.NewWriter(w)
		w.Header().Set("Content-Type", mpw.FormDataContentType())
		w.WriteHeader(http.StatusOK)
		pw, err := mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/json"}, "Content-Disposition": {"form-data; name=metadata"}})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(pw).Encode(retApp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=file; filename=fileContent"}})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fcBytes, err := base64.StdEncoding.DecodeString(retAppContent.FileContent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = pw.Write(fcBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retApp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		fcBytes, err := base64.StdEncoding.DecodeString(retAppContent.FileContent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(fcBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "set Accept: multipart/form-data, application/json or application/octet-stream", http.StatusMultipleChoices)
		return
	}
}

// deleteAppHandler handles DELETE operations on a particular App Name
func (h appHandler) deleteAppHandler(w http.ResponseWriter, r *http.Request) {
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
