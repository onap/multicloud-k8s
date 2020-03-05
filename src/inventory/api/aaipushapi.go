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
	"bytes"
	"crypto/tls"
	"encoding/json"
	con "github.com/onap/multicloud-k8s/src/inventory/constants"
	log "github.com/onap/multicloud-k8s/src/inventory/logutils"
	util "github.com/onap/multicloud-k8s/src/inventory/utils"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

/* Pushes each pod related details as vservers inot A&AI

{
                "vserver-id": "example20",
                "vserver-name": "POD-NAME",
                "vserver-name2": "Relese-name/Profile-name of the POD (Labels:release=profile-k8s)",
                "prov-status": "NAMESPACEofthPOD",
                "vserver-selflink": "example-vserver-selflink-val-57201",
                "in-maint": true,
                "is-closed-loop-disabled": true,
                "l-interfaces": {
                                "l-interface": [{
                                                "interface-name": "example-interface-name-val-20080",
												"is-port-mirrored": true,
												"in-maint": true,
												"is-ip-unnumbered": true,
                                                "l3-interface-ipv4-address-list": [{
                                                                "l3-interface-ipv4-address": "IP_Address",
                                                                "l3-interface-ipv4-prefix-length": "PORT"
                                                }]
                                }]
                }
}

*/
func PushVservers(podInfo con.PodInfoToAAI, cloudOwner, cloudRegion, tenantId string) string {

	AAI_URI := os.Getenv("onap-aai")
	AAI_Port := os.Getenv("aai-port")

	payload := "{\"vserver-name\":" + "\"" + podInfo.VserverName + "\"" + ", \"vserver-name2\":" + "\"" + podInfo.VserverName2 + "\"" + ", \"prov-status\":" + "\"" + podInfo.ProvStatus + "\"" + ",\"vserver-selflink\":" + "\"example-vserver-selflink-val-57201\", \"l-interfaces\": {\"l-interface\": [{\"interface-name\": \"example-interface-name-val-20080\",\"is-port-mirrored\": true,\"in-maint\": true,\"is-ip-unnumbered\": true,\"l3-interface-ipv4-address-list\": [{\"l3-interface-ipv4-address\":" + "\"" + podInfo.I3InterfaceIPv4Address + "\"" + ",\"l3-interface-ipv4-prefix-length\":" + "\"" + strconv.FormatInt(int64(podInfo.I3InterfaceIPvPrefixLength), 10) + "\"" + "}]}]}}"

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	url := AAI_URI + ":" + AAI_Port + con.AAI_EP + "cloud-region/" + cloudOwner + "/" + cloudRegion + "/tenants/tenant/" + tenantId + "/vservers/vserver/" + podInfo.VserverName

	var jsonStr = []byte(payload)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

	if err != nil {
		log.Error("Error while constructing Vserver PUT request")
		return
	}

	util.SetRequestHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("Error while executing Vserver PUT api")
		return
	}
	defer resp.Body.Close()

	return podInfo.VserverName
}

/* This links vservers to vf-module request payload */
func LinkVserverVFM(vnfID, vfmID, cloudOwner, cloudRegion, tenantId string, relList []con.RelationList) {

	AAI_URI := os.Getenv("onap-aai")
	AAI_Port := os.Getenv("aai-port")

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	apiToCR := AAI_URI + ":" + AAI_Port + con.AAI_EP + con.AAI_NEP + "/" + vnfID + "/vf-modules"
	req, err := http.NewRequest(http.MethodGet, apiToCR, nil)
	if err != nil {
		log.Error("Error while constructing VFModules GET api request")
		return

	}

	util.SetRequestHeaders(req)

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {
		log.Error("Error while executing VFModules GET api")
		return
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {

		log.Error("Error while reading vfmodules API response")
		return

	}

	var vfmodules con.VFModules

	json.Unmarshal([]byte(body), &vfmodules)

	vfmList := vfmodules.VFModules

	for key, vfmodule := range vfmList {

		if vfmodule.VFModuleId == vfmID {

			vfmodule.RelationshipList = map[string][]con.RelationList{"relationship": relList}

			vfmList = append(vfmList, vfmodule)

			vfmList[key] = vfmList[len(vfmList)-1] // Copy last element to index i.
			vfmList = vfmList[:len(vfmList)-1]

			//update vfmodule with vserver data

			vfmPayload, err := json.Marshal(vfmodule)

			if err != nil {

				log.Error("Error while marshalling vfmodule linked vserver info response")
				return

			}

			pushVFModuleToAAI(string(vfmPayload), vfmID, vnfID)
		}

	}

}

/*  Pushes vf-module enriched with vserver information */
func pushVFModuleToAAI(vfmPayload, vfmID, vnfID string) {

	AAI_URI := os.Getenv("onap-aai")
	AAI_Port := os.Getenv("aai-port")

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	url := AAI_URI + ":" + AAI_Port + con.AAI_NEP + vnfID + "/vfmodules/vf-module/" + vfmID

	var jsonStr = []byte(vfmPayload)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

	if err != nil {
		log.Error("Error while constructing a VFModule request to AAI")
		return
	}

	util.SetRequestHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		log.Error("Error while executing PUT request of VFModule to AAI")
		return

	}

	defer resp.Body.Close()

}
