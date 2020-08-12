package resourcebundlestate

import (
	"context"
	"log"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	v1beta1 "k8s.io/api/extensions/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddIngressController the new controller to the controller manager
func AddIngressController(mgr manager.Manager) error {
	return addIngressController(mgr, newIngressReconciler(mgr))
}

func addIngressController(mgr manager.Manager, r *ingressReconciler) error {
	// Create a new controller
	c, err := controller.New("Ingress-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Ingress
	// Predicate filters Ingress which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &v1beta1.Ingress{}}, &handler.EnqueueRequestForObject{}, &ingressPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newIngressReconciler(m manager.Manager) *ingressReconciler {
	return &ingressReconciler{client: m.GetClient()}
}

type ingressReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the ingress we watch.
func (r *ingressReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Ingress: %+v\n", req)

	ing := &v1beta1.Ingress{}
	err := r.client.Get(context.TODO(), req.NamespacedName, ing)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Ingress not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Ingress's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Ingress has been deleted.
			r.deleteIngressFromAllCRs(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get ingress: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this Ingress via the labelselector
	crSelector := returnLabel(ing.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Ingress")
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

	err = r.updateCRs(rbStatusList, ing)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteIngressFromAllCRs deletes ingress status from all the CRs when the Ingress itself has been deleted
// and we have not handled the updateCRs yet.
// Since, we don't have the Ingress's labels, we need to look at all the CRs in this namespace
func (r *ingressReconciler) deleteIngressFromAllCRs(namespacedName types.NamespacedName) error {

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

func (r *ingressReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, ing *v1beta1.Ingress) error {

	for _, cr := range crl.Items {
		// Ingress is not scheduled for deletion
		if ing.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, ing)
			if err != nil {
				return err
			}
		} else {
			// Ingress is scheduled for deletion
			r.deleteFromSingleCR(&cr, ing.Name)
		}
	}

	return nil
}

func (r *ingressReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, name string) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.IngressStatuses)
	for i, rstatus := range cr.Status.IngressStatuses {
		if rstatus.Name == name {
			//Delete that status from the array
			cr.Status.IngressStatuses[i] = cr.Status.IngressStatuses[length-1]
			cr.Status.IngressStatuses[length-1].Status = v1beta1.IngressStatus{}
			cr.Status.IngressStatuses = cr.Status.IngressStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for Ingress in CR")
	return nil
}

func (r *ingressReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, ing *v1beta1.Ingress) error {

	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.IngressStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == ing.Name {
			ing.Status.DeepCopyInto(&cr.Status.IngressStatuses[i].Status)
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
	cr.Status.IngressStatuses = append(cr.Status.IngressStatuses, v1beta1.Ingress{
		TypeMeta:   ing.TypeMeta,
		ObjectMeta: ing.ObjectMeta,
		Status:     ing.Status,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
