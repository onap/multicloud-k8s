/*
 * Copyright 2026 Deutsche Telekom AG
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

package resourcebundlestate

import (
	"testing"

	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

// TestEventHandler_MutatingEventsAreIgnored verifies the primary CR's
// EventHandler drops Delete/Update/Generic events (only creates are acted on),
// i.e. nothing is enqueued for those events.
func TestEventHandler_MutatingEventsAreIgnored(t *testing.T) {
	h := &EventHandler{}
	q := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	defer q.ShutDown()

	h.Delete(event.DeleteEvent{}, q)
	h.Update(event.UpdateEvent{}, q)
	h.Generic(event.GenericEvent{}, q)

	if q.Len() != 0 {
		t.Errorf("expected no items enqueued for Delete/Update/Generic, got %d", q.Len())
	}
}
