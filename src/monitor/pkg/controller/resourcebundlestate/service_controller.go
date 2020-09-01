package resourcebundlestate

import (
	"context"
	"log"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddServiceController the new controller to the controller manager
func AddServiceController(mgr manager.Manager) error {
	return addServiceController(mgr, newServiceReconciler(mgr))
}

func addServiceController(mgr manager.Manager, r *serviceReconciler) error {
	// Create a new controller
	c, err := controller.New("Service-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Services
	// Predicate filters Service which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForObject{}, &servicePredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newServiceReconciler(m manager.Manager) *serviceReconciler {
	return &serviceReconciler{client: m.GetClient()}
}

type serviceReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the services we watch.
func (r *serviceReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Service: %+v\n", req)

	svc := &corev1.Service{}
	err := r.client.Get(context.TODO(), req.NamespacedName, svc)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Service not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Service's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Service has been deleted.
			r.deleteServiceFromAllCRs(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get service: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this service via the labelselector
	crSelector := returnLabel(svc.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Service")
	}

	// Get the CRs which have this label and update them all
	// Ideally, we will have only one CR, but there is nothing
	// preventing the creation of multiple.
	// TODO: Consider using an admission validating webook to prevent multiple
	rbStatusList := &v1alpha1.ResourceBundleStateList{}
	err = listResources(r.client, req.Namespace, crSelector, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return reconcile.Result{}, nil
	}

	err = r.updateCRs(rbStatusList, svc)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteServiceFromAllCRs deletes service status from all the CRs when the Service itself has been deleted
// and we have not handled the updateCRs yet.
// Since, we don't have the service's labels, we need to look at all the CRs in this namespace
func (r *serviceReconciler) deleteServiceFromAllCRs(namespacedName types.NamespacedName) error {

	rbStatusList := &v1alpha1.ResourceBundleStateList{}
	err := listResources(r.client, namespacedName.Namespace, nil, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return nil
	}
	for _, cr := range rbStatusList.Items {
		r.deleteFromSingleCR(&cr, namespacedName.Name)
	}

	return nil
}

func (r *serviceReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, svc *corev1.Service) error {

	for _, cr := range crl.Items {
		// Service is not scheduled for deletion
		if svc.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, svc)
			if err != nil {
				return err
			}
		} else {
			// Service is scheduled for deletion
			r.deleteFromSingleCR(&cr, svc.Name)
		}
	}

	return nil
}

func (r *serviceReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, name string) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.ServiceStatuses)
	for i, rstatus := range cr.Status.ServiceStatuses {
		if rstatus.Name == name {
			//Delete that status from the array
			cr.Status.ServiceStatuses[i] = cr.Status.ServiceStatuses[length-1]
			cr.Status.ServiceStatuses[length-1].Status = corev1.ServiceStatus{}
			cr.Status.ServiceStatuses = cr.Status.ServiceStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for Service in CR")
	return nil
}

func (r *serviceReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, svc *corev1.Service) error {

	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.ServiceStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == svc.Name {
			svc.Status.DeepCopyInto(&cr.Status.ServiceStatuses[i].Status)
			err := r.client.Status().Update(context.TODO(), cr)
			if err != nil {
				log.Printf("failed to update rbstate: %v\n", err)
				return err
			}
			return nil
		}
	}

	// Exited for loop with no status found
	// Increment the number of tracked resources
	cr.Status.ResourceCount++

	// Add it to CR
	cr.Status.ServiceStatuses = append(cr.Status.ServiceStatuses, corev1.Service{
		TypeMeta:   svc.TypeMeta,
		ObjectMeta: svc.ObjectMeta,
		Status:     svc.Status,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
