/*
 * Copyright 2026 Deutsche Telekom AG
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resourcebundlestate

import (
	"testing"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	certsapi "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reqFor builds a reconcile.Request for the given name/namespace.
func reqFor(name, namespace string) reconcile.Request {
	return reconcile.Request{NamespacedName: types.NamespacedName{Name: name, Namespace: namespace}}
}

// --- Pod reconciler: full branch coverage ---------------------------------

func TestPodReconciler_AddsNewPodStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	pod := newPodMeta("pod1", emcoLabels())
	pod.Namespace = ns
	cl := newFakeClient(t, cr, pod)
	r := &podReconciler{client: cl}

	res, err := r.Reconcile(reqFor("pod1", ns))
	if err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}
	if res.Requeue {
		t.Error("did not expect requeue")
	}

	got := getRBState(t, cl, "cr1", ns)
	if got.Status.ResourceCount != 1 {
		t.Errorf("ResourceCount = %d, want 1", got.Status.ResourceCount)
	}
	if len(got.Status.PodStatuses) != 1 || got.Status.PodStatuses[0].Name != "pod1" {
		t.Errorf("PodStatuses = %v, want [pod1]", got.Status.PodStatuses)
	}
}

func TestPodReconciler_UpdatesExistingPodStatusInPlace(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	// Seed the CR with an existing status for pod1 and a resource count of 1.
	cr.Status.ResourceCount = 1
	cr.Status.PodStatuses = []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: ns},
		Status:     corev1.PodStatus{Phase: corev1.PodPending},
	}}
	// The live pod has progressed to Running.
	pod := newPodMeta("pod1", emcoLabels())
	pod.Namespace = ns
	pod.Status.Phase = corev1.PodRunning
	cl := newFakeClient(t, cr, pod)
	r := &podReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("pod1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	// Updated in place: count unchanged, phase reflects the live pod.
	if got.Status.ResourceCount != 1 {
		t.Errorf("ResourceCount = %d, want 1 (update in place, no increment)", got.Status.ResourceCount)
	}
	if len(got.Status.PodStatuses) != 1 {
		t.Fatalf("PodStatuses len = %d, want 1", len(got.Status.PodStatuses))
	}
	if got.Status.PodStatuses[0].Status.Phase != corev1.PodRunning {
		t.Errorf("pod phase = %v, want Running", got.Status.PodStatuses[0].Status.Phase)
	}
}

// TestPodReconciler_DeletionTimestampDoesNotPersistRemoval is a
// CHARACTERIZATION test: it locks in the CURRENT (buggy) behavior of the
// pod-deletion path so the regression net catches an accidental change while a
// deliberate fix stays visible.
//
// BUG (MULTICLOUD-1553): when a watched pod is marked for deletion,
// Reconcile -> updateCRs -> deleteFromSingleCR removes the status from an
// in-memory *copy* of the CR (crl.Items is ranged by value) and NEVER calls
// client.Status().Update(). Unlike the add/update path (updateSingleCR
// persists), the delete path drops its result on the floor, so the CR keeps
// the stale pod status forever. The same defect affects every resource
// reconciler in this package. A fix belongs in a separate, non-test CR; when
// it lands, flip these assertions to expect removal.
func TestPodReconciler_DeletionTimestampDoesNotPersistRemoval(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	cr.Status.ResourceCount = 1
	cr.Status.PodStatuses = []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: ns},
	}}
	now := metav1.Now()
	pod := newPodMeta("pod1", emcoLabels())
	pod.Namespace = ns
	pod.DeletionTimestamp = &now
	// A finalizer is required by the fake client to accept an object that
	// already has a deletion timestamp set.
	pod.Finalizers = []string{"keep"}
	cl := newFakeClient(t, cr, pod)
	r := &podReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("pod1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	// Current behavior: the removal is NOT persisted, so the stale status remains.
	if len(got.Status.PodStatuses) != 1 {
		t.Errorf("PodStatuses = %v; characterization expects the stale status to remain "+
			"(delete path does not persist). If the delete-path bug was fixed, update this test.",
			got.Status.PodStatuses)
	}
}

// TestPodReconciler_NotFoundDoesNotPersistRemoval is a CHARACTERIZATION test
// for the same delete-path defect via the not-found branch: when the pod no
// longer exists, deletePodFromAllCRs -> deleteFromSingleCR mutates an in-memory
// copy and never persists, so the CR keeps the stale status. See the bug note
// on TestPodReconciler_DeletionTimestampDoesNotPersistRemoval.
func TestPodReconciler_NotFoundDoesNotPersistRemoval(t *testing.T) {
	ns := "ns1"
	// CR already tracks pod1, but the pod no longer exists in the cluster.
	cr := newRBState("cr1", ns)
	cr.Status.ResourceCount = 1
	cr.Status.PodStatuses = []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: ns},
	}}
	cl := newFakeClient(t, cr)
	r := &podReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("pod1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	// Current behavior: not-found cleanup is not persisted.
	if len(got.Status.PodStatuses) != 1 {
		t.Errorf("PodStatuses = %v; characterization expects the stale status to remain "+
			"(not-found cleanup does not persist). If the delete-path bug was fixed, update this test.",
			got.Status.PodStatuses)
	}
}

func TestPodReconciler_NoTrackingCRIsNoOp(t *testing.T) {
	ns := "ns1"
	// A labeled pod exists, but there is no CR tracking it.
	pod := newPodMeta("pod1", emcoLabels())
	pod.Namespace = ns
	cl := newFakeClient(t, pod)
	r := &podReconciler{client: cl}

	res, err := r.Reconcile(reqFor("pod1", ns))
	if err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}
	if res.Requeue {
		t.Error("did not expect requeue when no CR tracks the pod")
	}
}

func TestServiceReconciler_UpdatesExistingStatusInPlace(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	cr.Status.ResourceCount = 1
	cr.Status.ServiceStatuses = []corev1.Service{{
		ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: ns},
	}}
	// Live service carries a status not yet reflected in the CR.
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: ns, Labels: emcoLabels()},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{
				Ingress: []corev1.LoadBalancerIngress{{IP: "1.2.3.4"}},
			},
		},
	}
	cl := newFakeClient(t, cr, svc)
	r := &serviceReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("svc1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	// Update in place: count stays 1, status is copied through.
	if got.Status.ResourceCount != 1 {
		t.Errorf("ResourceCount = %d, want 1 (update in place)", got.Status.ResourceCount)
	}
	if len(got.Status.ServiceStatuses) != 1 {
		t.Fatalf("ServiceStatuses len = %d, want 1", len(got.Status.ServiceStatuses))
	}
	ingress := got.Status.ServiceStatuses[0].Status.LoadBalancer.Ingress
	if len(ingress) != 1 || ingress[0].IP != "1.2.3.4" {
		t.Errorf("service status not copied through: %v", ingress)
	}
}

// --- Other resource reconcilers: representative add-path coverage ----------

func TestServiceReconciler_AddsNewServiceStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: ns, Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, svc)
	r := &serviceReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("svc1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.ServiceStatuses) != 1 || got.Status.ServiceStatuses[0].Name != "svc1" {
		t.Errorf("ServiceStatuses = %v, want [svc1]", got.Status.ServiceStatuses)
	}
	if got.Status.ResourceCount != 1 {
		t.Errorf("ResourceCount = %d, want 1", got.Status.ResourceCount)
	}
}

func TestDeploymentReconciler_AddsNewDeploymentStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: ns, Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, dep)
	r := &deploymentReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("dep1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.DeploymentStatuses) != 1 || got.Status.DeploymentStatuses[0].Name != "dep1" {
		t.Errorf("DeploymentStatuses = %v, want [dep1]", got.Status.DeploymentStatuses)
	}
}

func TestCsrReconciler_AddsNewCsrStatusClusterScoped(t *testing.T) {
	// CSRs are cluster-scoped, so the reconciler uses listClusterResources.
	// The CR lives in a namespace but is discovered across namespaces.
	cr := newRBState("cr1", "ns1")
	csr := &certsapi.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{Name: "csr1", Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, csr)
	r := &csrReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("csr1", "")); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", "ns1")
	if len(got.Status.CsrStatuses) != 1 || got.Status.CsrStatuses[0].Name != "csr1" {
		t.Errorf("CsrStatuses = %v, want [csr1]", got.Status.CsrStatuses)
	}
}

func TestConfigMapReconciler_AddsNewConfigMapStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "cm1", Namespace: ns, Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, cm)
	r := &configMapReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("cm1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.ConfigMapStatuses) != 1 || got.Status.ConfigMapStatuses[0].Name != "cm1" {
		t.Errorf("ConfigMapStatuses = %v, want [cm1]", got.Status.ConfigMapStatuses)
	}
}

func TestSecretReconciler_AddsNewSecretStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "sec1", Namespace: ns, Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, sec)
	r := &secretReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("sec1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.SecretStatuses) != 1 || got.Status.SecretStatuses[0].Name != "sec1" {
		t.Errorf("SecretStatuses = %v, want [sec1]", got.Status.SecretStatuses)
	}
}

func TestDaemonSetReconciler_AddsNewDaemonSetStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "ds1", Namespace: ns, Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, ds)
	r := &daemonSetReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("ds1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.DaemonSetStatuses) != 1 || got.Status.DaemonSetStatuses[0].Name != "ds1" {
		t.Errorf("DaemonSetStatuses = %v, want [ds1]", got.Status.DaemonSetStatuses)
	}
}

func TestIngressReconciler_AddsNewIngressStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	ing := &extv1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: "ing1", Namespace: ns, Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, ing)
	r := &ingressReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("ing1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.IngressStatuses) != 1 || got.Status.IngressStatuses[0].Name != "ing1" {
		t.Errorf("IngressStatuses = %v, want [ing1]", got.Status.IngressStatuses)
	}
}

func TestJobReconciler_AddsNewJobStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "job1", Namespace: ns, Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, job)
	r := &jobReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("job1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.JobStatuses) != 1 || got.Status.JobStatuses[0].Name != "job1" {
		t.Errorf("JobStatuses = %v, want [job1]", got.Status.JobStatuses)
	}
}

func TestStatefulSetReconciler_AddsNewStatefulSetStatus(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	sfs := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: "sfs1", Namespace: ns, Labels: emcoLabels()},
	}
	cl := newFakeClient(t, cr, sfs)
	r := &statefulSetReconciler{client: cl}

	if _, err := r.Reconcile(reqFor("sfs1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.StatefulSetStatuses) != 1 || got.Status.StatefulSetStatuses[0].Name != "sfs1" {
		t.Errorf("StatefulSetStatuses = %v, want [sfs1]", got.Status.StatefulSetStatuses)
	}
}

// --- deleteFromSingleCR: direct unit coverage of the swap-remove logic ------

func TestPodReconciler_DeleteFromSingleCR(t *testing.T) {
	r := &podReconciler{}
	cr := newRBState("cr1", "ns1")
	cr.Status.ResourceCount = 2
	cr.Status.PodStatuses = []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "pod2"}},
	}

	if err := r.deleteFromSingleCR(cr, "pod1"); err != nil {
		t.Fatalf("deleteFromSingleCR err: %v", err)
	}
	if cr.Status.ResourceCount != 1 {
		t.Errorf("ResourceCount = %d, want 1", cr.Status.ResourceCount)
	}
	if len(cr.Status.PodStatuses) != 1 || cr.Status.PodStatuses[0].Name != "pod2" {
		t.Errorf("PodStatuses = %v, want [pod2]", cr.Status.PodStatuses)
	}
}

func TestPodReconciler_DeleteFromSingleCR_MissingNameOnlyDecrementsCount(t *testing.T) {
	r := &podReconciler{}
	cr := newRBState("cr1", "ns1")
	cr.Status.ResourceCount = 1
	cr.Status.PodStatuses = []corev1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "pod1"}},
	}

	// Deleting a name that is not present still decrements the count (matching
	// the current implementation) and leaves the slice untouched.
	if err := r.deleteFromSingleCR(cr, "ghost"); err != nil {
		t.Fatalf("deleteFromSingleCR err: %v", err)
	}
	if len(cr.Status.PodStatuses) != 1 {
		t.Errorf("PodStatuses = %v, want unchanged [pod1]", cr.Status.PodStatuses)
	}
}

func TestServiceReconciler_DeleteFromSingleCR(t *testing.T) {
	r := &serviceReconciler{}
	cr := newRBState("cr1", "ns1")
	cr.Status.ResourceCount = 2
	cr.Status.ServiceStatuses = []corev1.Service{
		{ObjectMeta: metav1.ObjectMeta{Name: "svc1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "svc2"}},
	}

	if err := r.deleteFromSingleCR(cr, "svc2"); err != nil {
		t.Fatalf("deleteFromSingleCR err: %v", err)
	}
	if cr.Status.ResourceCount != 1 {
		t.Errorf("ResourceCount = %d, want 1", cr.Status.ResourceCount)
	}
	if len(cr.Status.ServiceStatuses) != 1 || cr.Status.ServiceStatuses[0].Name != "svc1" {
		t.Errorf("ServiceStatuses = %v, want [svc1]", cr.Status.ServiceStatuses)
	}
}

func TestDeploymentReconciler_DeleteFromSingleCR(t *testing.T) {
	r := &deploymentReconciler{}
	cr := newRBState("cr1", "ns1")
	cr.Status.ResourceCount = 1
	cr.Status.DeploymentStatuses = []appsv1.Deployment{
		{ObjectMeta: metav1.ObjectMeta{Name: "dep1"}},
	}

	if err := r.deleteFromSingleCR(cr, "dep1"); err != nil {
		t.Fatalf("deleteFromSingleCR err: %v", err)
	}
	if len(cr.Status.DeploymentStatuses) != 0 {
		t.Errorf("DeploymentStatuses = %v, want empty", cr.Status.DeploymentStatuses)
	}
}

var _ = v1alpha1.ResourceBundleState{}
