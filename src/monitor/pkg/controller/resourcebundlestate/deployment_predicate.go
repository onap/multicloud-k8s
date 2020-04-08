package resourcebundlestate

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type deploymentPredicate struct {
}

func (d *deploymentPredicate) Create(evt event.CreateEvent) bool {

	if evt.Meta == nil {
		return false
	}

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}

func (d *deploymentPredicate) Delete(evt event.DeleteEvent) bool {

	if evt.Meta == nil {
		return false
	}

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}

func (d *deploymentPredicate) Update(evt event.UpdateEvent) bool {

	if evt.MetaNew == nil {
		return false
	}

	labels := evt.MetaNew.GetLabels()
	return checkLabel(labels)
}

func (d *deploymentPredicate) Generic(evt event.GenericEvent) bool {

	labels := evt.Meta.GetLabels()
	return checkLabel(labels)
}
