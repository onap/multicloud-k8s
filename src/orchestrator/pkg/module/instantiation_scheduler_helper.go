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
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"container/heap"
	mtypes "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/types"
)

// ContollerTypePlacement denotes Placement Controller Type
const ContollerTypePlacement string = "placement"
// ContollerTypeAction denotes action Controller Type
const ContollerTypeAction string = "action"

// Controller Struct has metaData and Spec of Controller, additionally an internal field Index, which is neccessary for PQ
type Controller struct {
	Metadata mtypes.Metadata `json:"metadata"`
	Spec     ControllerSpec  `json:"spec"`
	Index    int    // this is required for the index in the heapArray
}

// ControllerSpec consists of Host, Port, Type and Priority
type ControllerSpec struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Type     string `json:"type"`
	Priority int    `json:"priority"`
}

// PrioritisedControlList contains PrioritisedList of PlacementControllers and ActionControllers
type PrioritisedControlList struct {
	pPlaCont []*Controller
	pActCont []*Controller
}

// PriorityQueue is the heapArray to store the Controllers
type PriorityQueue []*Controller

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us highest Priority controller
	// The lower the number, higher the priority
	return pq[i].Spec.Priority < pq[j].Spec.Priority
}

// Pop method returns the controller with the highest priority
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	c := old[n-1]
	c.Index = -1
	*pq = old[0 : n-1]
	return c
}

// Push method add a controller into the heapArray
func (pq *PriorityQueue) Push(c interface{}) {
	n := len(*pq)
	controller := c.(*Controller)
	controller.Index = n
	*pq = append(*pq, controller)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func getPrioritisedControllerList(p, ca, v, di string) (PrioritisedControlList, error) {
	listOfIntentTypes := make([]string, 0)
	listOfIntentNames := make([]string, 0) // shall contain the real controllerNames to be passed to controllerAPI
	iList, err := NewIntentClient().GetAllIntents(p, ca, v, di)
	if err != nil {
		return PrioritisedControlList{}, err
	}
	for _, eachmap := range iList.ListOfIntents {
		for name, value := range eachmap {
			listOfIntentTypes = append(listOfIntentTypes, name)
			listOfIntentNames = append(listOfIntentNames, value)
		}
	}

	listPC := make([]*Controller, 0)
	listAC := make([]*Controller, 0)

	for _, cn := range listOfIntentNames {
		c, err := NewClient().Controller.GetController(cn)

		if err != nil {
			return PrioritisedControlList{}, err
		}
		if c.Spec.Type == ContollerTypePlacement {
			// Collect in listPC
			listPC = append(listPC, &Controller{
				Metadata: mtypes.Metadata{
					Name: c.Metadata.Name,
					Description: c.Metadata.Description,
					UserData1: c.Metadata.UserData1,
					UserData2: c.Metadata.UserData2,
				},
				Spec: ControllerSpec {
					Host:     c.Spec.Host,
					Port:     c.Spec.Port,
					Type:     c.Spec.Type,
					Priority: c.Spec.Priority,
				},
				})
			//indexPC++
			// push into placementPQ
		} else if c.Spec.Type == ContollerTypeAction{
			// Collect in listAC
			listAC = append(listAC, &Controller{
				Metadata: mtypes.Metadata{
					Name: c.Metadata.Name,
					Description: c.Metadata.Description,
					UserData1: c.Metadata.UserData1,
					UserData2: c.Metadata.UserData2,
				},
				Spec: ControllerSpec {
					Host:     c.Spec.Host,
					Port:     c.Spec.Port,
					Type:     c.Spec.Type,
					Priority: c.Spec.Priority,
				},
				})
		} else {
			log.Info("Controller type undefined", log.Fields{"Controller type":c.Spec.Type, "ControllerName":c.Metadata.Name})
		}
	}

	pqPlacementCont := make(PriorityQueue, len(listPC))
	for i, eachPC := range listPC {
		pqPlacementCont[i] = eachPC
		pqPlacementCont[i].Index = i
	}
	prioritisedPlaControllerList := make([]*Controller, 0)
	heap.Init(&pqPlacementCont)
	for pqPlacementCont.Len() > 0 {
		cs := heap.Pop(&pqPlacementCont).(*Controller)
		prioritisedPlaControllerList = append(prioritisedPlaControllerList, cs)
	}

	pqActionCont := make(PriorityQueue, len(listAC))
	for i, eachAC := range listAC {
		pqActionCont[i] = eachAC
		pqActionCont[i].Index = i
	}
	prioritisedActControllerList := make([]*Controller, 0)
	heap.Init(&pqActionCont)
	for pqActionCont.Len() > 0 {
		cs := heap.Pop(&pqActionCont).(*Controller)
		prioritisedActControllerList = append(prioritisedActControllerList, cs)
	}

	prioritisedControlList := PrioritisedControlList{pPlaCont: prioritisedPlaControllerList, pActCont: prioritisedActControllerList}

	return prioritisedControlList, nil

}
