/*
Copyright 2020  Tech Mahindra.
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

package constants

import (

//corev1 "k8s.io/api/core/v1"

)

const (
	XFromAppId     = "SO"
	ContentType    = "application/json"
	Accept         = "application/json"
	XTransactionId = "get_aai_subscr"
)

const (
	AAI_EP   = "/aai/v14"
	AAI_CREP = "/cloud-infrastructure/cloud-regions/"
	AAI_NEP  = "/network/generic-vnfs/generic-vnf/"
)

const (
	MK8S_EP  = "/api/multicloud-k8s/v1/v1/instance/"
	MK8S_CEP = "/connectivity-info"
)

type PodInfoToAAI struct {
	VserverName                string
	VserverName2               string
	ProvStatus                 string
	I3InterfaceIPv4Address     string
	I3InterfaceIPvPrefixLength int32
	VnfId                      string
	VfmId                      string
	CloudRegion                string
}

type RData struct {
	RelationshipKey   string `json:"relationship-key"`
	RelationshipValue string `json:"relationship-value"`
}

type RelationList struct {
	RelatedTo         string     `json:"related-to"`
	RelatedLink       string     `json:"related-link"`
	RelationshipData  []RData    `json:"relationship-data"`
	RelatedToProperty []Property `json:"related-to-property"`
}

type TenantInfo struct {
	TenantId   string `json:"tenant-id"`
	TenantName string `json:"tenant-name"`
}

type Tenant struct {
	Tenants map[string][]TenantInfo `json:"tenants"`
}

type Property struct {
	PropertyKey   string `json:"property-key"`
	PropertyValue string `json:"property-value"`
}

type VFModule struct {
	VFModuleId           string                    `json:"vf-module-id"`
	VFModuleName         string                    `json:"vf-module-name"`
	HeatStackId          string                    `json:"heat-stack-id"`
	OrchestrationStatus  string                    `json:"orchestration-status"`
	ResourceVersion      string                    `json:"resource-version"`
	AutomatedAssignment  string                    `json:"automated-assignment"`
	IsBaseVfModule       string                    `json:"is-base-vf-module"`
	RelationshipList     map[string][]RelationList `json:"relationship-list"`
	ModelInvariantId     string                    `json:"model-invariant-id"`
	ModelVersionId       string                    `json:"model-version-id"`
	ModelCustomizationId string                    `json:"model-customization-id"`
	ModuleIndex          string                    `json:"module-index"`
}

type VFModules struct {
	VFModules []VFModule `json:"vf-module"`
}
