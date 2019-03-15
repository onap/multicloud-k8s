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

// CreateVnfRequest contains the VNF creation request parameters
type CreateVnfRequest struct {
	CloudRegionID string                   `json:"cloud_region_id"`
	CsarID        string                   `json:"csar_id"`
	RBName        string                   `json:"rb-name"`
	RBVersion     string                   `json:"rb-version"`
	ProfileName   string                   `json:"profile-name"`
	OOFParams     []map[string]interface{} `json:"oof_parameters"`
	NetworkParams NetworkParameters        `json:"network_parameters"`
	Name          string                   `json:"vnf_instance_name"`
	Description   string                   `json:"vnf_instance_description"`
}

// CreateVnfResponse contains the VNF creation response parameters
type CreateVnfResponse struct {
	VNFID         string              `json:"vnf_id"`
	CloudRegionID string              `json:"cloud_region_id"`
	Namespace     string              `json:"namespace"`
	VNFComponents map[string][]string `json:"vnf_components"`
}

// ListVnfsResponse contains the list of VNFs response parameters
type ListVnfsResponse struct {
	VNFs []string `json:"vnf_id_list"`
}

// NetworkParameters contains the networking info required by the VNF instance
type NetworkParameters struct {
	OAMI OAMIPParams `json:"oam_ip_address"`
	// Add other network parameters if necessary.
}

// OAMIPParams contains the management networking info required by the VNF instance
type OAMIPParams struct {
	ConnectionPoint string `json:"connection_point"`
	IPAddress       string `json:"ip_address"`
	WorkLoadName    string `json:"workload_name"`
}

// UpdateVnfRequest contains the VNF creation parameters
type UpdateVnfRequest struct {
	CloudRegionID string                   `json:"cloud_region_id"`
	CsarID        string                   `json:"csar_id"`
	OOFParams     []map[string]interface{} `json:"oof_parameters"`
	NetworkParams NetworkParameters        `json:"network_parameters"`
	Name          string                   `json:"vnf_instance_name"`
	Description   string                   `json:"vnf_instance_description"`
}

// UpdateVnfResponse contains the VNF update response parameters
type UpdateVnfResponse struct {
	DeploymentID string `json:"vnf_id"`
	Name         string `json:"name"`
}

// GetVnfResponse returns information about a specific VNF instance
type GetVnfResponse struct {
	VNFID         string              `json:"vnf_id"`
	CloudRegionID string              `json:"cloud_region_id"`
	Namespace     string              `json:"namespace"`
	VNFComponents map[string][]string `json:"vnf_components"`
}
