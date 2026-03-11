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
	"github.com/onap/multicloud-k8s/src/inventory/model"
	utils "github.com/onap/multicloud-k8s/src/inventory/utils"
	log "github.com/sirupsen/logrus"

	"encoding/json"
	"net/http"
	"os"
)

func ListInstances() ([]string, error) {

	MK8S_URI := os.Getenv("onap-multicloud-k8s")
	MK8S_Port := os.Getenv("multicloud-k8s-port")

	instancelist := MK8S_URI + ":" + MK8S_Port + con.MK8S_EP
	req, err := http.NewRequest(http.MethodGet, instancelist, nil)
	if err != nil {

		log.Error("Something went wrong while listing resources - contructing request: ", err)
		return nil, err
	}

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {
		log.Error("Something went wrong while listing resources - executing request: ", err)
		return nil, err
	}

	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var rlist []model.InstanceMiniResponse
	err = decoder.Decode(&rlist)

	resourceList := utils.ParseListInstanceResponse(rlist)

	return resourceList, nil

}

func GetConnection(cregion string) (model.Connection, error) {

	MK8S_URI := os.Getenv("onap-multicloud-k8s")
	MK8S_Port := os.Getenv("multicloud-k8s-port")

	instancelist := MK8S_URI + ":" + MK8S_Port + con.MK8S_CEP + cregion
	req, err := http.NewRequest(http.MethodGet, instancelist, nil)
	if err != nil {

		log.Error("Something went wrong while getting Connection resource - contructing request: ", err)
		return model.Connection{}, err
	}

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {
		log.Error("Something went wrong while getting Connection resource - executing request: ", err)
		return model.Connection{}, err
	}

	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var connection model.Connection
	err = decoder.Decode(&connection)

	return connection, nil

}

func CheckStatusForEachInstance(instanceID string) model.InstanceStatus {

	MK8S_URI := os.Getenv("onap-multicloud-k8s")
	MK8S_Port := os.Getenv("multicloud-k8s-port")

	instancelist := MK8S_URI + ":" + MK8S_Port + con.MK8S_EP + instanceID + "/status"

	req, err := http.NewRequest(http.MethodGet, instancelist, nil)
	if err != nil {
		log.Error("Error while checking instance status - building http request: ", err)
		return model.InstanceStatus{}
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {

		log.Error("Error while checking instance status - making rest request: ", err)
		return model.InstanceStatus{}
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var instStatus model.InstanceStatus
	err = decoder.Decode(&instStatus)

	return instStatus
}
