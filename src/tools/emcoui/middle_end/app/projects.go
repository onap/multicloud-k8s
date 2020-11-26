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
type ProjectMetadata struct {
	Metadata apiMetaData `json:"metadata"`
}

// CompAppHandler , This implements the orchworkflow interface
type projectHandler struct {
	orchURL      string
	orchInstance *OrchestrationHandler
}

func (h *projectHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	var cappList []CompositeApp
	if orch.treeFilter != nil {
		temp:=CompositeApp{}
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + orch.treeFilter.compositeAppName+"/"+
			orch.treeFilter.compositeAppVersion
		respcode, respdata, err := orch.apiGet(h.orchURL, orch.projectName+"_getcapps")
		fmt.Printf("Get capp status %s\n", respcode)
		if err != nil {
			return nil, 500
		}
		if respcode != 200 {
			return nil, respcode
		}
		fmt.Printf("Get capp status %s\n", respcode)
		json.Unmarshal(respdata, &temp)
		cappList = append(cappList, temp)
	} else {
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps"
		respcode, respdata, err := orch.apiGet(h.orchURL, orch.projectName+"_getcapps")
		fmt.Printf("Get capp status %s\n", respcode)
		if err != nil {
			return nil, 500
		}
		if respcode != 200 {
			return nil, respcode
		}
		fmt.Printf("Get capp status %s\n", respcode)
		json.Unmarshal(respdata, &cappList)
	}

	dataRead.compositeAppMap = make(map[string]*CompositeAppTree, len(cappList))
	for k, value := range cappList {
		fmt.Printf("%+v", cappList[k])
		var cappsDataInstance CompositeAppTree
		cappName := value.Metadata.Name
		cappsDataInstance.Metadata = value
		dataRead.compositeAppMap[cappName] = &cappsDataInstance
	}
	return nil, 200 
}

func (h *projectHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		orch.projectName

	respcode, respdata, err := orch.apiGet(h.orchURL, orch.projectName+"_getProject")
	if err != nil {
		return nil, 500
	}
	if respcode != 200 {
		return nil, respcode
	}
	fmt.Printf("Get project %s\n", respcode)
	json.Unmarshal(respdata, &dataRead.Metadata)
	return nil, respcode
}

func (h *projectHandler) deleteObject() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	cappList := dataRead.compositeAppMap
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		orch.projectName + "/composite-apps"
	for compositeAppName, compositeAppValue := range cappList {
		url := h.orchURL + "/" + compositeAppName + "/" + compositeAppValue.Metadata.Spec.Version
		fmt.Printf("Delete composite app %s\n", url)
		resp, err := orch.apiDel(url, compositeAppName+"_delcapp")
		if err != nil {
			return err
		}
		if resp != 204 {
			return resp
		}
		fmt.Printf("Delete composite app status %s\n", resp)
	}
	return nil
}

func (h *projectHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + orch.projectName
	fmt.Printf("Delete Project %s \n", h.orchURL)
	resp, err := orch.apiDel(h.orchURL, orch.projectName+"_delProject")
	if err != nil {
		return err
	}
	if resp != 204 {
		return resp
	}
	fmt.Printf("Delete Project status %s\n", resp)
	return nil
}

func (h *projectHandler) createAnchor() interface{} {
	orch := h.orchInstance

	projectCreate := ProjectMetadata{
		Metadata: apiMetaData{
			Name:        orch.projectName,
			Description: orch.projectDesc,
			UserData1:   "data 1",
			UserData2:   "data 2"},
	}

	jsonLoad, _ := json.Marshal(projectCreate)
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + orch.projectName
	resp, err := orch.apiPost(jsonLoad, h.orchURL, orch.projectName)
	if err != nil {
		return err
	}
	if resp != 201 {
		return resp
	}
	orch.version = "v1"
	fmt.Printf("projectHandler resp %s\n", resp)

	return nil
}

func (h *projectHandler) createObject() interface{} {
	return nil
}

func createProject(I orchWorkflow) interface{} {
	// 1. Create the Anchor point
	err := I.createAnchor()
	if err != nil {
		return err
	}
	return nil
}

func delProject(I orchWorkflow) interface{} {
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
