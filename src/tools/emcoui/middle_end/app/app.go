/*
=======================================================================
Copyright (c) 2017-2020 Aarna Networks, Inc.
All rights reserved.
======================================================================
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
          http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
========================================================================
*/

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/mux"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type deployServiceData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Spec        struct {
		ProjectName string     `json:"projectName"`
		Apps        []appsData `json:"appsData"`
	} `json:"spec"`
}

type deployDigData struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	CompositeAppName    string `json:"compositeApp"`
	CompositeProfile    string `json:"compositeProfile"`
	DigVersion          string `json:"version"`
	CompositeAppVersion string `json:"compositeAppVersion"`
	Spec                struct {
		ProjectName string     `json:"projectName"`
		Apps        []appsData `json:"appsData"`
	} `json:"spec"`
}

// Exists is for mongo $exists filter
type Exists struct {
	Exists string `json:"$exists"`
}

// This is the json payload that the  orchesration API expexts.
type appsData struct {
	Metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		FileName    string `json:"filename"`
	} `json:"metadata"`
	ProfileMetadata struct {
		Name     string `json:"name"`
		FileName string `json:"filename"`
	} `json:"profileMetadata"`
	Clusters []struct {
		Provider         string `json:"provider"`
		SelectedClusters []struct {
			Name       string `json:"name"`
			Interfaces []struct {
				NetworkName string `json:"networkName"`
				IP          string `json:"ip"`
				Subnet      string `json:"subnet"`
			} `json:"interfaces"`
		} `json:"selectedClusters"`
	} `json:"clusters"`
}

type DigsInProject struct {
	Metadata struct {
		Name                string `json:"name"`
		CompositeAppName    string `json:"compositeAppName"`
		CompositeAppVersion string `json:"compositeAppVersion"`
		Description         string `json:"description"`
		UserData1           string `userData1:"userData1"`
		UserData2           string `userData2:"userData2"`
	} `json:"metadata"`
	Spec struct {
		DigIntentsData    []DigDeployedIntents `json:"deployedIntents"`
		Profile           string               `json:"profile"`
		Version           string               `json:"version"`
		Lcloud            string               `json:"logicalCloud"`
		OverrideValuesObj []OverrideValues     `json:"overrideValues"`
		GpintArray        []*DigsGpint         `json:"GenericPlacementIntents,omitempty"`
		NwintArray        []*DigsNwint         `json:"networkCtlIntents,omitempty"`
	} `json:"spec"`
}

type DigsGpint struct {
	Metadata apiMetaData `json:"metadata,omitempty"`
	Spec     struct {
		AppIntentArray []PlacementIntentExport `json:"placementIntent,omitempty"`
	} `json:"spec,omitempty"`
}

type DigsNwint struct {
	Metadata apiMetaData `json:"metadata,omitempty"`
	Spec     struct {
		WorkloadIntentsArray []*WorkloadIntents `json:"WorkloadIntents,omitempty"`
	} `json:"spec,omitempty"`
}
type WorkloadIntents struct {
	Metadata apiMetaData `json:"metadata,omitempty"`
	Spec     struct {
		Interfaces []NwInterface `json:"interfaces,omitempty"`
	} `json:"spec,omitempty"`
}

// Project Tree
type ProjectTree struct {
	Metadata        ProjectMetadata
	compositeAppMap map[string]*CompositeAppTree
}

type treeTraverseFilter struct {
	compositeAppName    string
	compositeAppVersion string
	digName             string
}

// Composite app tree
type CompositeAppTree struct {
	Metadata         CompositeApp
	AppsDataArray    map[string]*AppsData
	ProfileDataArray map[string]*ProfilesData
	DigMap           map[string]*DigReadData
}

type DigReadData struct {
	DigpData       DeploymentIGP
	DigIntentsData DigpIntents
	GpintMap       map[string]*GpintData
	NwintMap       map[string]*NwintData
}

type GpintData struct {
	Gpint          GenericPlacementIntent
	AppIntentArray []PlacementIntent
}

type NwintData struct {
	Nwint     NetworkCtlIntent
	WrkintMap map[string]*WrkintData
}

type WrkintData struct {
	Wrkint     NetworkWlIntent
	Interfaces []NwInterface
}

type AppsData struct {
	App              CompositeApp
	CompositeProfile ProfileMeta
}

type ProfilesData struct {
	Profile     ProfileMeta
	AppProfiles []ProfileMeta
}

type ClusterMetadata struct {
	Metadata apiMetaData `json:"Metadata"`
}

type apiMetaData struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `userData1:"userData1"`
	UserData2   string `userData2:"userData2"`
}

// The interface
type orchWorkflow interface {
	createAnchor() interface{}
	createObject() interface{}
	getObject() (interface{}, interface{})
	getAnchor() (interface{}, interface{})
	deleteObject() interface{}
	deleteAnchor() interface{}
}

// MiddleendConfig The configmap of the middleent
type MiddleendConfig struct {
	OwnPort     string `json:"ownport"`
	Clm         string `json:"clm"`
	OrchService string `json:"orchestrator"`
	OvnService  string `json:"ovnaction"`
	Mongo       string `json:"mongo"`
}

// OrchestrationHandler interface, handling the composite app APIs
type OrchestrationHandler struct {
	MiddleendConf    MiddleendConfig
	client           http.Client
	compositeAppName string
	compositeAppDesc string
	AppName          string
	meta             []appsData
	DigData          deployDigData
	file             map[string]*multipart.FileHeader
	dataRead         *ProjectTree
	treeFilter       *treeTraverseFilter
	DigpReturnJson   []DigsInProject
	projectName      string
	projectDesc      string
	version          string
	response         struct {
		payload map[string][]byte
		status  map[string]int
	}
	digpIntents  map[string]string
	nwCtlIntents map[string]string
}

// NewAppHandler interface implementing REST callhandler
func NewAppHandler() *OrchestrationHandler {
	return &OrchestrationHandler{}
}

// GetHealth to check connectivity
func (h OrchestrationHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

}

func (h OrchestrationHandler) apiGet(url string, statusKey string) (interface{}, []byte, error) {
	// prepare and DEL API
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := h.client.Do(request)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)
	h.response.payload[statusKey] = data
	h.response.status[statusKey] = resp.StatusCode

	return resp.StatusCode, data, nil
}

func (h OrchestrationHandler) apiDel(url string, statusKey string) (interface{}, error) {
	// prepare and DEL API
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)
	h.response.payload[statusKey] = data
	h.response.status[statusKey] = resp.StatusCode

	return resp.StatusCode, nil
}

func (h OrchestrationHandler) apiPost(jsonLoad []byte, url string, statusKey string) (interface{}, error) {
	// prepare and POST API
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonLoad))
	if err != nil {
		return nil, err
	}
	resp, err := h.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)
	h.response.payload[statusKey] = data
	h.response.status[statusKey] = resp.StatusCode

	return resp.StatusCode, nil
}

func (h OrchestrationHandler) apiPostMultipart(jsonLoad []byte,
	fh *multipart.FileHeader, url string, statusKey string, fileName string) (interface{}, error) {
	// Open the file
	file, err := fh.Open()
	if err != nil {
		return nil, err
	}
	// Close the file later
	defer file.Close()
	// Buffer to store our request body as bytes
	var requestBody bytes.Buffer
	// Create a multipart writer
	multiPartWriter := multipart.NewWriter(&requestBody)
	// Initialize the file field. Arguments are the field name and file name
	// It returns io.Writer
	fileWriter, err := multiPartWriter.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}
	// Copy the actual file content to the field field's writer
	_, err = io.Copy(fileWriter, file)
	if err != nil {
		return nil, err
	}
	// Populate other fields
	fieldWriter, err := multiPartWriter.CreateFormField("metadata")
	if err != nil {
		return nil, err
	}

	_, err = fieldWriter.Write([]byte(jsonLoad))
	if err != nil {
		return nil, err
	}

	// We completed adding the file and the fields, let's close the multipart writer
	// So it writes the ending boundary
	multiPartWriter.Close()

	// By now our original request body should have been populated,
	// so let's just use it with our custom request
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, err
	}
	// We need to set the content type from the writer, it includes necessary boundary as well
	req.Header.Set("Content-Type", multiPartWriter.FormDataContentType())

	// Do the request
	resp, err := h.client.Do(req)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	defer resp.Body.Close()
	// Prepare the response
	data, _ := ioutil.ReadAll(resp.Body)
	h.response.payload[statusKey] = data
	h.response.status[statusKey] = resp.StatusCode

	return resp.StatusCode, nil
}
func (h *OrchestrationHandler) prepTreeReq(vars map[string]string) {
	// Initialise the project tree with target composite application.
	h.treeFilter = &treeTraverseFilter{}
	h.treeFilter.compositeAppName = vars["composite-app-name"]
	h.treeFilter.compositeAppVersion = vars["version"]
	h.treeFilter.digName = vars["deployment-intent-group-name"]
}

func (h *OrchestrationHandler) DelDig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.projectName = vars["project-name"]
	h.treeFilter = nil

	dataPoints := []string{"projectHandler", "compAppHandler",
		"digpHandler",
		"placementIntentHandler",
		"networkIntentHandler"}
	h.response.status = make(map[string]int)
	h.response.payload = make(map[string][]byte)

	// Initialise the project tree with target composite application.
	h.prepTreeReq(vars)

	h.dataRead = &ProjectTree{}
	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	// 1. Call DIG delte workflow
	fmt.Printf("Delete wflow start")
	deleteDataPoints := []string{"networkIntentHandler",
		"placementIntentHandler",
		"digpHandler"}
	retcode = h.deleteTree(deleteDataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	w.WriteHeader(204)
}

// Delete service workflow
func (h *OrchestrationHandler) DelSvc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.projectName = vars["project-name"]
	h.treeFilter = nil

	dataPoints := []string{"projectHandler", "compAppHandler",
		"digpHandler",
		"ProfileHandler"}
	h.response.status = make(map[string]int)
	h.response.payload = make(map[string][]byte)

	// Initialise the project tree with target composite application.
	h.prepTreeReq(vars)

	h.dataRead = &ProjectTree{}
	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	fmt.Printf("tree %+v\n", h.dataRead)
	// Check if a dig is present in this composite application
	if len(h.dataRead.compositeAppMap[vars["composite-app-name"]].DigMap) != 0 {
		w.WriteHeader(409)
		w.Write([]byte("Non emtpy DIG in service\n"))
		return
	}

	// 1. Call delte workflow
	fmt.Printf("Delete wflow start")
	deleteDataPoints := []string{"ProfileHandler",
		"compAppHandler"}
	retcode = h.deleteTree(deleteDataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	w.WriteHeader(204)
}

func (h *OrchestrationHandler) getData(I orchWorkflow) (interface{}, interface{}) {
	_, retcode := I.getAnchor()
	if retcode != 200 {
		return nil, retcode
	}
	dataPointData, retcode := I.getObject()
	if retcode != 200 {
		return nil, retcode
	}
	return dataPointData, retcode
}

func (h *OrchestrationHandler) deleteData(I orchWorkflow) (interface{}, interface{}) {
	_ = I.deleteObject()
	_ = I.deleteAnchor()
	return nil, 204 //FIXME
}

func (h *OrchestrationHandler) deleteTree(dataPoints []string) interface{} {
	//1. Fetch App data
	var I orchWorkflow
	for _, dataPoint := range dataPoints {
		switch dataPoint {
		case "projectHandler":
			temp := &projectHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.deleteData(I)
			if retcode != 204 {
				return retcode
			}
			break
		case "compAppHandler":
			temp := &compAppHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.deleteData(I)
			if retcode != 204 {
				return retcode
			}
			break
		case "ProfileHandler":
			temp := &ProfileHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.deleteData(I)
			if retcode != 204 {
				return retcode
			}
			break
		case "digpHandler":
			temp := &digpHandler{}
			temp.orchInstance = h
			I = temp
			fmt.Printf("delete digp\n")
			_, retcode := h.deleteData(I)
			if retcode != 204 {
				return retcode
			}
			break
		case "placementIntentHandler":
			temp := &placementIntentHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.deleteData(I)
			if retcode != 204 {
				return retcode
			}
			break
		case "networkIntentHandler":
			temp := &networkIntentHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.deleteData(I)
			if retcode != 204 {
				return retcode
			}
			break
		default:
			fmt.Printf("%s\n", dataPoint)
		}
	}
	return nil
}

func (h *OrchestrationHandler) constructTree(dataPoints []string) interface{} {
	//1. Fetch App data
	var I orchWorkflow
	for _, dataPoint := range dataPoints {
		switch dataPoint {
		case "projectHandler":
			temp := &projectHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.getData(I)
			if retcode != 200 {
				return retcode
			}
			break
		case "compAppHandler":
			temp := &compAppHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.getData(I)
			if retcode != 200 {
				return retcode
			}
			break
		case "ProfileHandler":
			temp := &ProfileHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.getData(I)
			if retcode != 200 {
				return retcode
			}
			break
		case "digpHandler":
			temp := &digpHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.getData(I)
			if retcode != 200 {
				return retcode
			}
			break
		case "placementIntentHandler":
			temp := &placementIntentHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.getData(I)
			if retcode != 200 {
				return retcode
			}
			break
		case "networkIntentHandler":
			temp := &networkIntentHandler{}
			temp.orchInstance = h
			I = temp
			_, retcode := h.getData(I)
			if retcode != 200 {
				return retcode
			}
			break
		default:
			fmt.Printf("%s\n", dataPoint)
		}
	}
	return nil
}

// This function partest he compositeapp tree read and populates the
// Dig tree
func (h *OrchestrationHandler) copyDigTree() {
	dataRead := h.dataRead
	h.DigpReturnJson = nil

	for compositeAppName, value := range dataRead.compositeAppMap {
		for _, digValue := range dataRead.compositeAppMap[compositeAppName].DigMap {
			Dig := DigsInProject{}
			SourceDigMetadata := digValue.DigpData.Metadata

			// Copy the metadata
			Dig.Metadata.Name = SourceDigMetadata.Name
			Dig.Metadata.CompositeAppName = compositeAppName
			Dig.Metadata.CompositeAppVersion = value.Metadata.Spec.Version
			Dig.Metadata.Description = SourceDigMetadata.Description
			Dig.Metadata.UserData1 = SourceDigMetadata.UserData1
			Dig.Metadata.UserData2 = SourceDigMetadata.UserData2

			// Populate the Spec of dig
			SourceDigSpec := digValue.DigpData.Spec
			Dig.Spec.DigIntentsData = digValue.DigIntentsData.Intent
			Dig.Spec.Profile = SourceDigSpec.Profile
			Dig.Spec.Version = SourceDigSpec.Version
			Dig.Spec.Lcloud = SourceDigSpec.Lcloud
			Dig.Spec.OverrideValuesObj = SourceDigSpec.OverrideValuesObj

			// Pupolate the generic placement intents
			SourceGpintMap := digValue.GpintMap
			for t, gpintValue := range SourceGpintMap {
				fmt.Printf("gpName value %s\n", t)
				localGpint := DigsGpint{}
				localGpint.Metadata = gpintValue.Gpint.Metadata
				//localGpint.Spec.AppIntentArray = gpintValue.AppIntentArray
				localGpint.Spec.AppIntentArray = make([]PlacementIntentExport, len(gpintValue.AppIntentArray))
				for k, _ := range gpintValue.AppIntentArray {
					localGpint.Spec.AppIntentArray[k].Metadata = gpintValue.AppIntentArray[k].Metadata
					localGpint.Spec.AppIntentArray[k].Spec.AppName =
						gpintValue.AppIntentArray[k].Spec.AppName
					localGpint.Spec.AppIntentArray[k].Spec.Intent.AllofCluster =
						make([]AllofExport, len(gpintValue.AppIntentArray[k].Spec.Intent.AllofCluster))
					for i, _ := range gpintValue.AppIntentArray[k].Spec.Intent.AllofCluster {
						localGpint.Spec.AppIntentArray[k].Spec.Intent.AllofCluster[i].ProviderName =
							gpintValue.AppIntentArray[k].Spec.Intent.AllofCluster[i].ProviderName
						localGpint.Spec.AppIntentArray[k].Spec.Intent.AllofCluster[i].ClusterName =
							gpintValue.AppIntentArray[k].Spec.Intent.AllofCluster[i].ClusterName
					}
				}

				Dig.Spec.GpintArray = append(Dig.Spec.GpintArray, &localGpint)
			}
			// Populate the Nwint intents
			SourceNwintMap := digValue.NwintMap
			for _, nwintValue := range SourceNwintMap {
				localNwint := DigsNwint{}
				localNwint.Metadata = nwintValue.Nwint.Metadata
				for _, wrkintValue := range nwintValue.WrkintMap {
					localWrkint := WorkloadIntents{}
					localWrkint.Metadata = wrkintValue.Wrkint.Metadata
					localWrkint.Spec.Interfaces = wrkintValue.Interfaces
					localNwint.Spec.WorkloadIntentsArray = append(localNwint.Spec.WorkloadIntentsArray,
						&localWrkint)
				}
				Dig.Spec.NwintArray = append(Dig.Spec.NwintArray, &localNwint)
			}
			h.DigpReturnJson = append(h.DigpReturnJson, Dig)
		}
	}
}

// GetSvc get the entrire tree under project/<composite app>/<version>
func (h *OrchestrationHandler) GetAllDigs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.version = vars["version"]
	h.projectName = vars["project-name"]
	h.response.status = make(map[string]int)
	h.response.payload = make(map[string][]byte)
	dataPoints := []string{"projectHandler", "compAppHandler",
		"digpHandler",
		"placementIntentHandler",
		"networkIntentHandler"}

	h.dataRead = &ProjectTree{}
	h.treeFilter = nil
	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	// copy dig tree
	h.copyDigTree()
	retval, _ := json.Marshal(h.DigpReturnJson)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(retval)
}

// GetSvc get the entrire tree under project/<composite app>/<version>
func (h *OrchestrationHandler) GetSvc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	h.treeFilter = nil
	h.compositeAppName = vars["composite-app-name"]
	h.version = vars["version"]
	h.projectName = vars["project-name"]
	h.response.status = make(map[string]int)
	h.response.payload = make(map[string][]byte)

	dataPoints := []string{"compAppHandler", "ProfileHandler",
		"digpHandler",
		"placementIntentHandler",
		"networkIntentHandler"}
	h.dataRead = &ProjectTree{}
	retcode := h.constructTree(dataPoints)
	if retcode != nil {
		if intval, ok := retcode.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
}

// CreateApp exported function which creates the composite application
func (h *OrchestrationHandler) CreateDig(w http.ResponseWriter, r *http.Request) {
	var jsonData deployDigData

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&jsonData)
	if err != nil {
		log.Printf("Failed to parse json")
		log.Fatalln(err)
	}

	h.DigData = jsonData

	if len(h.DigData.Spec.Apps) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Bad request, no app metadata\n"))
		return
	}

	h.client = http.Client{}

	// These maps will get populated by the return status and respones of each V2 API
	// that is called during the execution of the workflow.
	h.response.payload = make(map[string][]byte)
	h.response.status = make(map[string]int)

	// 4. Create DIG
	h.digpIntents = make(map[string]string)
	h.nwCtlIntents = make(map[string]string)
	igHandler := &digpHandler{}
	igHandler.orchInstance = h
	igpStatus := createDInents(igHandler)
	if igpStatus != nil {
		if intval, ok := igpStatus.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		w.Write(h.response.payload[h.compositeAppName+"_digp"])
		return
	}

	// 3. Create intents
	intentHandler := &placementIntentHandler{}
	intentHandler.orchInstance = h
	intentStatus := addPlacementIntent(intentHandler)
	if intentStatus != nil {
		if intval, ok := intentStatus.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		w.Write(h.response.payload[h.compositeAppName+"_gpint"])
		return
	}

	// If the metadata contains network interface request then call the
	// network intent related part of the workflow.
	if len(h.DigData.Spec.Apps[0].Clusters[0].SelectedClusters[0].Interfaces) != 0 {
		nwHandler := &networkIntentHandler{}
		nwHandler.orchInstance = h
		nwIntentStatus := addNetworkIntent(nwHandler)
		if nwIntentStatus != nil {
			if intval, ok := nwIntentStatus.(int); ok {
				w.WriteHeader(intval)
			} else {
				w.WriteHeader(500)
			}
			w.Write(h.response.payload[h.compositeAppName+"_nwctlint"])
			return
		}
	}

	w.WriteHeader(201)
	w.Write(h.response.payload[h.DigData.Name])
}

func (h *OrchestrationHandler) CreateApp(w http.ResponseWriter, r *http.Request) {
	var jsonData deployServiceData

	err := r.ParseMultipartForm(16777216)
	if err != nil {
		log.Fatalln(err)
	}

	// Populate the multipart.FileHeader MAP. The key will be the
	// filename itself. The metadata Map will be keyed on the application
	// name. The metadata has a field file name, so later we can parse the metadata
	// Map, and fetch the file headers from this file Map with keys as the filename.
	h.file = make(map[string]*multipart.FileHeader)
	for _, v := range r.MultipartForm.File {
		fh := v[0]
		h.file[fh.Filename] = fh
	}

	jsn := ([]byte(r.FormValue("servicePayload")))
	err = json.Unmarshal(jsn, &jsonData)
	if err != nil {
		log.Printf("Failed to parse json")
		log.Fatalln(err)
	}

	h.compositeAppName = jsonData.Name
	h.compositeAppDesc = jsonData.Description
	h.projectName = jsonData.Spec.ProjectName
	h.meta = jsonData.Spec.Apps

	// Sanity check. For each metadata there should be a
	// corresponding file in the multipart request. If it
	// not found we fail this API call.
	for i := range h.meta {
		switch {
		case h.file[h.meta[i].Metadata.FileName] == nil:
			t := fmt.Sprintf("File %s not in request", h.meta[i].Metadata.FileName)
			w.WriteHeader(400)
			w.Write([]byte(t))
			fmt.Printf("app file not found\n")
			return
		case h.file[h.meta[i].ProfileMetadata.FileName] == nil:
			t := fmt.Sprintf("File %s not in request", h.meta[i].ProfileMetadata.FileName)
			w.WriteHeader(400)
			w.Write([]byte(t))
			fmt.Printf("profile file not found\n")
			return
		default:
			fmt.Println("Good request")
		}
	}

	if len(h.meta) == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Bad request, no app metadata\n"))
		return
	}

	h.client = http.Client{}

	// These maps will get populated by the return status and respones of each V2 API
	// that is called during the execution of the workflow.
	h.response.payload = make(map[string][]byte)
	h.response.status = make(map[string]int)

	// 1. create the composite application. the compAppHandler implements the
	// orchWorkflow interface.
	appHandler := &compAppHandler{}
	appHandler.orchInstance = h
	appStatus := createCompositeapp(appHandler)
	if appStatus != nil {
		if intval, ok := appStatus.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		w.Write(h.response.payload[h.compositeAppName])
		return
	}

	// 2. create the composite application profiles
	profileHandler := &ProfileHandler{}
	profileHandler.orchInstance = h
	profileStatus := createProfile(profileHandler)
	if profileStatus != nil {
		if intval, ok := profileStatus.(int); ok {
			w.WriteHeader(intval)
		} else {
			w.WriteHeader(500)
		}
		w.Write(h.response.payload[h.compositeAppName+"_profile"])
		return
	}

	w.WriteHeader(201)
	w.Write(h.response.payload[h.compositeAppName])
}

func (h *OrchestrationHandler) createCluster(filename string, fh *multipart.FileHeader, clusterName string,
	jsonData ClusterMetadata) interface{} {
	url := "http://" + h.MiddleendConf.Clm + "/v2/cluster-providers/" + clusterName + "/clusters"

	jsonLoad, _ := json.Marshal(jsonData)

	status, err := h.apiPostMultipart(jsonLoad, fh, url, clusterName, filename)
	if err != nil {
		return err
	}
	if status != 201 {
		return status
	}
	fmt.Printf("cluster creation %s status %s\n", clusterName, status)
	return nil
}

func (h *OrchestrationHandler) CheckConnection(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	parse_err := r.ParseMultipartForm(16777216)
	if parse_err != nil {
		fmt.Printf("multipart error: %s", parse_err.Error())
		w.WriteHeader(500)
		return
	}

	var fh *multipart.FileHeader
	for _, v := range r.MultipartForm.File {
		fh = v[0]
	}
	file, err := fh.Open()
	if err != nil {
		fmt.Printf("Failed to open the file: %s", err.Error())
		w.WriteHeader(500)
		return
	}
	defer file.Close()

	// Read the kconfig
	kubeconfig, _ := ioutil.ReadAll(file)

	jsonData := ClusterMetadata{}
	jsn := ([]byte(r.FormValue("metadata")))
	err = json.Unmarshal(jsn, &jsonData)
	if err != nil {
		fmt.Printf("Failed to parse json")
		w.WriteHeader(500)
		return
	}
	fmt.Printf("metadata %+v\n", jsonData)

	// RESTConfigFromKubeConfig is a convenience method to give back
	// a restconfig from your kubeconfig bytes.
	config, err := clientcmd.RESTConfigFromKubeConfig(kubeconfig)
	if err != nil {
		fmt.Printf("Error while reading the kubeconfig: %s", err.Error())
		w.WriteHeader(500)
		return
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Failed to create clientset: %s", err.Error())
		w.WriteHeader(500)
		return
	}

	_, err = clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to establish the connection: %s", err.Error())
		w.WriteHeader(403)
		w.Write([]byte("Cluster connectivity failed\n"))
		return
	}

	fmt.Printf("Successfully established the connection\n")
	h.client = http.Client{}
	h.response.status = make(map[string]int)
	h.response.payload = make(map[string][]byte)

	status := h.createCluster(fh.Filename, fh, vars["cluster-provider-name"], jsonData)
	if status != nil {
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(200)
	w.Write(h.response.payload[vars["cluster-provider-name"]])
	return
}
