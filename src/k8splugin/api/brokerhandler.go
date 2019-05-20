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
	"io"
	"log"
	"net/http"

	"k8splugin/internal/app"
	"k8splugin/internal/helm"

	"github.com/gorilla/mux"
)

// Used to store the backend implementation objects
// Also simplifies the mocking needed for unit testing
type brokerInstanceHandler struct {
	// Interface that implements the Instance operations
	client app.InstanceManager
}

type brokerRequest struct {
	GenericVnfID                 string                 `json:"generic-vnf-id"`
	VFModuleID                   string                 `json:"vf-module-id"`
	VFModuleModelInvariantID     string                 `json:"vf-module-model-invariant-id"`
	VFModuleModelVersionID       string                 `json:"vf-module-model-version-id"`
	VFModuleModelCustomizationID string                 `json:"vf-module-model-customization-id"`
	OOFDirectives                map[string]interface{} `json:"oof_directives"`
	SDNCDirections               map[string]interface{} `json:"sdnc_directives"`
	UserDirectives               map[string]interface{} `json:"user_directives"`
	TemplateType                 string                 `json:"template_type"`
	TemplateData                 map[string]interface{} `json:"template_data"`
}

type brokerPOSTResponse struct {
	TemplateType     string                    `json:"template_type"`
	WorkloadID       string                    `json:"workload_id"`
	TemplateResponse []helm.KubernetesResource `json:"template_response"`
}

type brokerGETResponse struct {
	TemplateType   string `json:"template_type"`
	WorkloadID     string `json:"workload_id"`
	WorkloadStatus string `json:"workload_status"`
}

// getUserDirectiveValue parses the following kind of json
// "user_attributes": {
// 		"attributes": [
// 		{
// 			"attribute_value": "foo",
// 			"attribute_name": "bar"
// 		},
// 		{
// 			"attribute_value": "value2",
// 			"attribute_name": "name2"
// 		}
// 		]
// }
func (b brokerRequest) getUserDirectiveValue(inp string) string {
	attributes, ok := b.UserDirectives["attributes"].([]interface{})
	if !ok {
		log.Println("Unable to cast attributes to []interface{}")
		return ""
	}

	for _, value := range attributes {

		attribute, ok := value.(map[string]interface{})
		if !ok {
			log.Println("Unable to cast attribute to map[string]interface{}")
			return ""
		}

		attributename, ok := attribute["attribute_name"].(string)
		if !ok {
			log.Println("Unable to cast attribute_name to string")
			return ""
		}
		if attributename == inp {
			attributevalue, ok := attribute["attribute_value"].(string)
			if !ok {
				log.Println("Unable to cast attribute_value to string")
				return ""
			}

			return attributevalue
		}
	}
	return ""
}

func (b brokerInstanceHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cloudRegion := vars["cloud-region"]

	var req brokerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	switch {
	case err == io.EOF:
		http.Error(w, "Body empty", http.StatusBadRequest)
		return
	case err != nil:
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	// Check body for expected parameters
	if req.VFModuleModelCustomizationID == "" {
		http.Error(w, "vf-module-model-customization-id is empty", http.StatusBadRequest)
		return
	}

	rbName := req.getUserDirectiveValue("definition-name")
	if rbName == "" {
		http.Error(w, "definition-name is missing from user-directives", http.StatusBadRequest)
		return
	}

	rbVersion := req.getUserDirectiveValue("definition-version")
	if rbVersion == "" {
		http.Error(w, "definition-version is missing from user-directives", http.StatusBadRequest)
		return
	}

	profileName := req.getUserDirectiveValue("profile-name")
	if profileName == "" {
		http.Error(w, "profile-name is missing from user-directives", http.StatusBadRequest)
		return
	}

	// Setup the resource parameters for making the request
	var instReq app.InstanceRequest
	instReq.RBName = rbName
	instReq.RBVersion = rbVersion
	instReq.ProfileName = profileName
	instReq.CloudRegion = cloudRegion

	resp, err := b.client.Create(instReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brokerResp := brokerPOSTResponse{
		TemplateType:     "heat",
		WorkloadID:       resp.ID,
		TemplateResponse: resp.Resources,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(brokerResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// getHandler retrieves information about an instance via the ID
func (b brokerInstanceHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instID"]

	resp, err := b.client.Get(instanceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brokerResp := brokerGETResponse{
		TemplateType:   "heat",
		WorkloadID:     resp.ID,
		WorkloadStatus: "CREATED",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(brokerResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// deleteHandler method terminates an instance via the ID
func (b brokerInstanceHandler) deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	instanceID := vars["instID"]

	err := b.client.Delete(instanceID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
}
