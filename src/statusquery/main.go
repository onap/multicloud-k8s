package main

import (
        "encoding/json"
        con "github.com/VamshiKrishnaNemalikonda/GO/stringutil/constants"
        corev1 "k8s.io/api/core/v1"
        "os"
        "os/signal"
        "reflect"

        "bytes"
        "crypto/tls"
        "fmt"
        "io/ioutil"
        "net/http"
        "strconv"
        "time"
)

type InstanceRequest struct {
        RBName      string            `json:"rb-name"`
        RBVersion   string            `json:"rb-version"`
        ProfileName string            `json:"profile-name"`
        CloudRegion string            `json:"cloud-region"`
        Labels      map[string]string `json:"labels"`
}

type InstanceMiniResponse struct {
        ID        string          `json:"id"`
        Request   InstanceRequest `json:"request"`
        Namespace string          `json:"namespace"`
}

type PodStatus struct {
        Name        string           `json:"name"`
        Namespace   string           `json:"namespace"`
        Ready       bool             `json:"ready"`
        Status      corev1.PodStatus `json:"status,omitempty"`
        IPAddresses []string         `json:"ipaddresses"`
}

type InstanceStatus struct {
        Request         InstanceRequest  `json:"request"`
        Ready           bool             `json:"ready"`
        ResourceCount   int32            `json:"resourceCount"`
        PodStatuses     []PodStatus      `json:"podStatuses"`
        ServiceStatuses []corev1.Service `json:"serviceStatuses"`
}

type PodInfoToAAI struct {
        VserverName                string
        VserverName2               string
        ProvStatus                 string
        I3InterfaceIPv4Address     string
        I3InterfaceIPvPrefixLength int32
        VnfId                      string
        VfmId                      string
}

type RData struct {
        RelationshipKey   string `json:"relationship-key"`
        RelationshipValue string `json:"relationship-value"`
}

type RelationList struct {
        RelatedTo         string     `json:"related-to"`
        RelatedLink       string     `json:"related-link"`
        RelationshipData  []RData    `json:"relationship-data"`
        RelatedToProperty []Property `json:"related-to-property"`
}

type CRegion struct {
        CloudOwner         string                    `json:"cloud-owner"`
        CloudRegionId      string                    `json:"cloud-region-id"`
        CloudType          string                    `json:"cloud-type"`
        OwnerDefinedType   string                    `json:"owner-defined-type"`
        CloudRegionVersion string                    `json:"cloud-region-version"`
        CloudZone          string                    `json:"cloud-zone"`
        ResourceVersion    string                    `json:"resource-version"`
        ComplexName        string                    `json:"complex-name"`
        SriovAutomation    string                    `json:"sriov-automation"`
        CloudExtraInfo     string                    `json:"cloud-extra-info"`
        RelationshipList   map[string][]RelationList `json:"relationship-list"`
}

type CloudRegion struct {
        Regions []CRegion `json:"cloud-region"`
}

type TenantInfo struct {
        TenantId   string `json:"tenant-id"`
        TenantName string `json:"tenant-name"`
}

type Tenant struct {
        Tenants map[string][]TenantInfo `json:"tenants"`
}

type Property struct {
        PropertyKey   string `json:"property-key"`
        PropertyValue string `json:"property-value"`
}

type VFModule struct {
        VFModuleId           string                    `json:"vf-module-id"`
        VFModuleName         string                    `json:"vf-module-name"`
        HeatStackId          string                    `json:"heat-stack-id"`
        OrchestrationStatus  string                    `json:"orchestration-status"`
        ResourceVersion      string                    `json:"resource-version"`
        AutomatedAssignment  string                    `json:"automated-assignment"`
        IsBaseVfModule       string                    `json:"is-base-vf-module"`
        RelationshipList     map[string][]RelationList `json:"relationship-list"`
        ModelInvariantId     string                    `json:"model-invariant-id"`
        ModelVersionId       string                    `json:"model-version-id"`
        ModelCustomizationId string                    `json:"model-customization-id"`
        ModuleIndex          string                    `json:"module-index"`
}

type VFModules struct {
        VFModules []VFModule `json:"vf-module"`
}

func QueryAAI() {

        //logrus.Debug("executing QueryStatusAPI")

        instanceID := ListInstances()
        instanceStatus := CheckInstanceStatus(instanceID)
        podInfo := ParseStatusInstanceResponse(instanceStatus)

        for {

                PushPodInfoToAAI(podInfo, "vnf-id", "vfm-id")
                time.Sleep(360000 * time.Second)

        }

}

//righ now ListInstances method returning only once instance information - TBD, to loop for all the resources created
func ListInstances() []string {

        //logrus.Debug("ListInstances")

        instancelist := con.MK8S_URI + ":" + con.MK8S_Port + con.MK8S_EP
        req, err := http.NewRequest(http.MethodGet, instancelist, nil)
        if err != nil {

                //logrus.Error("Something went wrong while listing resources")
        }

        client := http.DefaultClient
        res, err := client.Do(req)

        if err != nil {
                //logrus.Error("Something went wrong while listing resources")
        }
        //logrus.Debug("List instances request processed")

        /*body, err := ioutil.ReadAll(res.Body)
          if err != nil {
                  panic(err)
          }
          fmt.Printf("%v", string(body))*/

        defer res.Body.Close()

        decoder := json.NewDecoder(res.Body)
        var rlist []InstanceMiniResponse
        err = decoder.Decode(&rlist)

        resourceList := parseListInstanceResponse(rlist)

        return resourceList

}

// Process the output of list instances to return specific instance ID
func parseListInstanceResponse(rlist []InstanceMiniResponse) []string {

        //logrus.Debug("Executing parseListInstanceResponse - to parse response")

        var resourceIdList []string

        //assume there is only one resource created
        for _, result := range rlist {

                resourceIdList = append(resourceIdList, result.ID)
        }

        return resourceIdList
}

func CheckInstanceStatus(instanceList []string) []InstanceStatus {

        var instStatusList []InstanceStatus

        for _, instance := range instanceList {

                instanceStatus := checkStatusForEachInstance(string(instance))

                instStatusList = append(instStatusList, instanceStatus)

        }

        return instStatusList
}

func checkStatusForEachInstance(instanceID string) InstanceStatus {

        //logrus.Debug("Executing checkStatusForEachInstance")

        instancelist := con.MK8S_URI + ":" + con.MK8S_Port + con.MK8S_EP + instanceID + "/status"

        req, err := http.NewRequest(http.MethodGet, instancelist, nil)
        if err != nil {
                //logrus.Error("Error while building http request")
        }

        client := http.DefaultClient
        resp, err := client.Do(req)
        if err != nil {

                //logrus.Error("Error while making rest request")
        }

        defer resp.Body.Close()

        decoder := json.NewDecoder(resp.Body)
        var instStatus InstanceStatus
        err = decoder.Decode(&instStatus)

        /*body, err := ioutil.ReadAll(resp.Body)
          if err != nil {
                  panic(err)
          }
          fmt.Printf("%v", string(body)) */

        return instStatus
}

func ParseStatusInstanceResponse(instanceStatusses []InstanceStatus) []PodInfoToAAI {

        //logrus.Debug("Executing ProcessInstanceRequest - to parse Status API response")

        var infoToAAI []PodInfoToAAI

        for _, instanceStatus := range instanceStatusses {

                var podInfo PodInfoToAAI

                sa := reflect.ValueOf(&instanceStatus).Elem()
                typeOf := sa.Type()
                for i := 0; i < sa.NumField(); i++ {
                        f := sa.Field(i)
                        fmt.Printf("%d: %s %s = %v\n", i,
                                typeOf.Field(i).Name, f.Type(), f.Interface())
                        if typeOf.Field(i).Name == "Request" {
                                fmt.Printf("it's a Request object \n")
                                request := f.Interface()
                                if ireq, ok := request.(InstanceRequest); ok {
                                        podInfo.VserverName2 = ireq.ProfileName

                                        for key, value := range ireq.Labels {
                                                if key == "vnf_id" {

                                                        podInfo.VnfId = value

                                                }
                                                if key == "vfm_id" {

                                                        podInfo.VfmId = value

                                                }
                                        }

                                } else {
                                        fmt.Printf("it's not a InstanceRequest \n")
                                }
                        }

                        if typeOf.Field(i).Name == "PodStatuses" {
                                fmt.Printf("it's a []PodStatus parameter \n")
                                ready := f.Interface()
                                if pss, ok := ready.([]PodStatus); ok {
                                        for _, ps := range pss {
                                                podInfo.VserverName = ps.Name
                                                podInfo.ProvStatus = ps.Namespace
                                                //fmt.Printf("%v\n", ps.IPAddresses)
                                        }

                                } else {
                                        fmt.Printf("it's not a InstanceRequest \n")
                                }
                        }
                }

                infoToAAI = append(infoToAAI, podInfo)

        }

        return infoToAAI

}

func PushPodInfoToAAI(podInfoToAAI []PodInfoToAAI, vnfID, vfmID string) {

        //logrus.Debug("Executing PushPodInfoToAAI ")

        var cloudOwner string
        var cloudRegion string

        cr := getCloudRegion()

        for _, cregion := range cr.Regions {
                if cregion.CloudType == "k8s" {

                        cloudOwner = cregion.CloudOwner
                        cloudRegion = cregion.CloudRegionId

                }

        }

        tenantId := getTenant(cloudOwner, cloudRegion)

        if &tenantId == nil {

                //logrus.Error("Tenant information not found")

        }

        var relList []RelationList

        for _, eachPod := range podInfoToAAI {

                vserverID := pushToAAI(eachPod, cloudOwner, cloudRegion, tenantId)

                rl := buildRelationshipDataForVFModule(eachPod.VserverName, vserverID, cloudOwner, cloudRegion, tenantId)
                relList = append(relList, rl)

                linkVserverVFM(eachPod.VnfId, eachPod.VfmId, cloudOwner, cloudRegion, tenantId, relList)
        }

        //linkVserverVFM(vnfID, vfmID, cloudOwner, cloudRegion, tenantId, relList)

}

func buildRelationshipDataForVFModule(vserverName, vserverID, cloudOwner, cloudRegion, tenantId string) RelationList {

        //logrus.Debug("Executing buildRelationshipDataForVFModule")

        rl := RelationList{"vserver", "/aai/v14/cloud-infrastructure/cloud-regions/cloud-region/" + cloudOwner + "/" + cloudRegion + "/tenants/tenant/" + tenantId + "/vservers/vserver/" + vserverID, []RData{RData{"cloud-region.cloud-owner", cloudOwner},
                RData{"cloud-region.cloud-region-id", cloudRegion},
                RData{"tenant.tenant-id", tenantId},
                RData{"vserver.vserver-id", vserverID}},
                []Property{Property{"vserver.vserver-name", vserverName}}}

        return rl

}

func pushToAAI(podInfo PodInfoToAAI, cloudOwner, cloudRegion, tenantId string) string {

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

func getCloudRegion() CloudRegion {

        //logrus.Debug("executing getCloudRegion")

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

        apiToCR := con.AAI_URI + ":" + con.AAI_Port + con.AAI_EP

        fmt.Println(apiToCR)

        req, err := http.NewRequest(http.MethodGet, apiToCR, nil)
        if err != nil {
                panic(err)
        }

        req.Header.Set("X-FromAppId", con.XFromAppId)
        req.Header.Set("Content-Type", con.ContentType)
        req.Header.Set("Accept", con.Accept)
        req.Header.Set("X-TransactionId", con.XTransactionId)
        req.Header.Set("Authorization", con.Authorization)

        client := http.DefaultClient
        res, err := client.Do(req)

        if err != nil {
                fmt.Printf("Something went wrong while retrieving Cloud region details")
        }

        //fmt.Printf("%v", string(body))

        defer res.Body.Close()

        //crJson := json.NewDecoder(res.Body)
        //var rlist []InstanceMiniResponse
        //err = decoder.Decode(&rlist)

        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
                panic(err)
        }
        //      fmt.Printf("%v", string(body))

        var cr CloudRegion

        json.Unmarshal([]byte(body), &cr)

        return cr

}

func getTenant(cloudOwner, cloudRegion string) string {

        //logrus.Debug("Executing getTenant")

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

        apiToCR := con.AAI_URI + ":" + con.AAI_Port + con.AAI_EP + "cloud-region/" + cloudOwner + "/" + cloudRegion + "?depth=all"
        req, err := http.NewRequest(http.MethodGet, apiToCR, nil)
        if err != nil {
                panic(err)
        }

        req.Header.Set("X-FromAppId", con.XFromAppId)
        req.Header.Set("Content-Type", con.ContentType)
        req.Header.Set("Accept", con.Accept)
        req.Header.Set("X-TransactionId", con.XTransactionId)
        req.Header.Set("Authorization", con.Authorization)

        client := http.DefaultClient
        res, err := client.Do(req)

        if err != nil {
                fmt.Printf("Something went wrong while getting Tenant details")
        }

        //fmt.Printf("%v", string(body))

        defer res.Body.Close()

        //crJson := json.NewDecoder(res.Body)
        //var rlist []InstanceMiniResponse
        //err = decoder.Decode(&rlist)

        body, err := ioutil.ReadAll(res.Body)
        if err != nil {
                panic(err)
        }
        //      fmt.Printf("%v", string(body))

        var tenant Tenant

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

func linkVserverVFM(vnfID, vfmID, cloudOwner, cloudRegion, tenantId string, relList []RelationList) {

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

        var vfmodules VFModules

        json.Unmarshal([]byte(body), &vfmodules)

        vfmList := vfmodules.VFModules

        for key, vfmodule := range vfmList {

                if vfmodule.VFModuleId == vfmID {

                        //logrus.Debug("vfmodule identified")

                        vfmodule.RelationshipList = map[string][]RelationList{"relationship": relList}

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

func main() {

        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt)
        fmt.Println("executing main")

        go QueryAAI()

        //podInfoToAAI := PodInfoToAAI{"testing31", "testing10", "testing10", "testing10", "10"}

        //PushPodInfoToAAI(podInfoToAAI)

        //QueryAAI()

        fmt.Println("awaiting signal")
        sig := <-c
        fmt.Println(sig)
        fmt.Println("exiting")

}
