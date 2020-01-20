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
	Authorization  = "Basic QUFJOkFBSQ=="
)

const (
	AAI_EP   = "/aai/v14"
	AAI_CREP = "/cloud-infrastructure/cloud-regions/"
	AAI_NEP  = "/network/generic-vnfs/generic-vnf/"
	AAI_URI  = "https://10.211.1.93"
	AAI_Port = "30233"
)

const (
	MK8S_EP   = "/api/multicloud-k8s/v1/v1/instance/"
	MK8S_CEP  = "/connectivity-info"
	MK8S_URI  = "https://multicloud-host"
	MK8S_Port = "30280"
)

type Connection struct {
	CloudRegion           string                 `json:"cloud-region"`
	CloudOwner            string                 `json:"cloud-owner"`
	Kubeconfig            string                 `json:"kubeconfig"`
	OtherConnectivityList ConnectivityRecordList `json:"other-connectivity-list"`
}

type ConnectivityRecordList struct {
	ConnectivityRecords []map[string]string `json:"connectivity-records"`
}

type InstanceRequest struct {
	RBName      string            `json:"rb-name"`
	RBVersion   string            `json:"rb-version"`
	ProfileName string            `json:"profile-name"`
	CloudRegion string            `json:"cloud-region"`
	Labels      map[string]string `json:"labels"`
}

type InstanceMiniResponse struct {
	ID        string          `json:"id"`
	Request   InstanceRequest `json:"request"`
	Namespace string          `json:"namespace"`
}

type PodStatus struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Ready     bool   `json:"ready"`
	//Status      corev1.PodStatus `json:"status,omitempty"`
	IPAddresses []string `json:"ipaddresses"`
}

type InstanceStatus struct {
	Request       InstanceRequest `json:"request"`
	Ready         bool            `json:"ready"`
	ResourceCount int32           `json:"resourceCount"`
	PodStatuses   []PodStatus     `json:"podStatuses"`
	//ServiceStatuses []corev1.Service `json:"serviceStatuses"`
}

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

type CRegion struct {
	CloudOwner         string                    `json:"cloud-owner"`
	CloudRegionId      string                    `json:"cloud-region-id"`
	CloudType          string                    `json:"cloud-type"`
	OwnerDefinedType   string                    `json:"owner-defined-type"`
	CloudRegionVersion string                    `json:"cloud-region-version"`
	CloudZone          string                    `json:"cloud-zone"`
	ResourceVersion    string                    `json:"resource-version"`
	ComplexName        string                    `json:"complex-name"`
	SriovAutomation    string                    `json:"sriov-automation"`
	CloudExtraInfo     string                    `json:"cloud-extra-info"`
	RelationshipList   map[string][]RelationList `json:"relationship-list"`
}

type CloudRegion struct {
	Regions []CRegion `json:"cloud-region"`
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
