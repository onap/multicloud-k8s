package resourcebundlestate

import (
	"context"
	"log"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddDaemonSetController the new controller to the controller manager
func AddDaemonSetController(mgr manager.Manager) error {
	return addDaemonSetController(mgr, newDaemonSetReconciler(mgr))
}

func addDaemonSetController(mgr manager.Manager, r *daemonSetReconciler) error {
	// Create a new controller
	c, err := controller.New("Daemonset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource DaemonSets
	// Predicate filters DaemonSets which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForObject{}, &daemonSetPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newDaemonSetReconciler(m manager.Manager) *daemonSetReconciler {
	return &daemonSetReconciler{client: m.GetClient()}
}

type daemonSetReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the daemonSets we watch.
func (r *daemonSetReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for DaemonSet: %+v\n", req)

	ds := &appsv1.DaemonSet{}
	err := r.client.Get(context.TODO(), req.NamespacedName, ds)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("DaemonSet not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the DaemonSet's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the DaemonSet has been deleted.
			r.deleteDaemonSetFromAllCRs(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get daemonSet: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this daemonSet via the labelselector
	crSelector := returnLabel(ds.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this DaemonSet")
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

	err = r.updateCRs(rbStatusList, ds)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteDaemonSetFromAllCRs deletes daemonSet status from all the CRs when the DaemonSet itself has been deleted
// and we have not handled the updateCRs yet.
// Since, we don't have the daemonSet's labels, we need to look at all the CRs in this namespace
func (r *daemonSetReconciler) deleteDaemonSetFromAllCRs(namespacedName types.NamespacedName) error {

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

func (r *daemonSetReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, ds *appsv1.DaemonSet) error {

	for _, cr := range crl.Items {
		// DaemonSet is not scheduled for deletion
		if ds.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, ds)
			if err != nil {
				return err
			}
		} else {
			// DaemonSet is scheduled for deletion
			r.deleteFromSingleCR(&cr, ds.Name)
		}
	}

	return nil
}

func (r *daemonSetReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, name string) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.DaemonSetStatuses)
	for i, rstatus := range cr.Status.DaemonSetStatuses {
		if rstatus.Name == name {
			//Delete that status from the array
			cr.Status.DaemonSetStatuses[i] = cr.Status.DaemonSetStatuses[length-1]
			cr.Status.DaemonSetStatuses[length-1].Status = appsv1.DaemonSetStatus{}
			cr.Status.DaemonSetStatuses = cr.Status.DaemonSetStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for DaemonSet in CR")
	return nil
}

func (r *daemonSetReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, ds *appsv1.DaemonSet) error {

	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.DaemonSetStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == ds.Name {
			ds.Status.DeepCopyInto(&cr.Status.DaemonSetStatuses[i].Status)
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
	cr.Status.DaemonSetStatuses = append(cr.Status.DaemonSetStatuses, appsv1.DaemonSet{
		TypeMeta:   ds.TypeMeta,
		ObjectMeta: ds.ObjectMeta,
		Status:     ds.Status,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
