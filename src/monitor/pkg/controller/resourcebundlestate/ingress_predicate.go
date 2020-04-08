package resourcebundlestate

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type ingressPredicate struct {
}

func (i *ingressPredicate) Create(evt event.CreateEvent) bool {

	if evt.Meta == nil {
		return false
	}

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}

func (i *ingressPredicate) Delete(evt event.DeleteEvent) bool {

	if evt.Meta == nil {
		return false
	}

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}

func (i *ingressPredicate) Update(evt event.UpdateEvent) bool {

	if evt.MetaNew == nil {
		return false
	}

	labels := evt.MetaNew.GetLabels()
	return checkLabel(labels)
}

func (i *ingressPredicate) Generic(evt event.GenericEvent) bool {

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}
