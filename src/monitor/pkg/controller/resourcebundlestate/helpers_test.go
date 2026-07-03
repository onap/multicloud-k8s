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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckLabel(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected bool
	}{
		{"has emco label", map[string]string{label: "d1"}, true},
		{"has emco label among others", map[string]string{"app": "x", label: "d1"}, true},
		{"missing emco label", map[string]string{"app": "x"}, false},
		{"empty labels", map[string]string{}, false},
		{"nil labels", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkLabel(tt.labels); got != tt.expected {
				t.Errorf("checkLabel(%v) = %v, want %v", tt.labels, got, tt.expected)
			}
		})
	}
}

func TestReturnLabel(t *testing.T) {
	t.Run("returns single-key selector when label present", func(t *testing.T) {
		got := returnLabel(map[string]string{label: "d1", "app": "ignored"})
		if got == nil {
			t.Fatal("expected non-nil selector")
		}
		if len(got) != 1 || got[label] != "d1" {
			t.Errorf("returnLabel = %v, want {%s: d1}", got, label)
		}
	})
	t.Run("returns nil when label absent", func(t *testing.T) {
		if got := returnLabel(map[string]string{"app": "x"}); got != nil {
			t.Errorf("returnLabel = %v, want nil", got)
		}
	})
	t.Run("returns nil for nil labels", func(t *testing.T) {
		if got := returnLabel(nil); got != nil {
			t.Errorf("returnLabel = %v, want nil", got)
		}
	})
}

func TestListResources(t *testing.T) {
	ns := "ns1"
	// Two CRs in ns1: one labeled, one not; plus one labeled CR in ns2.
	labeled := newRBState("cr-labeled", ns)
	unlabeled := &v1alpha1.ResourceBundleState{
		ObjectMeta: metav1.ObjectMeta{Name: "cr-unlabeled", Namespace: ns},
	}
	otherNs := newRBState("cr-otherns", "ns2")
	cl := newFakeClient(t, labeled, unlabeled, otherNs)

	t.Run("filters by namespace and label selector", func(t *testing.T) {
		list := &v1alpha1.ResourceBundleStateList{}
		if err := listResources(cl, ns, emcoLabels(), list); err != nil {
			t.Fatalf("listResources err: %v", err)
		}
		if len(list.Items) != 1 || list.Items[0].Name != "cr-labeled" {
			t.Errorf("expected only cr-labeled, got %d items: %v", len(list.Items), list.Items)
		}
	})

	t.Run("nil selector returns all in namespace", func(t *testing.T) {
		list := &v1alpha1.ResourceBundleStateList{}
		if err := listResources(cl, ns, nil, list); err != nil {
			t.Fatalf("listResources err: %v", err)
		}
		if len(list.Items) != 2 {
			t.Errorf("expected 2 items in %s, got %d", ns, len(list.Items))
		}
	})
}

func TestListClusterResources(t *testing.T) {
	// listClusterResources delegates to listResources with an empty namespace,
	// so it should span namespaces while still honoring the label selector.
	a := newRBState("cr-a", "ns1")
	b := newRBState("cr-b", "ns2")
	unlabeled := &v1alpha1.ResourceBundleState{
		ObjectMeta: metav1.ObjectMeta{Name: "cr-c", Namespace: "ns3"},
	}
	cl := newFakeClient(t, a, b, unlabeled)

	list := &v1alpha1.ResourceBundleStateList{}
	if err := listClusterResources(cl, emcoLabels(), list); err != nil {
		t.Fatalf("listClusterResources err: %v", err)
	}
	if len(list.Items) != 2 {
		t.Errorf("expected 2 labeled CRs across namespaces, got %d", len(list.Items))
	}
}

// newPodMeta builds a pod with the given name/labels for predicate tests.
func newPodMeta(name string, lbls map[string]string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: lbls},
	}
}
