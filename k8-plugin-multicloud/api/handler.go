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
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	pkgerrors "github.com/pkg/errors"
	// "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"k8-plugin-multicloud/csar"
	"k8-plugin-multicloud/db"
	"k8-plugin-multicloud/krd"
)

// GetVNFClient retrieve the client used to communicate with a Kubernetes Cluster
var GetVNFClient = func(kubeConfigPath string) (kubernetes.Clientset, error) {
	client, err := krd.GetKubeClient(kubeConfigPath)
	if err != nil {
		return client, err
	}
	return client, nil
}

func validateBody(body interface{}) error {
	switch b := body.(type) {
	case CreateVnfRequest:
		if b.CloudRegionID == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing CloudRegionID in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if b.CsarID == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing CsarID in POST request"), "CreateVnfRequest bad request")
			return werr
		}
		if strings.Contains(b.CloudRegionID, "|") || strings.Contains(b.Namespace, "|") {
			werr := pkgerrors.Wrap(errors.New("Character \"|\" not allowed in CSAR ID"), "CreateVnfRequest bad request")
			return werr
		}
	case UpdateVnfRequest:
		if b.CloudRegionID == "" || b.CsarID == "" {
			werr := pkgerrors.Wrap(errors.New("Invalid/Missing Data in PUT request"), "UpdateVnfRequest bad request")
			return werr
		}
	}
	return nil
}

// CreateHandler is the POST method creates a new VNF instance resource.
func CreateHandler(w http.ResponseWriter, r *http.Request) {
	var resource CreateVnfRequest

	if r.Body == nil {
		http.Error(w, "Body empty", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	err = validateBody(resource)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// (TODO): Read kubeconfig for specific Cloud Region from local file system
	// if present or download it from AAI
	// err := DownloadKubeConfigFromAAI(resource.CloudRegionID, os.Getenv("KUBE_CONFIG_DIR")
	kubeclient, err := GetVNFClient(os.Getenv("KUBE_CONFIG_DIR") + "/" + resource.CloudRegionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	/*
		uuid,
		{
			"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
			"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
		},
		nil
	*/
	externalVNFID, resourceNameMap, err := csar.CreateVNF(resource.CsarID, resource.CloudRegionID, resource.Namespace, &kubeclient)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Read Kubernetes Data information error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	// cloud1-default-uuid
	internalVNFID := resource.CloudRegionID + "-" + resource.Namespace + "-" + externalVNFID

	// Persist in AAI database.
	log.Printf("Cloud Region ID: %s, Namespace: %s, VNF ID: %s ", resource.CloudRegionID, resource.Namespace, externalVNFID)

	// krd.AddNetworkAnnotationsToPod(kubeData, resource.Networks)

	// "{"deployment":<>,"service":<>}"
	out, err := json.Marshal(resourceNameMap)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}
	serializedResourceNameMap := string(out)
	log.Println(serializedResourceNameMap)

	// key: cloud1-default-uuid
	// value: "{"deployment":<>,"service":<>}"
	err = db.DBconn.CreateEntry(internalVNFID, serializedResourceNameMap)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Create VNF deployment error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	resp := CreateVnfResponse{
		VNFID:         externalVNFID,
		CloudRegionID: resource.CloudRegionID,
		Namespace:     resource.Namespace,
		VNFComponents: resourceNameMap,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}
}

// ListHandler the existing VNF instances created in a given Kubernetes cluster
func ListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cloudRegionID := vars["cloudRegionID"]
	namespace := vars["namespace"]
	prefix := cloudRegionID + "-" + namespace

	internalVNFIDs, err := db.DBconn.ReadAll(prefix)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	if len(internalVNFIDs) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// TODO: There is an edge case where if namespace is passed but is missing some characters
	// trailing, it will print the result with those excluding characters. This is because of
	// the way I am trimming the Prefix. This fix is needed.

	var editedList []string

	for _, id := range internalVNFIDs {
		if len(id) > 0 {
			editedList = append(editedList, strings.TrimPrefix(id, prefix)[1:])
		}
	}

	if len(editedList) == 0 {
		editedList = append(editedList, "")
	}

	resp := ListVnfsResponse{
		VNFs: editedList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output VNF list error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}

}

// DeleteHandler method terminates an individual VNF instance.
func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cloudRegionID := vars["cloudRegionID"] // cloud1
	namespace := vars["namespace"]         // default
	externalVNFID := vars["externalVNFID"] // uuid

	// cloud1-default-uuid
	internalVNFID := cloudRegionID + "-" + namespace + "-" + externalVNFID

	// (TODO): Read kubeconfig for specific Cloud Region from local file system
	// if present or download it from AAI
	// err := DownloadKubeConfigFromAAI(resource.CloudRegionID, os.Getenv("KUBE_CONFIG_DIR")
	kubeclient, err := GetVNFClient(os.Getenv("KUBE_CONFIG_DIR") + "/" + cloudRegionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// key: cloud1-default-uuid
	// value: "{"deployment":<>,"service":<>}"
	serializedResourceNameMap, found, err := db.DBconn.ReadEntry(internalVNFID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if found == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	/*
		{
			"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
			"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
		},
	*/
	deserializedResourceNameMap := make(map[string][]string)
	err = json.Unmarshal([]byte(serializedResourceNameMap), &deserializedResourceNameMap)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Delete VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	err = csar.DestroyVNF(deserializedResourceNameMap, namespace, &kubeclient)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Delete VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	err = db.DBconn.DeleteEntry(internalVNFID)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Delete VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
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

// GetHandler retrieves information about a VNF instance by reading an individual VNF instance resource.
func GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	cloudRegionID := vars["cloudRegionID"] // cloud1
	namespace := vars["namespace"]         // default
	externalVNFID := vars["externalVNFID"] // uuid

	// cloud1-default-uuid
	internalVNFID := cloudRegionID + "-" + namespace + "-" + externalVNFID

	// key: cloud1-default-uuid
	// value: "{"deployment":<>,"service":<>}"
	serializedResourceNameMap, found, err := db.DBconn.ReadEntry(internalVNFID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if found == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	/*
		{
			"deployment": ["cloud1-default-uuid-sisedeploy1", "cloud1-default-uuid-sisedeploy2", ... ]
			"service": ["cloud1-default-uuid-sisesvc1", "cloud1-default-uuid-sisesvc2", ... ]
		},
	*/
	deserializedResourceNameMap := make(map[string][]string)
	err = json.Unmarshal([]byte(serializedResourceNameMap), &deserializedResourceNameMap)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Get VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
		return
	}

	resp := GetVnfResponse{
		VNFID:         externalVNFID,
		CloudRegionID: cloudRegionID,
		Namespace:     namespace,
		VNFComponents: deserializedResourceNameMap,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		werr := pkgerrors.Wrap(err, "Parsing output of new VNF error")
		http.Error(w, werr.Error(), http.StatusInternalServerError)
	}
}
