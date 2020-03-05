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
	"testing"
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

	var resourceIdList []k8splugin.InstanceStatus

	instanceRequest := k8splugin.InstanceRequest{"rb_name", "rb_version", "profile123456", "c_region", map[string]string{"generic-vnf-id": "123456789", "vf-module-id": "987654321"}}
	instanceStatus := k8splugin.InstanceStatus{instanceRequest, true, 12, []con.PodStatus{con.PodStatus{"pod123", "onap", true, []string{"10.211.1.100", "10.211.1.101"}}, con.PodStatus{"pod456", "default", true, []string{"10.211.1.200", "10.211.1.201"}}}}

	resourceIdList = append(resourceIdList, instanceStatus)

	podInfoToAAI := ParseStatusInstanceResponse(resourceIdList)

	for _, podInfo := range podInfoToAAI {

		if podInfo.VserverName == "pod123" {

			t.Error("Failed")

		}

		if podInfo.VserverName2 == "default" {

			t.Error("Failed")

		}

		if podInfo.ProvStatus == "profile123456" {

			t.Error("Failed")

		}

	}

}
