package main

import (
        con "github.com/onap/multicloud-k8s/src/statusquery/constants"
        executor "github.com/onap/multicloud-k8s/src/statusquery/apiexecutor"
        utils "github.com/onap/multicloud-k8s/src/statusquery/utils"
        "os"
        "os/signal"
        "time"
        "fmt"

)


func QueryAAI() {

        //logrus.Debug("executing QueryStatusAPI")

        instanceID := executor.ListInstances()
        instanceStatus := CheckInstanceStatus(instanceID)
        podInfo := utils.ParseStatusInstanceResponse(instanceStatus)

        for {

                PushPodInfoToAAI(podInfo, "vnf-id", "vfm-id")
                time.Sleep(360000 * time.Second)

        }

}


func CheckInstanceStatus(instanceList []string) []con.InstanceStatus {

        var instStatusList []con.InstanceStatus

        for _, instance := range instanceList {

                instanceStatus := executor.CheckStatusForEachInstance(string(instance))

                instStatusList = append(instStatusList, instanceStatus)

        }

        return instStatusList
}



func PushPodInfoToAAI(podInfoToAAI []con.PodInfoToAAI, vnfID, vfmID string) {

        //logrus.Debug("Executing PushPodInfoToAAI ")

        var cloudOwner string
        var cloudRegion string

        cr := executor.GetCloudRegion()

        for _, cregion := range cr.Regions {
                if cregion.CloudType == "k8s" {

                        cloudOwner = cregion.CloudOwner
                        cloudRegion = cregion.CloudRegionId

                }

        }

        tenantId := executor.GetTenant(cloudOwner, cloudRegion)

        if &tenantId == nil {

                //logrus.Error("Tenant information not found")

        }

        var relList []con.RelationList

        for _, eachPod := range podInfoToAAI {

                vserverID := executor.PushToAAI(eachPod, cloudOwner, cloudRegion, tenantId)

                rl := utils.BuildRelationshipDataForVFModule(eachPod.VserverName, vserverID, cloudOwner, cloudRegion, tenantId)
                relList = append(relList, rl)

                executor.LinkVserverVFM(eachPod.VnfId, eachPod.VfmId, cloudOwner, cloudRegion, tenantId, relList)
        }

        //linkVserverVFM(vnfID, vfmID, cloudOwner, cloudRegion, tenantId, relList)

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
