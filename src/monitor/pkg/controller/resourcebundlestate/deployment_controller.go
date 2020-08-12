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

// AddDeploymentController the new controller to the controller manager
func AddDeploymentController(mgr manager.Manager) error {
	return addDeploymentController(mgr, newDeploymentReconciler(mgr))
}

func addDeploymentController(mgr manager.Manager, r *deploymentReconciler) error {
	// Create a new controller
	c, err := controller.New("Deployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Deployments
	// Predicate filters Deployment which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}, &deploymentPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newDeploymentReconciler(m manager.Manager) *deploymentReconciler {
	return &deploymentReconciler{client: m.GetClient()}
}

type deploymentReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the deployments we watch.
func (r *deploymentReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Deployment: %+v\n", req)

	dep := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), req.NamespacedName, dep)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Deployment not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Deployment's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Deployment has been deleted.
			r.deleteDeploymentFromAllCRs(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get deployment: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this deployment via the labelselector
	crSelector := returnLabel(dep.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Deployment")
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

	err = r.updateCRs(rbStatusList, dep)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteDeploymentFromAllCRs deletes deployment status from all the CRs when the Deployment itself has been deleted
// and we have not handled the updateCRs yet.
// Since, we don't have the deployment's labels, we need to look at all the CRs in this namespace
func (r *deploymentReconciler) deleteDeploymentFromAllCRs(namespacedName types.NamespacedName) error {

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

func (r *deploymentReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, dep *appsv1.Deployment) error {

	for _, cr := range crl.Items {
		// Deployment is not scheduled for deletion
		if dep.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, dep)
			if err != nil {
				return err
			}
		} else {
			// Deployment is scheduled for deletion
			r.deleteFromSingleCR(&cr, dep.Name)
		}
	}

	return nil
}

func (r *deploymentReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, name string) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.DeploymentStatuses)
	for i, rstatus := range cr.Status.DeploymentStatuses {
		if rstatus.Name == name {
			//Delete that status from the array
			cr.Status.DeploymentStatuses[i] = cr.Status.DeploymentStatuses[length-1]
			cr.Status.DeploymentStatuses[length-1].Status = appsv1.DeploymentStatus{}
			cr.Status.DeploymentStatuses = cr.Status.DeploymentStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for Deployment in CR")
	return nil
}

func (r *deploymentReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, dep *appsv1.Deployment) error {

	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.DeploymentStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == dep.Name {
			dep.Status.DeepCopyInto(&cr.Status.DeploymentStatuses[i].Status)
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
	cr.Status.DeploymentStatuses = append(cr.Status.DeploymentStatuses, appsv1.Deployment{
		TypeMeta:   dep.TypeMeta,
		ObjectMeta: dep.ObjectMeta,
		Status:     dep.Status,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
