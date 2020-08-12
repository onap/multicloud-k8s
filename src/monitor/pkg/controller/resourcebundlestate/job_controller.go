package resourcebundlestate

import (
	"context"
	"log"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	v1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddJobController the new controller to the controller manager
func AddJobController(mgr manager.Manager) error {
	return addJobController(mgr, newJobReconciler(mgr))
}

func addJobController(mgr manager.Manager, r *jobReconciler) error {
	// Create a new controller
	c, err := controller.New("Job-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Jobs
	// Predicate filters Job which don't have the k8splugin label
	err = c.Watch(&source.Kind{Type: &v1.Job{}}, &handler.EnqueueRequestForObject{}, &jobPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newJobReconciler(m manager.Manager) *jobReconciler {
	return &jobReconciler{client: m.GetClient()}
}

type jobReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the jobs we watch.
func (r *jobReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Job: %+v\n", req)

	job := &v1.Job{}
	err := r.client.Get(context.TODO(), req.NamespacedName, job)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Job not found: %+v. Remove from CR if it is stored there.\n", req.NamespacedName)
			// Remove the Job's status from StatusList
			// This can happen if we get the DeletionTimeStamp event
			// after the Job has been deleted.
			r.deleteJobFromAllCRs(req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get Job: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this Job via the labelselector
	crSelector := returnLabel(job.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Job")
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

	err = r.updateCRs(rbStatusList, job)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// deleteJobFromAllCRs deletes job status from all the CRs when the Job itself has been deleted
// and we have not handled the updateCRs yet.
// Since, we don't have the job's labels, we need to look at all the CRs in this namespace
func (r *jobReconciler) deleteJobFromAllCRs(namespacedName types.NamespacedName) error {

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

func (r *jobReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, job *v1.Job) error {

	for _, cr := range crl.Items {
		// Job is not scheduled for deletion
		if job.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, job)
			if err != nil {
				return err
			}
		} else {
			// Job is scheduled for deletion
			r.deleteFromSingleCR(&cr, job.Name)
		}
	}

	return nil
}

func (r *jobReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, name string) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.JobStatuses)
	for i, rstatus := range cr.Status.JobStatuses {
		if rstatus.Name == name {
			//Delete that status from the array
			cr.Status.JobStatuses[i] = cr.Status.JobStatuses[length-1]
			cr.Status.JobStatuses[length-1].Status = v1.JobStatus{}
			cr.Status.JobStatuses = cr.Status.JobStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for Job in CR")
	return nil
}

func (r *jobReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, job *v1.Job) error {

	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.JobStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == job.Name {
			job.Status.DeepCopyInto(&cr.Status.JobStatuses[i].Status)
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
	cr.Status.JobStatuses = append(cr.Status.JobStatuses, v1.Job{
		TypeMeta:   job.TypeMeta,
		ObjectMeta: job.ObjectMeta,
		Status:     job.Status,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
