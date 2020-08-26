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
	pkgerrors "github.com/pkg/errors"
)

var appProfileJSONFile string = "json-schemas/metadata.json"

/* Used to store backend implementation objects
Also simplifies mocking for unit testing purposes
*/
type appProfileHandler struct {
	client moduleLib.AppProfileManager
}

// createAppProfileHandler handles the create operation
func (h appProfileHandler) createAppProfileHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	compositeProfile := vars["composite-profile-name"]

	var ap moduleLib.AppProfile
	var ac moduleLib.AppProfileContent

	// Implemenation using multipart form
	// Review and enable/remove at a later date
	// Set Max size to 16mb here
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&ap)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Verify JSON Body
	err, httpError := validation.ValidateJsonSchemaData(appProfileJSONFile, ap)
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

	ac.Profile = base64.StdEncoding.EncodeToString(content)

	// Name is required.
	if ap.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, ap, ac)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler handles the GET operations on AppProfile
func (h appProfileHandler) getAppProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	compositeProfile := vars["composite-profile-name"]
	name := vars["app-profile"]
	appName := r.URL.Query().Get("app-name")

	if len(name) != 0 && len(appName) != 0 {
		http.Error(w, pkgerrors.New("Invalid query").Error(), http.StatusInternalServerError)
		return
	}

	// handle the get all app profiles case - return a list of only the json parts
	if len(name) == 0 && len(appName) == 0 {
		var retList []moduleLib.AppProfile

		ret, err := h.client.GetAppProfiles(project, compositeApp, compositeAppVersion, compositeProfile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, ap := range ret {
			retList = append(retList, ap)
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

	var retAppProfile moduleLib.AppProfile
	var retAppProfileContent moduleLib.AppProfileContent

	if len(appName) != 0 {
		retAppProfile, err = h.client.GetAppProfileByApp(project, compositeApp, compositeAppVersion, compositeProfile, appName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		retAppProfileContent, err = h.client.GetAppProfileContentByApp(project, compositeApp, compositeAppVersion, compositeProfile, appName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		retAppProfile, err = h.client.GetAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		retAppProfileContent, err = h.client.GetAppProfileContent(project, compositeApp, compositeAppVersion, compositeProfile, name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
		if err := json.NewEncoder(pw).Encode(retAppProfile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=file"}})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kcBytes, err := base64.StdEncoding.DecodeString(retAppProfileContent.Profile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = pw.Write(kcBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/json":
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(retAppProfile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		kcBytes, err := base64.StdEncoding.DecodeString(retAppProfileContent.Profile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = w.Write(kcBytes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "set Accept: multipart/form-data, application/json or application/octet-stream", http.StatusMultipleChoices)
		return
	}
}

// deleteHandler handles the delete operations on AppProfile
func (h appProfileHandler) deleteAppProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	project := vars["project-name"]
	compositeApp := vars["composite-app-name"]
	compositeAppVersion := vars["composite-app-version"]
	compositeProfile := vars["composite-profile-name"]
	name := vars["app-profile"]

	err := h.client.DeleteAppProfile(project, compositeApp, compositeAppVersion, compositeProfile, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
