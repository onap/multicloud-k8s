package resourcebundlestate

import (
	"context"
	"log"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add the new controller to the controller manager
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func add(mgr manager.Manager, r *reconciler) error {
	// Create a new controller
	c, err := controller.New("ResourceBundleState-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ResourceBundleState
	err = c.Watch(&source.Kind{Type: &v1alpha1.ResourceBundleState{}}, &EventHandler{})
	if err != nil {
		return err
	}

	return nil
}

func newReconciler(m manager.Manager) *reconciler {
	return &reconciler{client: m.GetClient()}
}

type reconciler struct {
	// Stores an array of all the ResourceBundleState
	crList []v1alpha1.ResourceBundleState
	client client.Client
}

// Reconcile implements the loop that will manage the ResourceBundleState CR
// We only accept CREATE events here and any subsequent changes are ignored.
func (r *reconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	log.Printf("Reconciling ResourceBundleState %+v\n", req)

	rbstate := &v1alpha1.ResourceBundleState{}
	err := r.client.Get(context.TODO(), req.NamespacedName, rbstate)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			log.Printf("Object not found: %+v. Ignore as it must have been deleted.\n", req.NamespacedName)
			return reconcile.Result{}, nil
		}
		log.Printf("Failed to get object: %+v\n", req.NamespacedName)
		return reconcile.Result{}, err
	}

	err = r.updatePods(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding podstatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateServices(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding services: %v\n", err)
		return reconcile.Result{}, err
	}

	// TODO: Update this based on the statuses of the lower resources
	rbstate.Status.Ready = false
	err = r.client.Status().Update(context.TODO(), rbstate)
	if err != nil {
		log.Printf("failed to update rbstate: %v\n", err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *reconciler) updateServices(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Services created as well
	serviceList := &corev1.ServiceList{}
	err := listResources(r.client, rbstate.Namespace, selectors, serviceList)
	if err != nil {
		log.Printf("Failed to list services: %v", err)
		return err
	}

	rbstate.Status.ServiceStatuses = serviceList.Items
	return nil
}

func (r *reconciler) updatePods(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the pods tracked
	podList := &corev1.PodList{}
	err := listResources(r.client, rbstate.Namespace, selectors, podList)
	if err != nil {
		log.Printf("Failed to list pods: %v", err)
		return err
	}

	rbstate.Status.PodStatuses = []v1alpha1.PodStatus{}

	for _, pod := range podList.Items {
		resStatus := v1alpha1.PodStatus{
			ObjectMeta: pod.ObjectMeta,
			Ready:      false,
			Status:     pod.Status,
		}
		rbstate.Status.PodStatuses = append(rbstate.Status.PodStatuses, resStatus)
	}

	return nil
}
