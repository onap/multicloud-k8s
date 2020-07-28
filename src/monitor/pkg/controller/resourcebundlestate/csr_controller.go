package resourcebundlestate

import (
	"context"
	"log"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	certsapi "k8s.io/api/certificates/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddCsrController the new controller to the controller manager
func AddCsrController(mgr manager.Manager) error {
	return addCsrController(mgr, newCsrReconciler(mgr))
}

func addCsrController(mgr manager.Manager, r *csrReconciler) error {
	// Create a new controller
	c, err := controller.New("Csr-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Csrs
	// Predicate filters csrs which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &certsapi.CertificateSigningRequest{}}, &handler.EnqueueRequestForObject{}, &csrPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newCsrReconciler(m manager.Manager) *csrReconciler {
	return &csrReconciler{client: m.GetClient()}
}

type csrReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the csrs we watch.
func (r *csrReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Csr: %+v\n", req)

	csr := &certsapi.CertificateSigningRequest{}
	err := r.client.Get(context.TODO(), req.NamespacedName, csr)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Csr not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Csr's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the POD has been deleted.
			r.deleteCsrFromAllCRs(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get csr: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this csr via the labelselector
	crSelector := returnLabel(csr.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Csr")
	}

	// Get the CRs which have this label and update them all
	// Ideally, we will have only one CR, but there is nothing
	// preventing the creation of multiple.
	// TODO: Consider using an admission validating webook to prevent multiple
	rbStatusList := &v1alpha1.ResourceBundleStateList{}
	err = listClusterResources(r.client, crSelector, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return reconcile.Result{}, nil
	}

	err = r.updateCRs(rbStatusList, csr)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteCsrFromAllCRs deletes csr status from all the CRs when the POD itself has been deleted
// and we have not handled the updateCRs yet.
// Since, we don't have the csr's labels, we need to look at all the CRs in this namespace
func (r *csrReconciler) deleteCsrFromAllCRs(namespacedName types.NamespacedName) error {

	rbStatusList := &v1alpha1.ResourceBundleStateList{}
	err := listClusterResources(r.client, nil, rbStatusList)
	if err != nil || len(rbStatusList.Items) == 0 {
		log.Printf("Did not find any CRs tracking this resource\n")
		return nil
	}
	for _, cr := range rbStatusList.Items {
		r.deleteFromSingleCR(&cr, namespacedName.Name)
	}

	return nil
}

func (r *csrReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, csr *certsapi.CertificateSigningRequest) error {

	for _, cr := range crl.Items {
		// Csr is not scheduled for deletion
		if csr.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, csr)
			if err != nil {
				return err
			}
		} else {
			// Csr is scheduled for deletion
			r.deleteFromSingleCR(&cr, csr.Name)
		}
	}

	return nil
}

func (r *csrReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, name string) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.CsrStatuses)
	for i, rstatus := range cr.Status.CsrStatuses {
		if rstatus.Name == name {
			//Delete that status from the array
			cr.Status.CsrStatuses[i] = cr.Status.CsrStatuses[length-1]
			cr.Status.CsrStatuses[length-1] = certsapi.CertificateSigningRequest{}
			cr.Status.CsrStatuses = cr.Status.CsrStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for POD in CR")
	return nil
}

func (r *csrReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, csr *certsapi.CertificateSigningRequest) error {

	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.CsrStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == csr.Name {
			csr.Status.DeepCopyInto(&cr.Status.CsrStatuses[i].Status)
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
	cr.Status.CsrStatuses = append(cr.Status.CsrStatuses, certsapi.CertificateSigningRequest{
		TypeMeta:   csr.TypeMeta,
		ObjectMeta: csr.ObjectMeta,
		Status:     csr.Status,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
