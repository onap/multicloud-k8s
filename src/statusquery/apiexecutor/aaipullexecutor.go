package apiexecutor

import (

        con "github.com/onap/multicloud-k8s/src/statusquery/constants"
        "crypto/tls"
        "fmt"
        "io/ioutil"
        "net/http"
        "encoding/json"


)



func GetCloudRegion() con.CloudRegion {

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

        var cr con.CloudRegion

        json.Unmarshal([]byte(body), &cr)

        return cr

}

func GetTenant(cloudOwner, cloudRegion string) string {

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




