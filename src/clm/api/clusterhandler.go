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

	clusterPkg "github.com/onap/multicloud-k8s/src/clm/pkg/cluster"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"

	"github.com/gorilla/mux"
)

var cpJSONFile string = "json-schemas/metadata.json"
var ckvJSONFile string = "json-schemas/cluster-kv.json"
var clJSONFile string = "json-schemas/cluster-label.json"


// Used to store backend implementations objects
// Also simplifies mocking for unit testing purposes
type clusterHandler struct {
	// Interface that implements Cluster operations
	// We will set this variable with a mock interface for testing
	client clusterPkg.ClusterManager
}

// Create handles creation of the ClusterProvider entry in the database
func (h clusterHandler) createClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	var p clusterPkg.ClusterProvider

	err := json.NewDecoder(r.Body).Decode(&p)

	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(cpJSONFile, p)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	// Name is required.
	if p.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterProvider(p)
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

// Get handles GET operations on a particular ClusterProvider Name
// Returns a ClusterProvider
func (h clusterHandler) getClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var ret interface{}
	var err error

	if len(name) == 0 {
		ret, err = h.client.GetClusterProviders()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetClusterProvider(name)
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

// Delete handles DELETE operations on a particular ClusterProvider  Name
func (h clusterHandler) deleteClusterProviderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	err := h.client.DeleteClusterProvider(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Create handles creation of the Cluster entry in the database
func (h clusterHandler) createClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	var p clusterPkg.Cluster
	var q clusterPkg.ClusterContent

	// Implemenation using multipart form
	// Review and enable/remove at a later date
	// Set Max size to 16mb here
	err := r.ParseMultipartForm(16777216)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	jsn := bytes.NewBuffer([]byte(r.FormValue("metadata")))
	err = json.NewDecoder(jsn).Decode(&p)
	switch {
	case err == io.EOF:
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err, httpError := validation.ValidateJsonSchemaData(cpJSONFile, p)
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

	q.Kubeconfig = base64.StdEncoding.EncodeToString(content)

	// Name is required.
	if p.Metadata.Name == "" {
		http.Error(w, "Missing name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateCluster(provider, p, q)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	//	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(ret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Get handles GET operations on a particular Cluster Name
// Returns a Cluster
func (h clusterHandler) getClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	name := vars["name"]

	label := r.URL.Query().Get("label")
	if len(label) != 0 {
		ret, err := h.client.GetClustersWithLabel(provider, label)
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
		return
	}

	// handle the get all clusters case - return a list of only the json parts
	if len(name) == 0 {
		var retList []clusterPkg.Cluster

		ret, err := h.client.GetClusters(provider)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, cl := range ret {
			retList = append(retList, clusterPkg.Cluster{Metadata: cl.Metadata})
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

	retCluster, err := h.client.GetCluster(provider, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	retKubeconfig, err := h.client.GetClusterContent(provider, name)
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
		if err := json.NewEncoder(pw).Encode(retCluster); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pw, err = mpw.CreatePart(textproto.MIMEHeader{"Content-Type": {"application/octet-stream"}, "Content-Disposition": {"form-data; name=file; filename=kubeconfig"}})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		kcBytes, err := base64.StdEncoding.DecodeString(retKubeconfig.Kubeconfig)
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
		err = json.NewEncoder(w).Encode(retCluster)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "application/octet-stream":
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		kcBytes, err := base64.StdEncoding.DecodeString(retKubeconfig.Kubeconfig)
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

// Delete handles DELETE operations on a particular Cluster Name
func (h clusterHandler) deleteClusterHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	name := vars["name"]

	err := h.client.DeleteCluster(provider, name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Create handles creation of the ClusterLabel entry in the database
func (h clusterHandler) createClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	var p clusterPkg.ClusterLabel

	err := json.NewDecoder(r.Body).Decode(&p)

	err, httpError := validation.ValidateJsonSchemaData(clJSONFile, p)
	if err != nil {
		http.Error(w, err.Error(), httpError)
		return
	}

	// LabelName is required.
	if p.LabelName == "" {
		http.Error(w, "Missing label name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterLabel(provider, cluster, p)
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

// Get handles GET operations on a particular Cluster Label
// Returns a ClusterLabel
func (h clusterHandler) getClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	label := vars["label"]

	var ret interface{}
	var err error

	if len(label) == 0 {
		ret, err = h.client.GetClusterLabels(provider, cluster)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetClusterLabel(provider, cluster, label)
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

// Delete handles DELETE operations on a particular ClusterLabel Name
func (h clusterHandler) deleteClusterLabelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	label := vars["label"]

	err := h.client.DeleteClusterLabel(provider, cluster, label)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Create handles creation of the ClusterKvPairs entry in the database
func (h clusterHandler) createClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	var p clusterPkg.ClusterKvPairs

	err := json.NewDecoder(r.Body).Decode(&p)

		// Verify JSON Body
		err, httpError := validation.ValidateJsonSchemaData(ckvJSONFile, p)
		if err != nil {
			http.Error(w, err.Error(), httpError)
			return
		}

	// KvPairsName is required.
	if p.Metadata.Name == "" {
		http.Error(w, "Missing Key Value pair name in POST request", http.StatusBadRequest)
		return
	}

	ret, err := h.client.CreateClusterKvPairs(provider, cluster, p)
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

// Get handles GET operations on a particular Cluster Key Value Pair
// Returns a ClusterKvPairs
func (h clusterHandler) getClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	kvpair := vars["kvpair"]

	var ret interface{}
	var err error

	if len(kvpair) == 0 {
		ret, err = h.client.GetAllClusterKvPairs(provider, cluster)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		ret, err = h.client.GetClusterKvPairs(provider, cluster, kvpair)
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

// Delete handles DELETE operations on a particular Cluster Name
func (h clusterHandler) deleteClusterKvPairsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider-name"]
	cluster := vars["cluster-name"]
	kvpair := vars["kvpair"]

	err := h.client.DeleteClusterKvPairs(provider, cluster, kvpair)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
