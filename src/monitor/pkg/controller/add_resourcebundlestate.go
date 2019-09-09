package controller

import (
	"github.com/onap/multicloud-k8s/src/monitor/pkg/controller/resourcebundlestate"
)

func init() {
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.Add)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddPodController)
}
