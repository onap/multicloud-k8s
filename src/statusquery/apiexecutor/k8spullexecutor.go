
package apiexecutor

import (

        con "github.com/onap/multicloud-k8s/src/statusquery/constants"
        utils "github.com/onap/multicloud-k8s/src/statusquery/utils"

        "net/http"
        "encoding/json"

)


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
        var rlist []con.InstanceMiniResponse
        err = decoder.Decode(&rlist)

        resourceList := utils.ParseListInstanceResponse(rlist)

        return resourceList

}

func CheckStatusForEachInstance(instanceID string) con.InstanceStatus {

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
        var instStatus con.InstanceStatus
        err = decoder.Decode(&instStatus)

        /*body, err := ioutil.ReadAll(resp.Body)
          if err != nil {
                  panic(err)
          }
          fmt.Printf("%v", string(body)) */

        return instStatus
}

