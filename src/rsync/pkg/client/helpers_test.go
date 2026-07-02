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

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/fake"
)

func TestCreateNamespace(t *testing.T) {
	cs := fake.NewSimpleClientset()
	c := &Client{Clientset: cs}

	if err := c.CreateNamespace("ns1"); err != nil {
		t.Fatalf("CreateNamespace returned an error: %s", err)
	}

	ns, err := cs.CoreV1().Namespaces().Get("ns1", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Namespace was not created: %s", err)
	}
	if ns.Labels["name"] != "ns1" {
		t.Fatalf("Expected namespace label name=ns1, got %q", ns.Labels["name"])
	}
}

func TestDeleteNamespace(t *testing.T) {
	existing := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}}
	cs := fake.NewSimpleClientset(existing)
	c := &Client{Clientset: cs}

	if err := c.DeleteNamespace("ns1"); err != nil {
		t.Fatalf("DeleteNamespace returned an error: %s", err)
	}
	if _, err := cs.CoreV1().Namespaces().Get("ns1", metav1.GetOptions{}); err == nil {
		t.Fatal("Expected the namespace to have been deleted")
	}
}

func TestNodesReady(t *testing.T) {
	readyNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-ready"},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{{Type: v1.NodeReady, Status: v1.ConditionTrue}},
		},
	}
	notReadyNode := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "node-notready"},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{{Type: v1.NodeReady, Status: v1.ConditionFalse}},
		},
	}

	t.Run("Counts ready and total nodes", func(t *testing.T) {
		cs := fake.NewSimpleClientset(readyNode, notReadyNode)
		c := &Client{Clientset: cs}
		ready, total, err := c.NodesReady()
		if err != nil {
			t.Fatalf("NodesReady returned an error: %s", err)
		}
		if ready != 1 {
			t.Fatalf("Expected 1 ready node, got %d", ready)
		}
		if total != 2 {
			t.Fatalf("Expected 2 total nodes, got %d", total)
		}
	})

	t.Run("Handles a cluster with no nodes", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		c := &Client{Clientset: cs}
		ready, total, err := c.NodesReady()
		if err != nil {
			t.Fatalf("NodesReady returned an error: %s", err)
		}
		if ready != 0 || total != 0 {
			t.Fatalf("Expected 0/0 for an empty cluster, got %d/%d", ready, total)
		}
	})
}

func TestVersion(t *testing.T) {
	cs := fake.NewSimpleClientset()
	// The fake discovery client lets us pin the reported server version.
	cs.Discovery().(*fakediscovery.FakeDiscovery).FakedServerVersion = &version.Info{
		Major:      "1",
		Minor:      "16",
		GitVersion: "v1.16.9",
	}
	c := &Client{Clientset: cs}

	v, err := c.Version()
	if err != nil {
		t.Fatalf("Version returned an error: %s", err)
	}
	if v != "v1.16.9" {
		t.Fatalf("Expected version 'v1.16.9', got %q", v)
	}
}
