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

package utils

import (
	"net/http"
	"os"

	con "github.com/onap/multicloud-k8s/src/inventory/constants"
	"github.com/onap/multicloud-k8s/src/inventory/model"
)

/* Building relationship json to attach vserver details to vf-module*/
func BuildRelationshipDataForVFModule(vserverName, vserverID, cloudOwner, cloudRegion, tenantId string) con.RelationList {

	rl := con.RelationList{RelatedTo: "vserver", RelatedLink: "/aai/v14/cloud-infrastructure/cloud-regions/cloud-region/" + cloudOwner + "/" + cloudRegion + "/tenants/tenant/" + tenantId + "/vservers/vserver/" + vserverID, RelationshipData: []con.RData{con.RData{RelationshipKey: "cloud-region.cloud-owner", RelationshipValue: cloudOwner},
		con.RData{RelationshipKey: "cloud-region.cloud-region-id", RelationshipValue: cloudRegion},
		con.RData{RelationshipKey: "tenant.tenant-id", RelationshipValue: tenantId},
		con.RData{RelationshipKey: "vserver.vserver-id", RelationshipValue: vserverID}},
		RelatedToProperty: []con.Property{con.Property{PropertyKey: "vserver.vserver-name", PropertyValue: vserverName}}}

	return rl

}

func ParseListInstanceResponse(rlist []model.InstanceMiniResponse) []string {

	var resourceIdList []string

	//assume there is only one resource created
	for _, result := range rlist {

		resourceIdList = append(resourceIdList, result.ID)
	}

	return resourceIdList
}

/* Parse status api response to pull required information like Pod name, Profile name, namespace, ip details, vnf-id and vf-module-id*/
func ParseStatusInstanceResponse(instanceStatusses []model.InstanceStatus) []con.PodInfoToAAI {

	var infoToAAI []con.PodInfoToAAI

	for _, instanceStatus := range instanceStatusses {

		var podInfo con.PodInfoToAAI

		podInfo.VserverName2 = instanceStatus.Request.ProfileName
		podInfo.CloudRegion = instanceStatus.Request.CloudRegion

		for key, value := range instanceStatus.Request.Labels {
			if key == "generic-vnf-id" {
				podInfo.VnfId = value
			}
			if key == "vfmodule-id" {
				podInfo.VfmId = value
			}
		}

		for _, rs := range instanceStatus.ResourcesStatus {
			podInfo.VserverName = rs.Name
		}

		infoToAAI = append(infoToAAI, podInfo)

	}

	return infoToAAI

}

/* this sets http headers to request object*/
func SetRequestHeaders(req *http.Request) {
	authorization := os.Getenv("authorization")

	req.Header.Set("X-FromAppId", con.XFromAppId)
	req.Header.Set("Content-Type", con.ContentType)
	req.Header.Set("Accept", con.Accept)
	req.Header.Set("X-TransactionId", con.XTransactionId)
	req.Header.Set("Authorization", authorization)

}
