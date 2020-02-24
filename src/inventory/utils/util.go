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
	con "github.com/onap/multicloud-k8s/src/inventory/constants"
	k8splugin "github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"net/http"
	"os"
	"reflect"
)

/* Building relationship json to attach vserver details to vf-module*/
func BuildRelationshipDataForVFModule(vserverName, vserverID, cloudOwner, cloudRegion, tenantId string) con.RelationList {

	rl := con.RelationList{"vserver", "/aai/v14/cloud-infrastructure/cloud-regions/cloud-region/" + cloudOwner + "/" + cloudRegion + "/tenants/tenant/" + tenantId + "/vservers/vserver/" + vserverID, []con.RData{con.RData{"cloud-region.cloud-owner", cloudOwner},
		con.RData{"cloud-region.cloud-region-id", cloudRegion},
		con.RData{"tenant.tenant-id", tenantId},
		con.RData{"vserver.vserver-id", vserverID}},
		[]con.Property{con.Property{"vserver.vserver-name", vserverName}}}

	return rl

}

func ParseListInstanceResponse(rlist []k8splugin.InstanceMiniResponse) []string {

	var resourceIdList []string

	//assume there is only one resource created
	for _, result := range rlist {

		resourceIdList = append(resourceIdList, result.ID)
	}

	return resourceIdList
}

/* Parse status api response to pull required information like Pod name, Profile name, namespace, ip details, vnf-id and vf-module-id*/
func ParseStatusInstanceResponse(instanceStatusses []k8splugin.InstanceStatus) []con.PodInfoToAAI {

	var infoToAAI []con.PodInfoToAAI

	for _, instanceStatus := range instanceStatusses {

		var podInfo con.PodInfoToAAI

		sa := reflect.ValueOf(&instanceStatus).Elem()
		typeOf := sa.Type()
		for i := 0; i < sa.NumField(); i++ {
			f := sa.Field(i)
			if typeOf.Field(i).Name == "Request" {
				request := f.Interface()
				if ireq, ok := request.(k8splugin.InstanceRequest); ok {
					podInfo.VserverName2 = ireq.ProfileName
					podInfo.CloudRegion = ireq.CloudRegion

					for key, value := range ireq.Labels {
						if key == "generic-vnf-id" {

							podInfo.VnfId = value

						}
						if key == "vfmodule-id" {

							podInfo.VfmId = value

						}
					}

				} else {
					//fmt.Printf("it's not a InstanceRequest \n")
				}
			}

			if typeOf.Field(i).Name == "PodStatuses" {
				ready := f.Interface()
				if pss, ok := ready.([]con.PodStatus); ok {
					for _, ps := range pss {
						podInfo.VserverName = ps.Name
						podInfo.ProvStatus = ps.Namespace
					}

				} else {
					//fmt.Printf("it's not a InstanceRequest \n")
				}
			}
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
