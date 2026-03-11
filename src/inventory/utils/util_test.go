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
	"testing"

	"github.com/onap/multicloud-k8s/src/inventory/model"
)

func TestBuildRelationshipDataForVFModule(t *testing.T) {

	relList := BuildRelationshipDataForVFModule("vs_name", "vs1234", "CO", "CR", "tenant1234")

	if relList.RelatedTo != "vserver" {
		t.Error("Failed")
	}

	if (relList.RelatedLink) != "/aai/v14/cloud-infrastructure/cloud-regions/cloud-region/CO/CR/tenants/tenant/tenant1234/vservers/vserver/vs1234" {
		t.Error("Failed")
	}

	rdadaList := relList.RelationshipData

	for _, rdata := range rdadaList {

		if rdata.RelationshipKey == "cloud-region.cloud-region-id" {

			if rdata.RelationshipValue != "CR" {

				t.Error("Failed")

			}
		}

		if rdata.RelationshipKey == "tenant.tenant-id" {

			if rdata.RelationshipValue != "tenant1234" {

				t.Error("Failed")

			}
		}

		if rdata.RelationshipKey == "vserver.vserver-id" {

			if rdata.RelationshipValue != "vs1234" {

				t.Error("Failed")

			}
		}

		if rdata.RelationshipKey == "cloud-region.cloud-owner" {

			if rdata.RelationshipValue != "CO" {

				t.Error("Failed")

			}
		}

	}

	propertyList := relList.RelatedToProperty

	for _, property := range propertyList {

		if property.PropertyKey == "vserver.vserver-name" {

			if property.PropertyValue != "vs_name" {

				t.Error("Failed")

			}
		}

	}

}

func TestParseStatusInstanceResponse(t *testing.T) {

	var resourceIdList []model.InstanceStatus

	instanceRequest := model.InstanceRequest{
		RBName:      "rb_name",
		RBVersion:   "rb_version",
		ProfileName: "profile123456",
		ReleaseName: "release1",
		CloudRegion: "c_region",
		Labels:      map[string]string{"generic-vnf-id": "123456789", "vf-module-id": "987654321"},
	}
	instanceStatus := model.InstanceStatus{
		Request:       instanceRequest,
		Ready:         true,
		ResourceCount: 12,
		ResourcesStatus: []model.ResourceStatus{
			{Name: "pod123", GVK: model.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}},
			{Name: "pod456", GVK: model.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"}},
		},
	}

	resourceIdList = append(resourceIdList, instanceStatus)

	podInfoToAAI := ParseStatusInstanceResponse(resourceIdList)

	for _, podInfo := range podInfoToAAI {

		if podInfo.VserverName2 != "profile123456" {
			t.Error("Expected VserverName2 to be profile123456")
		}

		if podInfo.CloudRegion != "c_region" {
			t.Error("Expected CloudRegion to be c_region")
		}

		if podInfo.VnfId != "123456789" {
			t.Error("Expected VnfId to be 123456789")
		}

		if podInfo.VfmId != "987654321" {
			t.Error("Expected VfmId to be 987654321")
		}

		// Last resource name wins in the loop
		if podInfo.VserverName != "pod456" {
			t.Error("Expected VserverName to be pod456")
		}

	}

}
