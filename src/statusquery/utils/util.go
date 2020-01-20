package utils


import (
        con "github.com/onap/multicloud-k8s/src/statusquery/constants"
        "reflect"

        "fmt"
)



func BuildRelationshipDataForVFModule(vserverName, vserverID, cloudOwner, cloudRegion, tenantId string) con.RelationList {

        //logrus.Debug("Executing buildRelationshipDataForVFModule")

        rl := con.RelationList{"vserver", "/aai/v14/cloud-infrastructure/cloud-regions/cloud-region/" + cloudOwner + "/" + cloudRegion + "/tenants/tenant/" + tenantId + "/vservers/vserver/" + vserverID, []con.RData{con.RData{"cloud-region.cloud-owner", cloudOwner},
                con.RData{"cloud-region.cloud-region-id", cloudRegion},
                con.RData{"tenant.tenant-id", tenantId},
                con.RData{"vserver.vserver-id", vserverID}},
                []con.Property{con.Property{"vserver.vserver-name", vserverName}}}

        return rl

}

func ParseListInstanceResponse(rlist []con.InstanceMiniResponse) []string {

        //logrus.Debug("Executing parseListInstanceResponse - to parse response")

        var resourceIdList []string

        //assume there is only one resource created
        for _, result := range rlist {

                resourceIdList = append(resourceIdList, result.ID)
        }

        return resourceIdList
}


func ParseStatusInstanceResponse(instanceStatusses []con.InstanceStatus) []con.PodInfoToAAI {

        //logrus.Debug("Executing ProcessInstanceRequest - to parse Status API response")

        var infoToAAI []con.PodInfoToAAI

        for _, instanceStatus := range instanceStatusses {

                var podInfo con.PodInfoToAAI

                sa := reflect.ValueOf(&instanceStatus).Elem()
                typeOf := sa.Type()
                for i := 0; i < sa.NumField(); i++ {
                        f := sa.Field(i)
                        fmt.Printf("%d: %s %s = %v\n", i,
                                typeOf.Field(i).Name, f.Type(), f.Interface())
                        if typeOf.Field(i).Name == "Request" {
                                fmt.Printf("it's a Request object \n")
                                request := f.Interface()
                                if ireq, ok := request.(con.InstanceRequest); ok {
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
                                if pss, ok := ready.([]con.PodStatus); ok {
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
