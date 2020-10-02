package resourcebundlestate

import (
	"context"
	"log"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	certsapi "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
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
		log.Printf("Error adding servicestatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateConfigMaps(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding configmapstatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateDeployments(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding deploymentstatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateSecrets(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding secretstatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateDaemonSets(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding daemonSetstatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateIngresses(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding ingressStatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateJobs(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding jobstatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateStatefulSets(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding statefulSetstatuses: %v\n", err)
		return reconcile.Result{}, err
	}

	err = r.updateCsrs(rbstate, rbstate.Spec.Selector.MatchLabels)
	if err != nil {
		log.Printf("Error adding csrStatuses: %v\n", err)
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

	rbstate.Status.ServiceStatuses = []corev1.Service{}

	for _, svc := range serviceList.Items {
		resStatus := corev1.Service{
			TypeMeta:   svc.TypeMeta,
			ObjectMeta: svc.ObjectMeta,
			Status:     svc.Status,
		}
		rbstate.Status.ServiceStatuses = append(rbstate.Status.ServiceStatuses, resStatus)
	}

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

	rbstate.Status.PodStatuses = []corev1.Pod{}

	for _, pod := range podList.Items {
		resStatus := corev1.Pod{
			TypeMeta:   pod.TypeMeta,
			ObjectMeta: pod.ObjectMeta,
			Status:     pod.Status,
		}
		rbstate.Status.PodStatuses = append(rbstate.Status.PodStatuses, resStatus)
	}

	return nil
}

func (r *reconciler) updateConfigMaps(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the ConfigMaps created as well
	configMapList := &corev1.ConfigMapList{}
	err := listResources(r.client, rbstate.Namespace, selectors, configMapList)
	if err != nil {
		log.Printf("Failed to list configMaps: %v", err)
		return err
	}

	rbstate.Status.ConfigMapStatuses = []corev1.ConfigMap{}

	for _, cm := range configMapList.Items {
		resStatus := corev1.ConfigMap{
			TypeMeta:   cm.TypeMeta,
			ObjectMeta: cm.ObjectMeta,
		}
		rbstate.Status.ConfigMapStatuses = append(rbstate.Status.ConfigMapStatuses, resStatus)
	}

	return nil
}

func (r *reconciler) updateDeployments(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Deployments created as well
	deploymentList := &appsv1.DeploymentList{}
	err := listResources(r.client, rbstate.Namespace, selectors, deploymentList)
	if err != nil {
		log.Printf("Failed to list deployments: %v", err)
		return err
	}

	rbstate.Status.DeploymentStatuses = []appsv1.Deployment{}

	for _, dep := range deploymentList.Items {
		resStatus := appsv1.Deployment{
			TypeMeta:   dep.TypeMeta,
			ObjectMeta: dep.ObjectMeta,
			Status:     dep.Status,
		}
		rbstate.Status.DeploymentStatuses = append(rbstate.Status.DeploymentStatuses, resStatus)
	}

	return nil
}

func (r *reconciler) updateSecrets(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Secrets created as well
	secretList := &corev1.SecretList{}
	err := listResources(r.client, rbstate.Namespace, selectors, secretList)
	if err != nil {
		log.Printf("Failed to list secrets: %v", err)
		return err
	}

	rbstate.Status.SecretStatuses = []corev1.Secret{}

	for _, sec := range secretList.Items {
		resStatus := corev1.Secret{
			TypeMeta:   sec.TypeMeta,
			ObjectMeta: sec.ObjectMeta,
		}
		rbstate.Status.SecretStatuses = append(rbstate.Status.SecretStatuses, resStatus)
	}

	return nil
}

func (r *reconciler) updateDaemonSets(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the DaemonSets created as well
	daemonSetList := &appsv1.DaemonSetList{}
	err := listResources(r.client, rbstate.Namespace, selectors, daemonSetList)
	if err != nil {
		log.Printf("Failed to list DaemonSets: %v", err)
		return err
	}

	rbstate.Status.DaemonSetStatuses = []appsv1.DaemonSet{}

	for _, ds := range daemonSetList.Items {
		resStatus := appsv1.DaemonSet{
			TypeMeta:   ds.TypeMeta,
			ObjectMeta: ds.ObjectMeta,
			Status:     ds.Status,
		}
		rbstate.Status.DaemonSetStatuses = append(rbstate.Status.DaemonSetStatuses, resStatus)
	}

	return nil
}

func (r *reconciler) updateIngresses(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Ingresses created as well
	ingressList := &v1beta1.IngressList{}
	err := listResources(r.client, rbstate.Namespace, selectors, ingressList)
	if err != nil {
		log.Printf("Failed to list ingresses: %v", err)
		return err
	}

	rbstate.Status.IngressStatuses = []v1beta1.Ingress{}

	for _, ing := range ingressList.Items {
		resStatus := v1beta1.Ingress{
			TypeMeta:   ing.TypeMeta,
			ObjectMeta: ing.ObjectMeta,
			Status:     ing.Status,
		}
		rbstate.Status.IngressStatuses = append(rbstate.Status.IngressStatuses, resStatus)
	}

	return nil
}

func (r *reconciler) updateJobs(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the Services created as well
	jobList := &v1.JobList{}
	err := listResources(r.client, rbstate.Namespace, selectors, jobList)
	if err != nil {
		log.Printf("Failed to list jobs: %v", err)
		return err
	}

	rbstate.Status.JobStatuses = []v1.Job{}

	for _, job := range jobList.Items {
		resStatus := v1.Job{
			TypeMeta:   job.TypeMeta,
			ObjectMeta: job.ObjectMeta,
			Status:     job.Status,
		}
		rbstate.Status.JobStatuses = append(rbstate.Status.JobStatuses, resStatus)
	}

	return nil
}

func (r *reconciler) updateStatefulSets(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the StatefulSets created as well
	statefulSetList := &appsv1.StatefulSetList{}
	err := listResources(r.client, rbstate.Namespace, selectors, statefulSetList)
	if err != nil {
		log.Printf("Failed to list statefulSets: %v", err)
		return err
	}

	rbstate.Status.StatefulSetStatuses = []appsv1.StatefulSet{}

	for _, sfs := range statefulSetList.Items {
		resStatus := appsv1.StatefulSet{
			TypeMeta:   sfs.TypeMeta,
			ObjectMeta: sfs.ObjectMeta,
			Status:     sfs.Status,
		}
		rbstate.Status.StatefulSetStatuses = append(rbstate.Status.StatefulSetStatuses, resStatus)
	}

	return nil
}

func (r *reconciler) updateCsrs(rbstate *v1alpha1.ResourceBundleState,
	selectors map[string]string) error {

	// Update the CR with the csrs tracked
	csrList := &certsapi.CertificateSigningRequestList{}
	err := listResources(r.client, "", selectors, csrList)
	if err != nil {
		log.Printf("Failed to list csrs: %v", err)
		return err
	}

	rbstate.Status.CsrStatuses = []certsapi.CertificateSigningRequest{}

	for _, csr := range csrList.Items {
		resStatus := certsapi.CertificateSigningRequest{
			TypeMeta:   csr.TypeMeta,
			ObjectMeta: csr.ObjectMeta,
			Status:     csr.Status,
		}
		rbstate.Status.CsrStatuses = append(rbstate.Status.CsrStatuses, resStatus)
	}

	return nil
}
