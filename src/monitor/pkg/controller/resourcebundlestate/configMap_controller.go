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

// AddConfigMapController the new controller to the controller manager
func AddConfigMapController(mgr manager.Manager) error {
	return addConfigMapController(mgr, newConfigMapReconciler(mgr))
}

func addConfigMapController(mgr manager.Manager, r *configMapReconciler) error {
	// Create a new controller
	c, err := controller.New("ConfigMap-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource ConfigMaps
	// Predicate filters Service which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &corev1.ConfigMap{}}, &handler.EnqueueRequestForObject{}, &configMapPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newConfigMapReconciler(m manager.Manager) *configMapReconciler {
	return &configMapReconciler{client: m.GetClient()}
}

type configMapReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the ConfigMaps we watch.
func (r *configMapReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for ConfigMap: %+v\n", req)

	cm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), req.NamespacedName, cm)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("ConfigMap not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the ConfigMap's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the ConfigMap has been deleted.
			r.deleteConfigMapFromAllCRs(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get ConfigMap: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this ConfigMap via the labelselector
	crSelector := returnLabel(cm.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this ConfigMap")
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

	err = r.updateCRs(rbStatusList, cm)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteConfigMapFromAllCRs deletes ConfigMap status from all the CRs when the ConfigMap itself has been deleted
// and we have not handled the updateCRs yet.
// Since, we don't have the ConfigMap's labels, we need to look at all the CRs in this namespace
func (r *configMapReconciler) deleteConfigMapFromAllCRs(namespacedName types.NamespacedName) error {

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

func (r *configMapReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, cm *corev1.ConfigMap) error {

	for _, cr := range crl.Items {
		// ConfigMap is not scheduled for deletion
		if cm.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, cm)
			if err != nil {
				return err
			}
		} else {
			// ConfigMap is scheduled for deletion
			r.deleteFromSingleCR(&cr, cm.Name)
		}
	}

	return nil
}

func (r *configMapReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, name string) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.ConfigMapStatuses)
	for i, rstatus := range cr.Status.ConfigMapStatuses {
		if rstatus.Name == name {
			//Delete that status from the array
			cr.Status.ConfigMapStatuses[i] = cr.Status.ConfigMapStatuses[length-1]
			cr.Status.ConfigMapStatuses = cr.Status.ConfigMapStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for ConfigMapStatuses in CR")
	return nil
}

func (r *configMapReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, cm *corev1.ConfigMap) error {

	// Update status after searching for it in the list of resourceStatuses
	for _, rstatus := range cr.Status.ConfigMapStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == cm.Name {
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
	cr.Status.ConfigMapStatuses = append(cr.Status.ConfigMapStatuses, corev1.ConfigMap{
		TypeMeta:   cm.TypeMeta,
		ObjectMeta: cm.ObjectMeta,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
