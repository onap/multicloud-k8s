package controller

import (
	"github.com/onap/multicloud-k8s/src/monitor/pkg/controller/resourcebundlestate"
)

func init() {
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.Add)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddPodController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddServiceController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddConfigMapController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddDeploymentController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddSecretController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddDaemonSetController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddIngressController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddJobController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddStatefulSetController)
	AddToManagerFuncs = append(AddToManagerFuncs, resourcebundlestate.AddCsrController)
}
