package resourcebundlestate

import (
	"context"
	"log"

	"monitor/pkg/apis/k8splugin/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// AddPodController the new controller to the controller manager
func AddPodController(mgr manager.Manager) error {
	return addPodController(mgr, newPodReconciler(mgr))
}

func addPodController(mgr manager.Manager, r *podReconciler) error {
	// Create a new controller
	c, err := controller.New("ResourceBundleState-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to secondar resource Pods
	// TODO: See if we can use predicates to filter pods
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForObject{}, &podPredicate{})
	if err != nil {
		return err
	}

	return nil
}

func newPodReconciler(m manager.Manager) *podReconciler {
	return &podReconciler{client: m.GetClient()}
}

type podReconciler struct {
	client client.Client
}

// Reconcile implements the loop that will update the ResourceBundleState CR
// whenever we get any updates from all the pods we watch.
func (r *podReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Updating ResourceBundleState for Pod: %+v\n", req)

	pod := &corev1.Pod{}
	err := r.client.Get(context.TODO(), req.NamespacedName, pod)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Pod not found: %+v. Ignore as it must have been deleted.\n", req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get pod: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	// Find the CRs which track this pod via the labelselector
	crSelector := returnLabel(pod.GetLabels())
	if crSelector == nil {
		log.Println("We should not be here. The predicate should have filtered this Pod")
	}

	listOptions := &client.ListOptions{
		Namespace:     pod.Namespace,
		LabelSelector: labels.SelectorFromSet(crSelector),
	}

	// Get the CRs which have this label and update them all
	// Ideally, we will have only one CR, but there is nothing
	// preventing the creation of multiple.
	// TODO: Consider using an admission validating webook to prevent multiple
	rbStatusList := &v1alpha1.ResourceBundleStateList{}
	err = r.client.List(context.TODO(), listOptions, rbStatusList)
	if err != nil {
		log.Printf("Failed to list CRs: %v", err)
		return reconcile.Result{}, err
	}

	if len(rbStatusList.Items) == 0 {
		log.Printf("No CRs to track this resource created yet.\n")
		return reconcile.Result{}, nil
	}

	err = r.updateCRs(rbStatusList, pod)
	if err != nil {
		// Requeue the update
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *podReconciler) updateCRs(crl *v1alpha1.ResourceBundleStateList, pod *corev1.Pod) error {

	for _, cr := range crl.Items {
		// Pod is not scheduled for deletion
		if pod.DeletionTimestamp == nil {
			err := r.updateSingleCR(&cr, pod)
			if err != nil {
				return err
			}
		} else {
			// Pod is scheduled for deletion
			r.deleteFromSingleCR(&cr, pod)
		}
	}

	return nil
}

func (r *podReconciler) deleteFromSingleCR(cr *v1alpha1.ResourceBundleState, pod *corev1.Pod) error {
	cr.Status.ResourceCount--
	length := len(cr.Status.PodStatuses)
	for i, rstatus := range cr.Status.PodStatuses {
		if rstatus.Name == pod.Name {
			//Delete that status from the array
			cr.Status.PodStatuses[i] = cr.Status.PodStatuses[length-1]
			cr.Status.PodStatuses[length-1] = v1alpha1.PodStatus{}
			cr.Status.PodStatuses = cr.Status.PodStatuses[:length-1]
			return nil
		}
	}

	log.Println("Did not find a status for POD in CR")
	return nil
}

func (r *podReconciler) updateSingleCR(cr *v1alpha1.ResourceBundleState, pod *corev1.Pod) error {

	// Update status after searching for it in the list of resourceStatuses
	for i, rstatus := range cr.Status.PodStatuses {
		// Look for the status if we already have it in the CR
		if rstatus.Name == pod.Name {
			pod.Status.DeepCopyInto(&cr.Status.PodStatuses[i].Status)
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
	cr.Status.PodStatuses = append(cr.Status.PodStatuses, v1alpha1.PodStatus{
		ObjectMeta: pod.ObjectMeta,
		Ready:      false,
		Status:     pod.Status,
	})

	err := r.client.Status().Update(context.TODO(), cr)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return err
	}

	return nil
}
