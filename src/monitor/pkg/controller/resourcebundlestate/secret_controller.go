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

// AddSecretController the new controller to the controller manager
func AddSecretController(mgr manager.Manager) error {
	return addSecretController(mgr, newSecretReconciler(mgr))
}

func addSecretController(mgr manager.Manager, r *secretReconciler) error {
	// Create a new controller
	c, err := controller.New("Secret-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Secret
	// Predicate filters Secret which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{}, &secretPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newSecretReconciler(m manager.Manager) *secretReconciler {
	return &secretReconciler{client: m.GetClient()}
}

type secretReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the Secrets we watch.
func (r *secretReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Secret: %+v\n", req)

	sec := &corev1.Secret{}
	err := r.client.Get(context.TODO(), req.NamespacedName, sec)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Secret not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Secret's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Secret has been deleted.
			r.deleteSecretFromAllCRs(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get Secret: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this Secret via the labelselector
	crSelector := returnLabel(sec.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Secret")
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

	err = r.updateCRs(rbStatusList, sec)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteSecretFromAllCRs deletes Secret status from all the CRs when the Secret itself has been deleted
// and we have not handled the updateCRs yet.
// Since, we don't have the Secret's labels, we need to look at all the CRs in this namespace
func (r *secretReconciler) deleteSecretFromAllCRs(namespacedName types.NamespacedName) error {

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

func (r *secretReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, sec *corev1.Secret) error {

	for _, cr := range crl.Items {
		// Secret is not scheduled for deletion
		if sec.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, sec)
			if err != nil {
				return err
			}
		} else {
			// Secret is scheduled for deletion
			r.deleteFromSingleCR(&cr, sec.Name)
		}
	}

	return nil
}

func (r *secretReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, name string) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.SecretStatuses)
	for i, rstatus := range cr.Status.SecretStatuses {
		if rstatus.Name == name {
			//Delete that status from the array
			cr.Status.SecretStatuses[i] = cr.Status.SecretStatuses[length-1]
			cr.Status.SecretStatuses = cr.Status.SecretStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for SecretStatuses in CR")
	return nil
}

func (r *secretReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, sec *corev1.Secret) error {

	// Update status after searching for it in the list of resourceStatuses
	for _, rstatus := range cr.Status.SecretStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == sec.Name {
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
	cr.Status.SecretStatuses = append(cr.Status.SecretStatuses, corev1.Secret{
		TypeMeta:   sec.TypeMeta,
		ObjectMeta: sec.ObjectMeta,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
