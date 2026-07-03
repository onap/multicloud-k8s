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
	"context"
	"testing"

	"github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	k8sscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// label is the selector the monitor operator filters watched resources on.
const label = "emco/deployment-id"

// deploymentID is the value used across the tests for the emco label.
const deploymentID = "test-deployment"

// testScheme builds a runtime scheme registering the built-in Kubernetes
// types (pods, services, deployments, csrs, ...) plus the ResourceBundleState
// CRD, so the controller-runtime fake client can (de)serialize every object the
// reconcilers touch.
func testScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := k8sscheme.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add k8s scheme: %v", err)
	}
	if err := v1alpha1.SchemeBuilder.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add v1alpha1 scheme: %v", err)
	}
	return scheme
}

// newFakeClient returns a controller-runtime fake client seeded with objs and
// backed by testScheme. It is the offline stand-in for a real cluster's
// kube-apiserver used by every reconciler test in this package.
func newFakeClient(t *testing.T, objs ...runtime.Object) client.Client {
	t.Helper()
	return fake.NewFakeClientWithScheme(testScheme(t), objs...)
}

// labels returns the emco label map with the shared deployment id.
func emcoLabels() map[string]string {
	return map[string]string{label: deploymentID}
}

// newRBState builds a ResourceBundleState CR carrying the emco label so the
// secondary reconcilers' label selector can find it.
func newRBState(name, namespace string) *v1alpha1.ResourceBundleState {
	return &v1alpha1.ResourceBundleState{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    emcoLabels(),
		},
	}
}

// getRBState fetches a ResourceBundleState CR back out of the fake client so a
// test can assert on the status the reconciler wrote.
func getRBState(t *testing.T, cl client.Client, name, namespace string) *v1alpha1.ResourceBundleState {
	t.Helper()
	got := &v1alpha1.ResourceBundleState{}
	err := cl.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, got)
	if err != nil {
		t.Fatalf("failed to get ResourceBundleState %s/%s: %v", namespace, name, err)
	}
	return got
}
