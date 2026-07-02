/*
Copyright 2026 Deutsche Telekom AG
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package client

import (
	"testing"

	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestApprove(t *testing.T) {
	approvalJSON := []byte(`{"lastUpdateTime":"2020-01-01T00:00:00Z","message":"approved by test","reason":"AutoApproved","type":"Approved"}`)

	t.Run("Appends an approval condition to an existing CSR", func(t *testing.T) {
		csr := &certificatesv1beta1.CertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{Name: "csr1"},
		}
		cs := fake.NewSimpleClientset(csr)
		c := &Client{Clientset: cs}

		if err := c.Approve("csr1", approvalJSON); err != nil {
			t.Fatalf("Approve returned an error: %s", err)
		}

		updated, err := cs.CertificatesV1beta1().CertificateSigningRequests().Get("csr1", metav1.GetOptions{})
		if err != nil {
			t.Fatalf("Failed to fetch updated CSR: %s", err)
		}
		if len(updated.Status.Conditions) != 1 {
			t.Fatalf("Expected 1 condition, got %d", len(updated.Status.Conditions))
		}
		cond := updated.Status.Conditions[0]
		if cond.Type != certificatesv1beta1.CertificateApproved {
			t.Fatalf("Expected condition type 'Approved', got %q", cond.Type)
		}
		if cond.Reason != "AutoApproved" {
			t.Fatalf("Expected reason 'AutoApproved', got %q", cond.Reason)
		}
	})

	t.Run("Returns an error for malformed approval json", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		c := &Client{Clientset: cs}
		if err := c.Approve("csr1", []byte("{not json")); err == nil {
			t.Fatal("Expected an error for malformed approval json, got nil")
		}
	})

	t.Run("Returns an error when the CSR does not exist", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		c := &Client{Clientset: cs}
		if err := c.Approve("missing", approvalJSON); err == nil {
			t.Fatal("Expected an error for a missing CSR, got nil")
		}
	})
}
