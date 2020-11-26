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
	"log"
)

type DeploymentIGP struct {
	Metadata apiMetaData `json:"metadata"`
	Spec     DigpSpec    `json:"spec"`
}

type DigpSpec struct {
	Profile           string           `json:"profile"`
	Version           string           `json:"version"`
	Lcloud            string           `json:"logical-cloud"`
	OverrideValuesObj []OverrideValues `json:"override-values"`
}

// OverrideValues ...
type OverrideValues struct {
	AppName   string            `json:"app-name"`
	ValuesObj map[string]string `json:"values"`
}

type IgpIntents struct {
	Metadata apiMetaData `json:"metadata"`
	Spec     AppIntents  `json:"spec"`
}

type AppIntents struct {
	Intent map[string]string `json:"intent"`
}

type DigpIntents struct {
	Intent []DigDeployedIntents `json:"intent"`
}
type DigDeployedIntents struct {
	GenericPlacementIntent string `json:"genericPlacementIntent"`
	Ovnaction              string `json:"ovnaction"`
}

// digpHandler implements the orchworkflow interface
type digpHandler struct {
	orchURL      string
	orchInstance *OrchestrationHandler
}

func (h *digpHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	retcode := 200
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		var digpList []DeploymentIGP
		// This for the cases where the dig name is in the URL
		if orch.treeFilter != nil && orch.treeFilter.digName != ""{
			temp:=DeploymentIGP{}
			h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups/" + orch.treeFilter.digName
			retcode, retval, err := orch.apiGet(h.orchURL, orch.compositeAppName+"_digp")
			fmt.Printf("Get Digp in composite app %s status %d\n", compositeAppMetadata.Name, retcode)
			if err != nil {
				fmt.Printf("Failed to read digp")
				return nil, 500
			}
			if retcode != 200 {
				fmt.Printf("Failed to read digp")
			return nil, retcode
			}
			json.Unmarshal(retval, &temp)
			digpList = append(digpList, temp)
		} else {
			h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
				orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
				"/" + compositeAppSpec.Version +
				"/deployment-intent-groups"
			retcode, retval, err := orch.apiGet(h.orchURL, orch.compositeAppName+"_digp")
			fmt.Printf("Get Digp in composite app %s status %d\n", compositeAppMetadata.Name, retcode)
			if err != nil {
				fmt.Printf("Failed to read digp")
				return nil, 500
			}
			if retcode != 200 {
				fmt.Printf("Failed to read digp")
			return nil, retcode
			}
			json.Unmarshal(retval, &digpList)
		}

		compositeAppValue.DigMap = make(map[string]*DigReadData, len(digpList))
		for _, digpValue := range digpList {
			var Dig DigReadData
			Dig.DigpData = digpValue
			compositeAppValue.DigMap[digpValue.Metadata.Name] = &Dig
		}
	}
	return nil, retcode
}

func (h *digpHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version +
			"/deployment-intent-groups/"
		digpList := compositeAppValue.DigMap
		for digName, digValue := range digpList {
			url := h.orchURL + digName + "/intents"
			retcode, retval, err := orch.apiGet(url, compositeAppMetadata.Name+"_digpIntents")
			fmt.Printf("Get Dig int composite app %s Dig %s status %d \n", orch.compositeAppName,
				digName, retcode)
			if err != nil {
				fmt.Printf("Failed to read digp intents")
				return nil, 500
			}
			if retcode != 200 {
				fmt.Printf("Failed to read digp intents")
				return nil, retcode

			}
			err = json.Unmarshal(retval, &digValue.DigIntentsData)
			if err != nil {
				fmt.Printf("Failed to read intents %s\n", err)
			}
		}
	}
	return nil, 200
}

func (h *digpHandler) deleteObject() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		digpList := compositeAppValue.DigMap
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version +
			"/deployment-intent-groups/"

		for digName, _ := range digpList {
			url := h.orchURL + digName + "/intents/PlacementIntent"
			fmt.Printf("dlete intents %s\n", url)
			resp, err := orch.apiDel(url, orch.compositeAppName+"_deldigintents")
			if err != nil {
				return err
			}
			if resp != 204 {
				return resp
			}
			fmt.Printf("Delete dig intents resp %s\n", resp)
		}
	}
	return nil
}

func (h *digpHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		digpList := compositeAppValue.DigMap
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version +
			"/deployment-intent-groups/"

		// loop through all the intents in the dig
		for digName, _ := range digpList {
			url := h.orchURL + digName
			turl := h.orchURL + digName + "/terminate"
			fmt.Printf("delete intents %s\n", url)
			jsonLoad, _ := json.Marshal("{}")
			resp, err := orch.apiPost(jsonLoad, turl, orch.compositeAppName+"_terminatedig")
			//Not checking the status of terminate FIXME
			resp, err = orch.apiDel(url, orch.compositeAppName+"_deldig")
			if err != nil {
				return err
			}
			if resp != 204 {
				return resp
			}
			fmt.Printf("Delete dig resp %s\n", resp)
		}
	}
	return nil
}

func (h *digpHandler) createAnchor() interface{} {
	digData := h.orchInstance.DigData
	orch := h.orchInstance

	digp := DeploymentIGP{
		Metadata: apiMetaData{
			Name:        digData.Name,
			Description: digData.Description,
			UserData1:   "data 1",
			UserData2:   "data2"},
		Spec: DigpSpec{
			Profile:           digData.CompositeProfile,
			Version:           digData.DigVersion,
			Lcloud:            "unused_logical_cloud",
			OverrideValuesObj: make([]OverrideValues, len(digData.Spec.Apps)),
		},
	}
	overrideVals := digp.Spec.OverrideValuesObj
	for i, value := range digData.Spec.Apps {
		overrideVals[i].ValuesObj = make(map[string]string)
		overrideVals[i].AppName = value.Metadata.Name
		overrideVals[i].ValuesObj["Values.global.dcaeCollectorIp"] = "1.8.0"
	}

	jsonLoad, _ := json.Marshal(digp)

	// POST the generic placement intent
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" + digData.Spec.ProjectName +
		"/composite-apps/" + digData.CompositeAppName + "/" + digData.CompositeAppVersion +
		"/deployment-intent-groups"
	resp, err := orch.apiPost(jsonLoad, h.orchURL, digData.Name)
	if err != nil {
		return err
	}
	if resp != 201 {
		return resp
	}
	orch.digpIntents["generic-placement-intent"] = digData.CompositeAppName + "_gpint"
	orch.nwCtlIntents["network-controller-intent"] = digData.CompositeAppName + "_nwctlint"
	fmt.Printf("Deloyment intent group resp %s\n", resp)

	return nil
}

func (h *digpHandler) createObject() interface{} {
	digData := h.orchInstance.DigData
	orch := h.orchInstance
	intentName := "PlacementIntent"
	igp := IgpIntents{
		Metadata: apiMetaData{
			Name:        intentName,
			Description: "NA",
			UserData1:   "data 1",
			UserData2:   "data2"},
	}
	if len(digData.Spec.Apps[0].Clusters[0].SelectedClusters[0].Interfaces) != 0 {
		igp.Spec.Intent = make(map[string]string)
		igp.Spec.Intent["genericPlacementIntent"] = orch.digpIntents["generic-placement-intent"]
		igp.Spec.Intent["ovnaction"] = orch.nwCtlIntents["network-controller-intent"]
	} else {
		igp.Spec.Intent = make(map[string]string)
		igp.Spec.Intent["genericPlacementIntent"] = orch.digpIntents["generic-placement-intent"]
	}

	url := h.orchURL + "/" + digData.Name + "/intents"
	jsonLoad, _ := json.Marshal(igp)
	status, err := orch.apiPost(jsonLoad, url, intentName)
	fmt.Printf("DIG name req %s", string(jsonLoad))
	if err != nil {
		log.Fatalln(err)
	}
	if status != 201 {
		return status
	}
	fmt.Printf("Placement intent %s status %s %s\n", intentName, status, url)

	return nil
}

func createDInents(I orchWorkflow) interface{} {
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

func delDigp(I orchWorkflow) interface{} {
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
