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

package api

import (
	con "github.com/onap/multicloud-k8s/src/inventory/constants"
	log "github.com/onap/multicloud-k8s/src/inventory/logutils"
	utils "github.com/onap/multicloud-k8s/src/inventory/utils"
	k8sint "github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	k8scon "github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"

	"encoding/json"
	"net/http"
	"os"
)

func ListInstances() []string {

	MK8S_URI := os.Getenv("onap-multicloud-k8s")
	MK8S_Port := os.Getenv("multicloud-k8s-port")

	instancelist := MK8S_URI + ":" + MK8S_Port + con.MK8S_EP
	req, err := http.NewRequest(http.MethodGet, instancelist, nil)
	if err != nil {

		log.Error("Something went wrong while listing resources - contructing request")
		return
	}

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {
		log.Error("Something went wrong while listing resources - executing request")
		return
	}

	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var rlist []k8sint.InstanceMiniResponse
	err = decoder.Decode(&rlist)

	resourceList := utils.ParseListInstanceResponse(rlist)

	return resourceList

}

func GetConnection(cregion string) k8scon.Connection {

	MK8S_URI := os.Getenv("onap-multicloud-k8s")
	MK8S_Port := os.Getenv("multicloud-k8s-port")

	instancelist := MK8S_URI + ":" + MK8S_Port + con.MK8S_CEP + cregion
	req, err := http.NewRequest(http.MethodGet, instancelist, nil)
	if err != nil {

		log.Error("Something went wrong while getting Connection resource - contructing request")
		return
	}

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {
		log.Error("Something went wrong while getting Connection resource - executing request")
		return
	}

	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var connection k8scon.Connection
	err = decoder.Decode(&connection)

	return connection

}

func CheckStatusForEachInstance(instanceID string) k8sint.InstanceStatus {

	MK8S_URI := os.Getenv("onap-multicloud-k8s")
	MK8S_Port := os.Getenv("multicloud-k8s-port")

	instancelist := MK8S_URI + ":" + MK8S_Port + con.MK8S_EP + instanceID + "/status"

	req, err := http.NewRequest(http.MethodGet, instancelist, nil)
	if err != nil {
		log.Error("Error while checking instance status - building http request")
		return
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {

		log.Error("Error while checking instance status - making rest request")
		return
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var instStatus k8sint.InstanceStatus
	err = decoder.Decode(&instStatus)

	return instStatus
}
