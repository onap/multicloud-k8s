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

// ProfileData captures per app profile
type ProfileData struct {
	Name        string            `json:"profileName"`
	AppProfiles map[string]string `json:"appProfile"`
}

// ProfileMeta is metadta for the profile APIs
type ProfileMeta struct {
	Metadata apiMetaData `json:"metadata"`
	Spec     ProfileSpec `json:"spec"`
}

// ProfileSpec is the spec for the profile APIs
type ProfileSpec struct {
	AppName string `json:"app-name"`
}

// ProfileHandler This implements the orchworkflow interface
type ProfileHandler struct {
	orchURL      string
	orchInstance *OrchestrationHandler
	response     struct {
		payload map[string][]byte
		status  map[string]string
	}
}

func (h *ProfileHandler) getObject() (interface{}, interface{}) {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	retcode := 200
	for _, compositeAppValue := range dataRead.compositeAppMap {
		var profileList []ProfileMeta
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version + "/composite-profiles"
		for profileName, profileValue := range compositeAppValue.ProfileDataArray {
			url := h.orchURL + "/" + profileName + "/profiles"
			retcode, respval, err := orch.apiGet(url, compositeAppMetadata.Name+"_getprofiles")
			fmt.Printf("Get app profiles status %d\n", retcode)
			if err != nil {
				fmt.Printf("Failed to read profile %s\n", profileName)
				return nil, 500
			}
			if retcode != 200 {
				fmt.Printf("Failed to read profile %s\n", profileName)
				return nil, retcode
			}
			json.Unmarshal(respval, &profileList)
			profileValue.AppProfiles = make([]ProfileMeta, len(profileList))
			for appProfileIndex, appProfile := range profileList {
				profileValue.AppProfiles[appProfileIndex] = appProfile
			}
		}
	}
	return nil, retcode
}

func (h *ProfileHandler) getAnchor() (interface{}, interface{}) {
	orch := h.orchInstance
	respcode := 200
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		var profilemetaList []ProfileMeta
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version + "/composite-profiles"

		respcode, respdata, err := orch.apiGet(h.orchURL, compositeAppMetadata.Name+"_getcprofile")
		if err != nil {
			fmt.Printf("Failed to get composite profiles\n")
			return nil, 500
		}
		if respcode != 200 {
			fmt.Printf("composite profile GET status %d\n", respcode)
			return nil, respcode
		}
		json.Unmarshal(respdata, &profilemetaList)
		compositeAppValue.ProfileDataArray = make(map[string]*ProfilesData, len(profilemetaList))
		for _, value := range profilemetaList {
			ProfilesDataInstance := ProfilesData{}
			ProfilesDataInstance.Profile = value
			compositeAppValue.ProfileDataArray[value.Metadata.Name] = &ProfilesDataInstance
		}
	}
	return nil, respcode
}

func (h *ProfileHandler) deleteObject() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version + "/composite-profiles/"
		for profileName, profileValue := range compositeAppValue.ProfileDataArray {
			for _, appProfileValue := range profileValue.AppProfiles {
				url := h.orchURL + profileName + "/profiles/" + appProfileValue.Metadata.Name

				fmt.Printf("Delete app profiles %s\n", url)
				resp, err := orch.apiDel(url, compositeAppMetadata.Name+"_delappProfiles")
				if err != nil {
					return err
				}
				if resp != 204 {
					return resp
				}
				fmt.Printf("Delete profiles status %s\n", resp)
			}
		}
	}
	return nil
}

func (h *ProfileHandler) deleteAnchor() interface{} {
	orch := h.orchInstance
	dataRead := h.orchInstance.dataRead
	for _, compositeAppValue := range dataRead.compositeAppMap {
		compositeAppMetadata := compositeAppValue.Metadata.Metadata
		compositeAppSpec := compositeAppValue.Metadata.Spec
		h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
			orch.projectName + "/composite-apps/" + compositeAppMetadata.Name +
			"/" + compositeAppSpec.Version + "/composite-profiles/"

		for profileName, _ := range compositeAppValue.ProfileDataArray {
			url := h.orchURL + profileName
			fmt.Printf("Delete profile %s\n", url)
			resp, err := orch.apiDel(url, compositeAppMetadata.Name+"_delProfile")
			if err != nil {
				return err
			}
			if resp != 204 {
				return resp
			}
			fmt.Printf("Delete profile status %s\n", resp)
		}
	}
	return nil
}

func (h *ProfileHandler) createAnchor() interface{} {
	orch := h.orchInstance

	profileCreate := ProfileMeta{
		Metadata: apiMetaData{
			Name:        orch.compositeAppName + "_profile",
			Description: "Profile created from middleend",
			UserData1:   "data 1",
			UserData2:   "data2"},
	}
	jsonLoad, _ := json.Marshal(profileCreate)
	h.orchURL = "http://" + orch.MiddleendConf.OrchService + "/v2/projects/" +
		orch.projectName + "/composite-apps"
	url := h.orchURL + "/" + orch.compositeAppName + "/" + "v1" + "/composite-profiles"
	resp, err := orch.apiPost(jsonLoad, url, orch.compositeAppName+"_profile")
	if err != nil {
		return err
	}
	if resp != 201 {
		return resp
	}
	fmt.Printf("ProfileHandler resp %s\n", resp)

	return nil
}

func (h *ProfileHandler) createObject() interface{} {
	orch := h.orchInstance

	for i := range orch.meta {
		fileName := orch.meta[i].ProfileMetadata.FileName
		appName := orch.meta[i].Metadata.Name
		profileName := orch.meta[i].Metadata.Name + "_profile"

		// Upload the application helm chart
		fh := orch.file[fileName]
		profileAdd := ProfileMeta{
			Metadata: apiMetaData{
				Name:        profileName,
				Description: "NA",
				UserData1:   "data 1",
				UserData2:   "data2"},
			Spec: ProfileSpec{
				AppName: appName},
		}
		compositeProfilename := orch.compositeAppName + "_profile"

		url := h.orchURL + "/" + orch.compositeAppName + "/" + "v1" + "/" +
			"composite-profiles" + "/" + compositeProfilename + "/profiles"
		jsonLoad, _ := json.Marshal(profileAdd)
		status, err := orch.apiPostMultipart(jsonLoad, fh, url, profileName, fileName)
		if err != nil {
			log.Fatalln(err)
		}
		if status != 201 {
			return status
		}
		fmt.Printf("CompositeProfile Profile  %s status %s %s\n", profileName, status, url)
	}

	return nil
}

func createProfile(I orchWorkflow) interface{} {
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

func delProfileData(I orchWorkflow) interface{} {
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
