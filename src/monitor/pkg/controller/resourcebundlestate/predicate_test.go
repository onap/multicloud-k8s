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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// allPredicates returns one instance of every resource predicate in this
// package. They all share the same emco-label filtering contract, so the tests
// assert the contract uniformly across them.
func allPredicates() map[string]predicate.Predicate {
	return map[string]predicate.Predicate{
		"pod":         &podPredicate{},
		"service":     &servicePredicate{},
		"configMap":   &configMapPredicate{},
		"deployment":  &deploymentPredicate{},
		"secret":      &secretPredicate{},
		"daemonSet":   &daemonSetPredicate{},
		"ingress":     &ingressPredicate{},
		"job":         &jobPredicate{},
		"statefulSet": &statefulSetPredicate{},
		"csr":         &csrPredicate{},
	}
}

func metaWith(lbls map[string]string) metav1.Object {
	return &metav1.ObjectMeta{Labels: lbls}
}

func TestPredicateCreate(t *testing.T) {
	for name, p := range allPredicates() {
		t.Run(name+"/labeled passes", func(t *testing.T) {
			if !p.Create(event.CreateEvent{Meta: metaWith(emcoLabels())}) {
				t.Error("expected Create to pass for labeled object")
			}
		})
		t.Run(name+"/unlabeled filtered", func(t *testing.T) {
			if p.Create(event.CreateEvent{Meta: metaWith(map[string]string{"app": "x"})}) {
				t.Error("expected Create to filter unlabeled object")
			}
		})
		t.Run(name+"/nil meta filtered", func(t *testing.T) {
			if p.Create(event.CreateEvent{Meta: nil}) {
				t.Error("expected Create to filter nil meta")
			}
		})
	}
}

func TestPredicateDelete(t *testing.T) {
	for name, p := range allPredicates() {
		t.Run(name+"/labeled passes", func(t *testing.T) {
			if !p.Delete(event.DeleteEvent{Meta: metaWith(emcoLabels())}) {
				t.Error("expected Delete to pass for labeled object")
			}
		})
		t.Run(name+"/nil meta filtered", func(t *testing.T) {
			if p.Delete(event.DeleteEvent{Meta: nil}) {
				t.Error("expected Delete to filter nil meta")
			}
		})
	}
}

func TestPredicateUpdate(t *testing.T) {
	for name, p := range allPredicates() {
		t.Run(name+"/labeled new meta passes", func(t *testing.T) {
			if !p.Update(event.UpdateEvent{MetaNew: metaWith(emcoLabels())}) {
				t.Error("expected Update to pass for labeled new object")
			}
		})
		t.Run(name+"/nil new meta filtered", func(t *testing.T) {
			if p.Update(event.UpdateEvent{MetaNew: nil}) {
				t.Error("expected Update to filter nil new meta")
			}
		})
	}
}

func TestPredicateGeneric(t *testing.T) {
	for name, p := range allPredicates() {
		t.Run(name+"/labeled passes", func(t *testing.T) {
			if !p.Generic(event.GenericEvent{Meta: metaWith(emcoLabels())}) {
				t.Error("expected Generic to pass for labeled object")
			}
		})
		t.Run(name+"/unlabeled filtered", func(t *testing.T) {
			if p.Generic(event.GenericEvent{Meta: metaWith(nil)}) {
				t.Error("expected Generic to filter object without emco label")
			}
		})
	}
}
