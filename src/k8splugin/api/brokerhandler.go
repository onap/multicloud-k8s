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
	"net/http"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"

	"github.com/gorilla/mux"
)

// Used to store the backend implementation objects
// Also simplifies the mocking needed for unit testing
type brokerInstanceHandler struct {
	// Interface that implements the Instance operations
	client app.InstanceManager
}

type brokerRequest struct {
	GenericVnfID                 string       `json:"generic-vnf-id"`
	VFModuleID                   string       `json:"vf-module-id"`
	VFModuleModelInvariantID     string       `json:"vf-module-model-invariant-id"`
	VFModuleModelVersionID       string       `json:"vf-module-model-version-id"`
	VFModuleModelCustomizationID string       `json:"vf-module-model-customization-id"`
	OOFDirectives                directive    `json:"oof_directives"`
	SDNCDirectives               directive    `json:"sdnc_directives"`
	UserDirectives               directive    `json:"user_directives"`
	TemplateType                 string       `json:"template_type"`
	TemplateData                 templateData `json:"template_data"`
}

type directive struct {
	Attributes []attribute `json:"attributes"`
}

type attribute struct {
	Key   string `json:"attribute_name"`
	Value string `json:"attribute_value"`
}

type templateData struct {
	StackName       string `json:"stack_name"` //Only this property is relevant (exported)
	disableRollback string `json:"disable_rollback"`
	environment     string `json:"environment"`
	parameters      string `json:"parameters"`
	template        string `json:"template"`
	timeoutMins     string `json:"timeout_mins"`
}

type brokerPOSTResponse struct {
	TemplateType         string                    `json:"template_type"`
	WorkloadID           string                    `json:"workload_id"`
	TemplateResponse     []helm.KubernetesResource `json:"template_response"`
	WorkloadStatus       string                    `json:"workload_status"`
	WorkloadStatusReason map[string]interface{}    `json:"workload_status_reason"`
}

type brokerGETResponse struct {
	TemplateType         string                 `json:"template_type"`
	WorkloadID           string                 `json:"workload_id"`
	WorkloadStatus       string                 `json:"workload_status"`
	WorkloadStatusReason map[string]interface{} `json:"workload_status_reason"`
}

type brokerDELETEResponse struct {
	TemplateType         string                 `json:"template_type"`
	WorkloadID           string                 `json:"workload_id"`
	WorkloadStatus       string                 `json:"workload_status"`
	WorkloadStatusReason map[string]interface{} `json:"workload_status_reason"`
}

// Convert directives stored in broker request to map[string]string format with
// merge including precedence provided
func (b brokerRequest) convertDirectives() map[string]string {
	extractedAttributes := make(map[string]string)
	for _, section := range [3]directive{b.SDNCDirectives, b.OOFDirectives, b.UserDirectives} {
		for _, attribute := range section.Attributes {
			extractedAttributes[attribute.Key] = attribute.Value
		}
	}
	return extractedAttributes
}

func (b brokerInstanceHandler) createHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cloudRegion := vars["cloud-region"]

	var req brokerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	log.Info("Broker API Payload", log.Fields{
		"Body":    r.Body,
		"Payload": req,
	})
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

	if req.VFModuleModelInvariantID == "" {
		http.Error(w, "vf-module-model-invariant-id is empty", http.StatusBadRequest)
		return
	}

	if req.VFModuleModelVersionID == "" {
		http.Error(w, "vf-module-model-version-id is empty", http.StatusBadRequest)
		return
	}

	if req.GenericVnfID == "" {
		http.Error(w, "generic-vnf-id is empty", http.StatusBadRequest)
		return
	}
	if req.VFModuleID == "" {
		http.Error(w, "vf-module-id is empty", http.StatusBadRequest)
		return
	}

	if req.TemplateData.StackName == "" {
		http.Error(w, "stack_name is missing from template_data", http.StatusBadRequest)
		return
	}

	directives := req.convertDirectives()
	profileName, ok := directives["k8s-rb-profile-name"]
	if !ok {
		http.Error(w, "k8s-rb-profile-name is missing from directives", http.StatusBadRequest)
		return
	}

	releaseName, ok := directives["k8s-rb-instance-release-name"]
	if !ok {
		//Release name is not mandatory argument, but we're not using profile's default
		//as it could conflict if someone wanted to instantiate single profile multiple times
		releaseName = req.VFModuleID
	}

	// Setup the resource parameters for making the request
	var instReq app.InstanceRequest
	instReq.RBName = req.VFModuleModelInvariantID
	instReq.RBVersion = req.VFModuleModelCustomizationID
	instReq.ProfileName = profileName
	instReq.CloudRegion = cloudRegion
	instReq.ReleaseName = releaseName
	instReq.Labels = map[string]string{
		"stack-name": req.TemplateData.StackName,
	}
	instReq.OverrideValues = directives

	log.Info("Instance API Payload", log.Fields{
		"payload": instReq,
	})
	resp, err := b.client.Create(instReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	brokerResp := brokerPOSTResponse{
		TemplateType:     "heat",
		WorkloadID:       resp.ID,
		TemplateResponse: resp.Resources,
		WorkloadStatus:   "CREATE_COMPLETE",
	}
	log.Info("Broker API Response", log.Fields{
		"response": brokerResp,
	})

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
		WorkloadStatus: "CREATE_COMPLETE",
	}

	log.Info("Broker API Response", log.Fields{
		"response": brokerResp,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(brokerResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// findHandler retrieves information about an instance via the ID
func (b brokerInstanceHandler) findHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	//name is an alias for stack-name from the so adapter
	name := vars["name"]
	responses, _ := b.client.Find("", "", "", map[string]string{"stack-name": name})

	brokerResp := brokerGETResponse{
		TemplateType:   "heat",
		WorkloadID:     "",
		WorkloadStatus: "GET_COMPLETE",
		WorkloadStatusReason: map[string]interface{}{
			//treating stacks as an array of map[string]interface{} types
			"stacks": []map[string]interface{}{},
		},
	}

	if len(responses) != 0 {
		//Return the first object that matches.
		resp := responses[0]
		brokerResp.WorkloadID = resp.ID
		brokerResp.WorkloadStatus = "CREATE_COMPLETE"
		brokerResp.WorkloadStatusReason["stacks"] = []map[string]interface{}{
			{
				"stack_status": "CREATE_COMPLETE",
				"id":           resp.ID,
			},
		}
	}

	log.Info("Broker API Response", log.Fields{
		"response": brokerResp,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(brokerResp)
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

	brokerResp := brokerDELETEResponse{
		TemplateType:   "heat",
		WorkloadID:     instanceID,
		WorkloadStatus: "DELETE_COMPLETE",
	}
	log.Info("Broker API Response", log.Fields{
		"response": brokerResp,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(brokerResp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
