/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package module

import (
	"container/heap"

	"fmt"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	client "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/contextupdateclient"
	rsyncclient "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/installappclient"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
	mtypes "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/types"
	pkgerrors "github.com/pkg/errors"
)

// ControllerTypePlacement denotes "placement" Controller Type
const ControllerTypePlacement string = "placement"

// ControllerTypeAction denotes "action" Controller Type
const ControllerTypeAction string = "action"

// rsyncName denotes the name of the rsync controller
const rsyncName = "rsync"

// ControllerElement consists of controller and an internal field - index
type ControllerElement struct {
	controller controller.Controller
	index      int // used for indexing the HeapArray
}

// PrioritizedControlList contains PrioritizedList of PlacementControllers and ActionControllers
type PrioritizedControlList struct {
	pPlaCont []controller.Controller
	pActCont []controller.Controller
}

// PriorityQueue is the heapArray to store the Controllers
type PriorityQueue []*ControllerElement

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us highest Priority controller
	// The lower the number, higher the priority
	return pq[i].controller.Spec.Priority < pq[j].controller.Spec.Priority
}

// Pop method returns the controller with the highest priority
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	c := old[n-1]
	c.index = -1
	*pq = old[0 : n-1]
	return c
}

// Push method add a controller into the heapArray
func (pq *PriorityQueue) Push(c interface{}) {
	n := len(*pq)
	controllerElement := c.(*ControllerElement)
	controllerElement.index = n
	*pq = append(*pq, controllerElement)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func getPrioritizedControllerList(p, ca, v, di string) (PrioritizedControlList, map[string]string, error) {
	listOfControllers := make([]string, 0) // shall contain the real controllerNames to be passed to controllerAPI
	mapOfControllers := make(map[string]string)

	iList, err := NewIntentClient().GetAllIntents(p, ca, v, di)
	if err != nil {
		return PrioritizedControlList{}, map[string]string{}, err
	}
	for _, eachmap := range iList.ListOfIntents {
		for controller, controllerIntent := range eachmap {
			if controller != GenericPlacementIntentName {
				listOfControllers = append(listOfControllers, controller)
				mapOfControllers[controller] = controllerIntent
			}
		}
	}

	listPC := make([]*ControllerElement, 0)
	listAC := make([]*ControllerElement, 0)

	for _, cn := range listOfControllers {
		c, err := NewClient().Controller.GetController(cn)

		if err != nil {
			return PrioritizedControlList{}, map[string]string{}, err
		}
		if c.Spec.Type == ControllerTypePlacement {
			// Collect in listPC
			listPC = append(listPC, &ControllerElement{controller: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:        c.Metadata.Name,
					Description: c.Metadata.Description,
					UserData1:   c.Metadata.UserData1,
					UserData2:   c.Metadata.UserData2,
				},
				Spec: controller.ControllerSpec{
					Host:     c.Spec.Host,
					Port:     c.Spec.Port,
					Type:     c.Spec.Type,
					Priority: c.Spec.Priority,
				},
			}})
		} else if c.Spec.Type == ControllerTypeAction {
			// Collect in listAC
			listAC = append(listAC, &ControllerElement{controller: controller.Controller{
				Metadata: mtypes.Metadata{
					Name:        c.Metadata.Name,
					Description: c.Metadata.Description,
					UserData1:   c.Metadata.UserData1,
					UserData2:   c.Metadata.UserData2,
				},
				Spec: controller.ControllerSpec{
					Host:     c.Spec.Host,
					Port:     c.Spec.Port,
					Type:     c.Spec.Type,
					Priority: c.Spec.Priority,
				},
			}})
		} else {
			log.Info("Controller type undefined", log.Fields{"Controller type": c.Spec.Type, "ControllerName": c.Metadata.Name})
		}
	}

	pqPlacementCont := make(PriorityQueue, len(listPC))
	for i, eachPC := range listPC {
		pqPlacementCont[i] = &ControllerElement{controller: eachPC.controller, index: i}
	}
	prioritizedPlaControllerList := make([]controller.Controller, 0)
	heap.Init(&pqPlacementCont)
	for pqPlacementCont.Len() > 0 {
		ce := heap.Pop(&pqPlacementCont).(*ControllerElement)

		prioritizedPlaControllerList = append(prioritizedPlaControllerList, ce.controller)
	}

	pqActionCont := make(PriorityQueue, len(listAC))
	for i, eachAC := range listAC {
		pqActionCont[i] = &ControllerElement{controller: eachAC.controller, index: i}
	}
	prioritizedActControllerList := make([]controller.Controller, 0)
	heap.Init(&pqActionCont)
	for pqActionCont.Len() > 0 {
		ce := heap.Pop(&pqActionCont).(*ControllerElement)
		prioritizedActControllerList = append(prioritizedActControllerList, ce.controller)
	}

	prioritizedControlList := PrioritizedControlList{pPlaCont: prioritizedPlaControllerList, pActCont: prioritizedActControllerList}

	return prioritizedControlList, mapOfControllers, nil

}

/*
callGrpcForControllerList method shall take in a list of controllers, a map of contollers to controllerIntentNames and contextID. It invokes the context
updation through the grpc client for the given list of controllers.
*/
func callGrpcForControllerList(cl []controller.Controller, mc map[string]string, contextid interface{}) error {
	for _, c := range cl {
		controller := c.Metadata.Name
		controllerIntentName := mc[controller]
		appContextID := fmt.Sprintf("%v", contextid)
		err := client.InvokeContextUpdate(controller, controllerIntentName, appContextID)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller
and then sets the RsyncInfo global variable.
*/
func queryDBAndSetRsyncInfo() (rsyncclient.RsyncInfo, error) {
	client := controller.NewControllerClient()
	vals, _ := client.GetControllers()
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfo := rsyncclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return rsyncclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

/*
callRsyncInstall method shall take in the app context id and invokes the rsync service via grpc
*/
func callRsyncInstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = rsyncclient.InvokeInstallApp(appContextID)
	if err != nil {
		return err
	}
	return nil
}

/*
callRsyncUninstall method shall take in the app context id and invokes the rsync service via grpc
*/
func callRsyncUninstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = rsyncclient.InvokeUninstallApp(appContextID)
	if err != nil {
		return err
	}
	return nil
}

/*
deleteExtraClusters method shall delete the extra cluster handles for each AnyOf cluster present in the etcd after the grpc call for context updation.
*/
func deleteExtraClusters(apps []App, ct appcontext.AppContext) error {
	for _, app := range apps {
		an := app.Metadata.Name
		gmap, err := ct.GetClusterGroupMap(an)
		if err != nil {
			return err
		}
		for gr, cl := range gmap {
			for i, cn := range cl {
				// avoids deleting the first cluster
				if i > 0 {
					ch, err := ct.GetClusterHandle(an, cn)
					if err != nil {
						return err
					}
					err = ct.DeleteCluster(ch)
					if err != nil {
						return err
					}
					log.Info("::Deleted cluster for::", log.Fields{"appName": an, "GroupNumber": gr, "ClusterName": cn})
				}
			}

		}
	}
	return nil
}
