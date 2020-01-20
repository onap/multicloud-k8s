package apiexecutor

import (

        con "github.com/onap/multicloud-k8s/src/statusquery/constants"

        "bytes"
        "crypto/tls"
        "io/ioutil"
        "net/http"
        "strconv"
        "encoding/json"

)


func PushToAAI(podInfo con.PodInfoToAAI, cloudOwner, cloudRegion, tenantId string) string {

        //logrus.Debug("Executing pushToAAI")

        payload := "{\"vserver-name\":" + "\"" + podInfo.VserverName + "\"" + ", \"vserver-name2\":" + "\"" + podInfo.VserverName2 + "\"" + ", \"prov-status\":" + "\"" + podInfo.ProvStatus + "\"" + ",\"vserver-selflink\":" + "\"example-vserver-selflink-val-57201\", \"l-interfaces\": {\"l-interface\": [{\"interface-name\": \"example-interface-name-val-20080\",\"is-port-mirrored\": true,\"in-maint\": true,\"is-ip-unnumbered\": true,\"l3-interface-ipv4-address-list\": [{\"l3-interface-ipv4-address\":" + "\"" + podInfo.I3InterfaceIPv4Address + "\"" + ",\"l3-interface-ipv4-prefix-length\":" + "\"" + strconv.FormatInt(int64(podInfo.I3InterfaceIPvPrefixLength), 10) + "\"" + "}]}]}}"

        //logrus.Debug("payload to A&AI request : " + payload)

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        //url := "https://10.212.1.30:30233/aai/v14/cloud-infrastructure/cloud-regions/cloud-region/CloudOwner/RegionOne/tenants/tenant/0b82ba13bb88428bbfd14f0c2c9177c7/vservers/vserver/" + podInfo.VserverName

        url := con.AAI_URI + ":" + con.AAI_Port + con.AAI_EP + "cloud-region/" + cloudOwner + "/" + cloudRegion + "/tenants/tenant/" + tenantId + "/vservers/vserver/" + podInfo.VserverName

        var jsonStr = []byte(payload)

        req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

        if err != nil {
                //logrus.Error("Error while constructing Vserver PUT request")
        }
        req.Header.Set("X-FromAppId", "SO")
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Accept", "application/json")
        req.Header.Set("X-TransactionId", "get_aai_subscr")
        req.Header.Set("Authorization", "Basic QUFJOkFBSQ==")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
                //logrus.Error("Error while executing Vserver PUT api")
        }
        defer resp.Body.Close()

        //_ , err := ioutil.ReadAll(resp.Body)
        //if err != nil {
        //logrus.Error("Error while reading Vserver PUT response")

        //}
        //logrus.Debug("response Body:" + string(body))

        return podInfo.VserverName
}

func LinkVserverVFM(vnfID, vfmID, cloudOwner, cloudRegion, tenantId string, relList []con.RelationList) {

        //logrus.Debug("Executing linkVserverVFM")

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

        apiToCR := con.AAI_URI + ":" + con.AAI_Port + con.AAI_EP + con.AAI_NEP + "/" + vnfID + "/vf-modules"
        req, err := http.NewRequest(http.MethodGet, apiToCR, nil)
        if err != nil {
                //logrus.Error("Error while constructing VFModules GET api request")

        }

        req.Header.Set("X-FromAppId", "SO")
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Accept", "application/json")
        req.Header.Set("X-TransactionId", "get_aai_subscr")
        req.Header.Set("Authorization", "Basic QUFJOkFBSQ==")

        client := http.DefaultClient
        res, err := client.Do(req)

        if err != nil {
                //logrus.Error("Error while executing VFModules GET api")
        }

        //fmt.Printf("%v", string(body))

        defer res.Body.Close()

        //crJson := json.NewDecoder(res.Body)
        //var rlist []InstanceMiniResponse
        //err = decoder.Decode(&rlist)

        body, err := ioutil.ReadAll(res.Body)
        if err != nil {

                //logrus.Error("Error while reading vfmodules API response")

        }
        //      fmt.Printf("%v", string(body))

        var vfmodules con.VFModules

        json.Unmarshal([]byte(body), &vfmodules)

        vfmList := vfmodules.VFModules

        for key, vfmodule := range vfmList {

                if vfmodule.VFModuleId == vfmID {

                        //logrus.Debug("vfmodule identified")

                        vfmodule.RelationshipList = map[string][]con.RelationList{"relationship": relList}

                        vfmList = append(vfmList, vfmodule)

                        vfmList[key] = vfmList[len(vfmList)-1] // Copy last element to index i.
                        vfmList = vfmList[:len(vfmList)-1]

                        //update vfmodule with vserver data

                        vfmPayload, err := json.Marshal(vfmodule)

                        if err != nil {

                                //logrus.Error("Error while marshalling vfmodule linked vserver info response")

                        }

                        pushVFModuleToAAI(string(vfmPayload), vfmID, vnfID)
                }

        }

        //vfmodules.VFModules = vfmList

        //vfmPayload, err := json.Marshal(vfmodules)
        //if err != nil {

        //logrus.Error("Error while marshalling vfmodule linked vserver info response")

        //}

        //pushVFModuleToAAI(string(vfmPayload), vfmID, vnfID)

}

func pushVFModuleToAAI(vfmPayload, vfmID, vnfID string) {

        //logrus.Debug("Executing pushVFModuleToAAI")

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        url := con.AAI_URI + ":" + con.AAI_Port + con.AAI_NEP + vnfID + "/vfmodules/vf-module/" + vfmID

        //logrus.Debug(vfmPayload)

        var jsonStr = []byte(vfmPayload)

        req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

        if err != nil {
                //logrus.Error("Error while constructing a VFModule request to AAI")
        }
        req.Header.Set("X-FromAppId", "SO")
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Accept", "application/json")
        req.Header.Set("X-TransactionId", "get_aai_subscr")
        req.Header.Set("Authorization", "Basic QUFJOkFBSQ==")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {

                //logrus.Error("Error while executing PUT request of VFModule to AAI")

        }

        defer resp.Body.Close()

        //fmt.Println("response Status:" + resp.Status)
        //body, err := ioutil.ReadAll(resp.Body)
        //if err != nil {
        //logrus.Error("Error while reading VFModule response to AAI")

        //}
        //logrus.Debug("response Body:" + string(body))

}
