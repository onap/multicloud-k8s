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
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
)

// newControllerElement builds a ControllerElement with the given priority.
func newControllerElement(name string, priority int) *ControllerElement {
	return &ControllerElement{
		controller: controller.Controller{
			Spec: controller.ControllerSpec{Priority: priority},
		},
	}
}

func TestPriorityQueueLenLessSwap(t *testing.T) {
	pq := PriorityQueue{
		newControllerElement("c1", 5),
		newControllerElement("c2", 1),
	}

	if pq.Len() != 2 {
		t.Fatalf("Len returned %d; expected 2", pq.Len())
	}

	// Lower priority number => higher priority => Less is true.
	if !pq.Less(1, 0) {
		t.Fatalf("Less(1,0) expected true (priority 1 < priority 5)")
	}
	if pq.Less(0, 1) {
		t.Fatalf("Less(0,1) expected false (priority 5 !< priority 1)")
	}

	pq.Swap(0, 1)
	if pq[0].controller.Spec.Priority != 1 || pq[1].controller.Spec.Priority != 5 {
		t.Fatalf("Swap did not swap the elements: %v", pq)
	}
	if pq[0].index != 0 || pq[1].index != 1 {
		t.Fatalf("Swap did not update indices: %v", pq)
	}
}

func TestPriorityQueuePushPop(t *testing.T) {
	pq := make(PriorityQueue, 0)

	pq.Push(newControllerElement("c1", 5))
	pq.Push(newControllerElement("c2", 1))
	pq.Push(newControllerElement("c3", 3))

	if pq.Len() != 3 {
		t.Fatalf("after Push, Len returned %d; expected 3", pq.Len())
	}
	// Push should set the index of each element.
	if pq[2].index != 2 {
		t.Fatalf("Push did not set index on the last element: %d", pq[2].index)
	}

	popped := pq.Pop().(*ControllerElement)
	if popped.controller.Spec.Priority != 3 {
		t.Fatalf("Pop returned unexpected element (priority %d); expected the last pushed (3)",
			popped.controller.Spec.Priority)
	}
	if popped.index != -1 {
		t.Fatalf("Pop did not reset the popped element's index to -1: %d", popped.index)
	}
	if pq.Len() != 2 {
		t.Fatalf("after Pop, Len returned %d; expected 2", pq.Len())
	}
}

// TestPriorityQueueHeapOrder verifies the heap yields elements in ascending
// priority-number order (highest logical priority first) when driven via the
// container/heap package.
func TestPriorityQueueHeapOrder(t *testing.T) {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)

	heap.Push(&pq, newControllerElement("c1", 5))
	heap.Push(&pq, newControllerElement("c2", 1))
	heap.Push(&pq, newControllerElement("c3", 3))

	var order []int
	for pq.Len() > 0 {
		ce := heap.Pop(&pq).(*ControllerElement)
		order = append(order, ce.controller.Spec.Priority)
	}

	expected := []int{1, 3, 5}
	if len(order) != len(expected) {
		t.Fatalf("heap yielded %d elements; expected %d", len(order), len(expected))
	}
	for i := range expected {
		if order[i] != expected[i] {
			t.Fatalf("heap order = %v; expected %v", order, expected)
		}
	}
}
