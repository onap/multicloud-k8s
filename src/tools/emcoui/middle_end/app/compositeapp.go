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
	"encoding/json"
	"fmt"
)

// CompositeApp application structure
type CompositeApp struct {
	Metadata apiMetaData      `json:"metadata"`
	Spec     compositeAppSpec `json:"spec"`
}

type compositeAppSpec struct {
	Version string `json:"version"`
}

// compAppHandler , This implements the orchworkflow interface
type compAppHandler struct {
	orchURL      string
	orchInstance *OrchestrationHandler
}

// CompositeAppKey is the mongo key to fetch apps in a composite app
type CompositeAppKey struct {
	Cname    string      `json:"compositeapp"`
	Project  string      `json:"project"`
	Cversion string      `json:"compositeappversion"`
	App      interface{} `json:"app"`
}

func (h *compAppHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	respcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version + "/apps"

		respcode, respdata, err := orch.apiGet(h.orchURL, orch.compositeAppName+"_getapps")
		if err != nil {
			return nil, 500
		}
		if respcode != 200 {
			return nil, respcode
		}
		fmt.Printf("Get app status %s\n", respcode)
		compositeAppValue.AppsDataArray = make(map[string]*AppsData, len(respdata))
		var appList []CompositeApp
		json.Unmarshal(respdata, &appList)
		for _, value := range appList {
			var appsDataInstance AppsData
			appName := value.Metadata.Name
			appsDataInstance.App = value
			compositeAppValue.AppsDataArray[appName] = &appsDataInstance
		}
	}
	return nil, respcode
}

func (h *compAppHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	respcode := 200
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version

		respcode, _, err := orch.apiGet(h.orchURL, orch.compositeAppName+"_getcompositeapp")
		if err != nil {
			return nil, 500
		}
		if respcode != 200 {
			return nil, respcode
		}
		fmt.Printf("Get composite App %s\n", respcode)
		//json.Unmarshal(respdata, &dataRead.CompositeApp)
	}
	return nil, respcode
}

func (h *compAppHandler) deleteObject() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version
		appList := compositeAppValue.AppsDataArray
		for _, value := range appList {
			url := h.orchURL + "/apps/" + value.App.Metadata.Name
			fmt.Printf("Delete app %s\n", url)
			resp, err := orch.apiDel(url, compositeAppMetadata.Name+"_delapp")
			if err != nil {
				return err
			}
			if resp != 204 {
				return resp
			}
			fmt.Printf("Delete app status %s\n", resp)
		}
	}
	return nil
}

func (h *compAppHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version
		fmt.Printf("Delete composite app %s\n", h.orchURL)
		resp, err := orch.apiDel(h.orchURL, compositeAppMetadata.Name+"_delcompapp")
		if err != nil {
			return err
		}
		if resp != 204 {
			return resp
		}
		fmt.Printf("Delete compapp status %s\n", resp)
	}
	return nil
}

// CreateAnchor creates the anchor point for composite applications,
// profiles, intents etc. For example Anchor for the composite application
// will create the composite application resource in the the DB, and all apps
// will get created and uploaded under this anchor point.
func (h *compAppHandler) createAnchor() interface{} {
	orch := h.orchInstance

	compAppCreate := CompositeApp{
		Metadata: apiMetaData{
			Name:        orch.compositeAppName,
			Description: orch.compositeAppDesc,
			UserData1:   "data 1",
			UserData2:   "data 2"},
		Spec: compositeAppSpec{
			Version: "v1"},
	}

	jsonLoad, _ := json.Marshal(compAppCreate)
	tem := CompositeApp{}
	json.Unmarshal(jsonLoad, &tem)
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		orch.projectName + "/composite-apps"
	resp, err := orch.apiPost(jsonLoad, h.orchURL, orch.compositeAppName)
	if err != nil {
		return err
	}
	if resp != 201 {
		return resp
	}
	orch.version = "v1"
	fmt.Printf("compAppHandler resp %s\n", resp)

	return nil
}

func (h *compAppHandler) createObject() interface{} {
	orch := h.orchInstance
	for i := range orch.meta {
		fileName := orch.meta[i].Metadata.FileName
		appName := orch.meta[i].Metadata.Name
		appDesc := orch.meta[i].Metadata.Description

		// Upload the application helm chart
		fh := orch.file[fileName]
		compAppAdd := CompositeApp{
			Metadata: apiMetaData{
				Name:        appName,
				Description: appDesc,
				UserData1:   "data 1",
				UserData2:   "data2"},
		}
		url := h.orchURL + "/" + orch.compositeAppName + "/" + orch.version + "/apps"

		jsonLoad, _ := json.Marshal(compAppAdd)

		status, err := orch.apiPostMultipart(jsonLoad, fh, url, appName, fileName)
		if err != nil {
			return err
		}
		if status != 201 {
			return status
		}
		fmt.Printf("Composite app %s createObject status %s\n", appName, status)
	}

	return nil
}

func createCompositeapp(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.createAnchor()
	if err != nil {
		return err
	}
	// 2. Create the Objects
	err = I.createObject()
	if err != nil {
		return err
	}
	return nil
}

func delCompositeapp(I orchWorkflow) interface{} {
	// 1. Delete the object
	err := I.deleteObject()
	if err != nil {
		return err
	}
	// 2. Delete the Anchor
	err = I.deleteAnchor()
	if err != nil {
		return err
	}
	return nil
}
