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
	"crypto/tls"
	"encoding/json"
	"fmt"
	con "github.com/onap/multicloud-k8s/src/inventory/constants"
	util "github.com/onap/multicloud-k8s/src/inventory/utils"
	"io/ioutil"
	"net/http"
)

func GetTenant(cloudOwner, cloudRegion string) string {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	apiToCR := con.AAI_URI + ":" + con.AAI_Port + con.AAI_EP + "cloud-region/" + cloudOwner + "/" + cloudRegion + "?depth=all"
	req, err := http.NewRequest(http.MethodGet, apiToCR, nil)
	if err != nil {
		fmt.Printf("No tenants found for cloud region")

	}

	util.SetRequestHeaders(req)

	client := http.DefaultClient
	res, err := client.Do(req)

	if err != nil {
		fmt.Printf("Something went wrong while getting Tenant details")
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {

		fmt.Printf("Can't read Tenant response")

	}

	var tenant con.Tenant

	json.Unmarshal([]byte(body), &tenant)

	for k, v := range tenant.Tenants {
		if k == "tenant" {
			for _, val := range v {
				fmt.Printf(val.TenantId)
				return val.TenantId

			}
		}
	}

	return ""

}
