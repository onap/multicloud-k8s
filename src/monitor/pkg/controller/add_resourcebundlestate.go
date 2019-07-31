package controller

import (
	"monitor/pkg/controller/resourcebundlestate"
)

func init() {
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.Add)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddPodController)
}
