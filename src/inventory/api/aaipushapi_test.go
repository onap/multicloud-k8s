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
	con "github.com/onap/multicloud-k8s/src/inventory/constants"
	util "github.com/onap/multicloud-k8s/src/inventory/utils"
	"net/http"
	"os"
	"testing"
)

func TestPushVservers(t *testing.T) {

	cloudOwner := "CloudOwner"
	cloudRegion := "RegionOne"
	tenantId := "tenant123"

	AAI_URI := os.Getenv("onap-aai")
	AAI_Port := os.Getenv("aai-port")

	payload := "{\"vserver-name\":" + "\"" + "pod123" + "\"" + ", \"vserver-name2\":" + "\"" + "profile123" + "\"" + ", \"prov-status\":" + "\"" + "default" + "\"" + ",\"vserver-selflink\":" + "\"example-vserver-selflink-val-57201\", \"l-interfaces\": {\"l-interface\": [{\"interface-name\": \"example-interface-name-val-20080\",\"is-port-mirrored\": true,\"in-maint\": true,\"is-ip-unnumbered\": true,\"l3-interface-ipv4-address-list\": [{\"l3-interface-ipv4-address\":" + "\"" + "10.214.22.220" + "\"" + ",\"l3-interface-ipv4-prefix-length\":" + "\"" + "667" + "\"" + "}]}]}}"

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	url := AAI_URI + ":" + AAI_Port + con.AAI_EP + "cloud-region/" + cloudOwner + "/" + cloudRegion + "/tenants/tenant/" + tenantId + "/vservers/vserver/" + "pod123"

	var jsonStr = []byte(payload)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

	if err != nil {

		t.Error("Failed: Error consructing request")
	}

	util.SetRequestHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		t.Error("Failed: Error while executing request ")
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("recieved unexpeted response ")
	}

}

func TestLinkVserverVFM(t *testing.T) {

	vnfID := "vnf123456"
	AAI_URI := os.Getenv("onap-aai")
	AAI_Port := os.Getenv("aai-port")

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	apiToCR := AAI_URI + ":" + AAI_Port + con.AAI_EP + con.AAI_NEP + "/" + vnfID + "/vf-modules"
	req, err := http.NewRequest(http.MethodGet, apiToCR, nil)
	if err != nil {

		t.Error("Failed: Error while constructing new request")

	}

	util.SetRequestHeaders(req)

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {

		t.Error("Failed: Error while executing request")
	}

	if res.StatusCode != http.StatusOK {
		t.Fatalf("recieved unexpeted response ")
	}

}

func TestPushVFModuleToAAI(t *testing.T) {

	vnfID := "vnf123456"
	vfmID := "vfm123456"
	vfmPayload := ""

	AAI_URI := os.Getenv("onap-aai")
	AAI_Port := os.Getenv("aai-port")

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	url := AAI_URI + ":" + AAI_Port + con.AAI_NEP + vnfID + "/vfmodules/vf-module/" + vfmID

	var jsonStr = []byte(vfmPayload)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

	if err != nil {
		t.Error("Failed: Error while executing request")

	}

	util.SetRequestHeaders(req)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {

		t.Error("Failed: Error while executing request")
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("recieved unexpeted response ")
	}

}
