package resourcebundlestate

import (
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

// EventHandler adds some specific handling for certain types of events
// related to the ResourceBundleState CR.
type EventHandler struct {
	handler.EnqueueRequestForObject
}

// Delete ignores any delete operations on a ResourceBundleState CR
func (p *EventHandler) Delete(evt event.DeleteEvent, q workqueue.RateLimitingInterface) {
	return
}

// Update ignores any update operations on a ResourceBundleState CR
func (p *EventHandler) Update(evt event.UpdateEvent, q workqueue.RateLimitingInterface) {
	return
}

// Generic ignores any generic operations on a ResourceBundleState CR
func (p *EventHandler) Generic(evt event.GenericEvent, q workqueue.RateLimitingInterface) {
	return
}
