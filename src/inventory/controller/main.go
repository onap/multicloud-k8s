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

package main

import (
	executor "github.com/onap/multicloud-k8s/src/inventory/api"
	con "github.com/onap/multicloud-k8s/src/inventory/constants"
        k8splugin "github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	utils "github.com/onap/multicloud-k8s/src/inventory/utils"
	"os"
	"os/signal"
	"time"
)

/* Root function which periodically polls status api for all the instances in the k8splugin and update the status information accordingly to AAI  */
func QueryAAI() {

	for {
		instanceList := executor.ListInstances()
		statusList := CheckInstanceStatus(instanceList)
		podList := utils.ParseStatusInstanceResponse(statusList)
		PushPodInfoToAAI(podList)
		time.Sleep(360000 * time.Second)
	}

}

func CheckInstanceStatus(instanceList []string) []k8splugin.InstanceStatus {

	var instStatusList []k8splugin.InstanceStatus

	for _, instance := range instanceList {

		instanceStatus := executor.CheckStatusForEachInstance(string(instance))

		instStatusList = append(instStatusList, instanceStatus)

	}

	return instStatusList
}

func PushPodInfoToAAI(podList []con.PodInfoToAAI) {

	var relList []con.RelationList

	for _, pod := range podList {

		connection := executor.GetConnection(pod.CloudRegion)

		tenantId := executor.GetTenant(connection.CloudOwner, pod.CloudRegion)

		vserverID := executor.PushVservers(pod, connection.CloudOwner, pod.CloudRegion, tenantId)

		rl := utils.BuildRelationshipDataForVFModule(pod.VserverName, vserverID, connection.CloudOwner, pod.CloudRegion, tenantId)
		relList = append(relList, rl)

		executor.LinkVserverVFM(pod.VnfId, pod.VfmId, connection.CloudOwner, pod.CloudRegion, tenantId, relList)
	}

}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	//go QueryAAI()

	<-c

}
