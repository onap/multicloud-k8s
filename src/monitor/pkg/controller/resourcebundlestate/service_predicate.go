package resourcebundlestate

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type servicePredicate struct {
}

func (s *servicePredicate) Create(evt event.CreateEvent) bool {

	if evt.Meta == nil {
		return false
	}

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}

func (s *servicePredicate) Delete(evt event.DeleteEvent) bool {

	if evt.Meta == nil {
		return false
	}

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}

func (s *servicePredicate) Update(evt event.UpdateEvent) bool {

	if evt.MetaNew == nil {
		return false
	}

	labels := evt.MetaNew.GetLabels()
	return checkLabel(labels)
}

func (s *servicePredicate) Generic(evt event.GenericEvent) bool {

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}
