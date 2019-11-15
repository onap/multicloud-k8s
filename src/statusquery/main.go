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

        instancelist := "http://10.211.1.20:30280/api/multicloud-k8s/v1/v1/instance"
        req, err := http.NewRequest(http.MethodGet, instancelist, nil)
        if err != nil {
                panic(err)
        }

        client := http.DefaultClient
        res, err := client.Do(req)

        if err != nil {
                fmt.Printf("Something went wring while listing resources")
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

        instancelist := "http://10.211.1.20:30280/api/multicloud-k8s/v1/v1/instance/" + instanceID + "/status"

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

        fmt.Printf("Executing PushPodInfoToAAI ")

        //Need to parse the podInfoToAAI to input AAI api.
        fmt.Printf("You can check podInfoToAAI structure details below")
        fmt.Printf(podInfo.VserverName)
        fmt.Printf(podInfo.VserverName2)
        fmt.Printf(podInfo.ProvStatus)

        http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
        url := "https://10.212.1.30:30233/aai/v14/cloud-infrastructure/cloud-regions/cloud-region/CloudOwner/RegionOne/tenants/tenant/0b82ba13bb88428bbfd14f0c2c9177c7/vservers/vserver/k8spod-test"

        var jsonStr = []byte(`{"vserver-id": "k8spod-test","vserver-name": "k8spod-test","vserver-selflink": "example-vserver-selflink-val-57201-test"}`)

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

        //log.Debug("response Status:" + resp.Status)
        //log.Debug("response Headers:" + resp.Header)
        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
        }
        fmt.Printf("response Body:" + string(body))

}

func main() {

        //c := make(chan os.Signal, 1)
        //signal.Notify(c, os.Interrupt)
        fmt.Println("executing main")

        //go QueryAAI()

        QueryAAI()

        //fmt.Println("awaiting signal")
        //sig := <-c
        //fmt.Println(sig)
        //fmt.Println("exiting")

}
