package main

import (
        "bytes"
        "crypto/tls"
        "fmt"
        "io/ioutil"
        "net/http"
        "reflect"
        //"os"
        //"os/signal"
        //"time"
        "encoding/json"
        con "github.com/onap/multicloud-k8s/src/statusquery/constants"
        corev1 "k8s.io/api/core/v1"
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
        I3InterfaceIPvPrefixLength string
}

type RData struct {
        RelationshipKey   string `json:"relationship-key"`
        RelationshipValue string `json:"relationship-value"`
}

type RelationList struct {
        RelatedTo         string  `json:"related-to"`
        RelationshipLabel string  `json:"relationship-label"`
        RelatedLink       string  `json:"related-link"`
        RelationshipData  []RData `json:"relationship-data"`
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

func QueryAAI() {

        fmt.Println("executing QueryStatusAPI")

        instanceID := ListInstances()
        instanceStatus := CheckInstanceStatus(instanceID)
        podInfo := ParseStatusInstanceResponse(instanceStatus)

        for _, reqData := range podInfo {

                PushPodInfoToAAI(reqData)

                //time.Sleep(360000 * time.Second)
        }

}

//righ now ListInstances method returning only once instance information - TBD, to loop for all the resources created
func ListInstances() []string {

        fmt.Println("ListInstances")

        instancelist := con.MK8S_URI + ":" + con.MK8S_Port + con.MK8S_Endpoint
        req, err := http.NewRequest(http.MethodGet, instancelist, nil)
        if err != nil {
                panic(err)
        }

        client := http.DefaultClient
        res, err := client.Do(req)

        if err != nil {
                fmt.Printf("Something went wrong while listing resources")
        }
        fmt.Println("List instances request processed")

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

        fmt.Printf("Executing parseListInstanceResponse - to parse response")

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

        instancelist := con.MK8S_URI + ":" + con.MK8S_Port + con.MK8S_Endpoint + instanceID + "/status"

        req, err := http.NewRequest(http.MethodGet, instancelist, nil)
        if err != nil {
                panic(err)
        }

        client := http.DefaultClient
        resp, err := client.Do(req)
        if err != nil {
                panic(err)
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

        return InstanceStatus{}
}

func ParseStatusInstanceResponse(instanceStatusses []InstanceStatus) []PodInfoToAAI {

        fmt.Printf("Executing ProcessInstanceRequest - to parse Status API response")

        for _, instanceStatus := range instanceStatusses {

                var pita PodInfoToAAI
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
                                        pita.VserverName2 = ireq.ProfileName
                                } else {
                                        fmt.Printf("it's not a InstanceRequest \n")
                                }
                        }

                        if typeOf.Field(i).Name == "PodStatuses" {
                                fmt.Printf("it's a []PodStatus parameter \n")
                                ready := f.Interface()
                                if pss, ok := ready.([]PodStatus); ok {
                                        for _, ps := range pss {
                                                pita.VserverName = ps.Name
                                                pita.ProvStatus = ps.Namespace
                                                //fmt.Printf("%v\n", ps.IPAddresses)
                                        }

                                } else {
                                        fmt.Printf("it's not a InstanceRequest \n")
                                }
                        }
                }

        }

        return []PodInfoToAAI{}

}

func PushPodInfoToAAI(podInfo PodInfoToAAI) {

        fmt.Println("executing PushPodInfoToAAI")
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

        //Need to parse the podInfoToAAI to input AAI api.
        fmt.Printf("You can check podInfoToAAI structure details below")
        fmt.Printf(podInfo.VserverName)
        fmt.Printf(podInfo.VserverName2)
        fmt.Printf(podInfo.ProvStatus)
        fmt.Printf(podInfo.I3InterfaceIPv4Address)

        //Structure of payload we've as part of AAI request
        /*{
                          "vserver-id": "example",
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
          } */

        payload := "{\"vserver-name\":" + "\"" + podInfo.VserverName + "\"" + ", \"vserver-name2\":" + "\"" + podInfo.VserverName2 + "\"" + ", \"prov-status\":" + "\"" + podInfo.ProvStatus + "\"" + ",\"vserver-selflink\":" + "\"example-vserver-selflink-val-57201\", \"l-interfaces\": {\"l-interface\": [{\"interface-name\": \"example-interface-name-val-20080\",\"is-port-mirrored\": true,\"in-maint\": true,\"is-ip-unnumbered\": true,\"l3-interface-ipv4-address-list\": [{\"l3-interface-ipv4-address\":" + "\"" + podInfo.I3InterfaceIPv4Address + "\"" + ",\"l3-interface-ipv4-prefix-length\":" + "\"" + podInfo.I3InterfaceIPvPrefixLength + "\"" + "}]}]}}"

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        url := con.AAI_URI + ":" + con.AAI_Port + con.AAI_Endpoint + "cloud-region/" + cloudOwner + "/" + cloudRegion + "/tenants/tenant/" + tenantId + "/vservers/vserver/" + podInfo.VserverName

        fmt.Println(payload)
        fmt.Println(url)

        var jsonStr = []byte(payload)

        req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

        if err != nil {
                fmt.Printf("error while creating new request to AAI -> ")
                //fmt.Printf(err)
        }
        req.Header.Set("X-FromAppId", "SO")
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Accept", "application/json")
        req.Header.Set("X-TransactionId", "get_aai_subscr")
        req.Header.Set("Authorization", "Basic QUFJOkFBSQ==")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
                fmt.Printf("error while pushing to AAI -> ")
                //logrus.Errorln(err)
        }
        defer resp.Body.Close()

        fmt.Println("response Status:" + resp.Status)
        //log.Debug("response Headers:" + resp.Header)
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
        }
        fmt.Printf("response Body:" + string(body))

}

func getCloudRegion() CloudRegion {

        fmt.Println("executing getCloudRegion")

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

        apiToCR := con.AAI_URI + ":" + con.AAI_Port + con.AAI_Endpoint

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

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

        apiToCR := con.AAI_URI + ":" + con.AAI_Port + con.AAI_Endpoint + "cloud-region/" + cloudOwner + "/" + cloudRegion + "?depth=all"
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