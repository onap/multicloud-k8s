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
)

// TestPrimaryReconciler_AggregatesAllResourceKinds seeds one of every watched
// resource kind (all carrying the emco selector) plus the ResourceBundleState
// CR, then drives the primary reconciler and asserts each status slice was
// populated. This is the operator's central status-aggregation path.
func TestPrimaryReconciler_AggregatesAllResourceKinds(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	// The primary reconciler reads the selector off the CR spec.
	cr.Spec.Selector = &metav1.LabelSelector{MatchLabels: emcoLabels()}

	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: ns, Labels: emcoLabels()}}
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: ns, Labels: emcoLabels()}}
	cm := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "cm1", Namespace: ns, Labels: emcoLabels()}}
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: ns, Labels: emcoLabels()}}
	sec := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "sec1", Namespace: ns, Labels: emcoLabels()}}
	ds := &appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "ds1", Namespace: ns, Labels: emcoLabels()}}
	ing := &extv1beta1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "ing1", Namespace: ns, Labels: emcoLabels()}}
	job := &batchv1.Job{ObjectMeta: metav1.ObjectMeta{Name: "job1", Namespace: ns, Labels: emcoLabels()}}
	sfs := &appsv1.StatefulSet{ObjectMeta: metav1.ObjectMeta{Name: "sfs1", Namespace: ns, Labels: emcoLabels()}}
	csr := &certsapi.CertificateSigningRequest{ObjectMeta: metav1.ObjectMeta{Name: "csr1", Labels: emcoLabels()}}

	cl := newFakeClient(t, cr, pod, svc, cm, dep, sec, ds, ing, job, sfs, csr)
	r := &reconciler{client: cl}

	if _, err := r.Reconcile(reqFor("cr1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	checks := []struct {
		name string
		got  int
	}{
		{"PodStatuses", len(got.Status.PodStatuses)},
		{"ServiceStatuses", len(got.Status.ServiceStatuses)},
		{"ConfigMapStatuses", len(got.Status.ConfigMapStatuses)},
		{"DeploymentStatuses", len(got.Status.DeploymentStatuses)},
		{"SecretStatuses", len(got.Status.SecretStatuses)},
		{"DaemonSetStatuses", len(got.Status.DaemonSetStatuses)},
		{"IngressStatuses", len(got.Status.IngressStatuses)},
		{"JobStatuses", len(got.Status.JobStatuses)},
		{"StatefulSetStatuses", len(got.Status.StatefulSetStatuses)},
		{"CsrStatuses", len(got.Status.CsrStatuses)},
	}
	for _, c := range checks {
		if c.got != 1 {
			t.Errorf("%s length = %d, want 1", c.name, c.got)
		}
	}
	// The current implementation always sets Ready=false (see the TODO in
	// controller.go); lock that behavior in so a future change is deliberate.
	if got.Status.Ready {
		t.Error("Status.Ready = true, want false (current implementation)")
	}
}

// TestPrimaryReconciler_FiltersUnlabeledResources verifies the selector is
// honored: resources without the emco label are not aggregated into the CR.
func TestPrimaryReconciler_FiltersUnlabeledResources(t *testing.T) {
	ns := "ns1"
	cr := newRBState("cr1", ns)
	cr.Spec.Selector = &metav1.LabelSelector{MatchLabels: emcoLabels()}

	tracked := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: ns, Labels: emcoLabels()}}
	untracked := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: ns, Labels: map[string]string{"app": "other"}}}

	cl := newFakeClient(t, cr, tracked, untracked)
	r := &reconciler{client: cl}

	if _, err := r.Reconcile(reqFor("cr1", ns)); err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}

	got := getRBState(t, cl, "cr1", ns)
	if len(got.Status.PodStatuses) != 1 || got.Status.PodStatuses[0].Name != "pod1" {
		t.Errorf("PodStatuses = %v, want only [pod1]", got.Status.PodStatuses)
	}
}

// TestPrimaryReconciler_NotFoundIsNoOp verifies a reconcile for a CR that no
// longer exists returns cleanly without error.
func TestPrimaryReconciler_NotFoundIsNoOp(t *testing.T) {
	cl := newFakeClient(t)
	r := &reconciler{client: cl}

	res, err := r.Reconcile(reqFor("missing", "ns1"))
	if err != nil {
		t.Fatalf("Reconcile err: %v", err)
	}
	if res.Requeue {
		t.Error("did not expect requeue for missing CR")
	}
}

var _ = v1alpha1.ResourceBundleStatus{}
