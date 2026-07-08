/*
Copyright 2018 Intel Corporation.
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

package utils

import (
	"context"
	"testing"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	fakeclient "k8s.io/client-go/kubernetes/fake"
)

// int32Ptr returns a pointer to the given int32 value.
func int32Ptr(i int32) *int32 {
	return &i
}

// newTestDeployment builds a Deployment with the provided replicas and strategy.
func newTestDeployment(replicas int32, strategyType apps.DeploymentStrategyType,
	maxSurge, maxUnavailable *intstrutil.IntOrString) *apps.Deployment {
	return &apps.Deployment{
		Spec: apps.DeploymentSpec{
			Replicas: int32Ptr(replicas),
			Strategy: apps.DeploymentStrategy{
				Type: strategyType,
				RollingUpdate: &apps.RollingUpdateDeployment{
					MaxSurge:       maxSurge,
					MaxUnavailable: maxUnavailable,
				},
			},
		},
	}
}

func TestIsRollingUpdate(t *testing.T) {
	testCases := []struct {
		label    string
		strategy apps.DeploymentStrategyType
		expected bool
	}{
		{
			label:    "RollingUpdate strategy returns true",
			strategy: apps.RollingUpdateDeploymentStrategyType,
			expected: true,
		},
		{
			label:    "Recreate strategy returns false",
			strategy: apps.RecreateDeploymentStrategyType,
			expected: false,
		},
		{
			label:    "Empty strategy returns false",
			strategy: apps.DeploymentStrategyType(""),
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			d := &apps.Deployment{
				Spec: apps.DeploymentSpec{
					Strategy: apps.DeploymentStrategy{Type: testCase.strategy},
				},
			}
			if got := IsRollingUpdate(d); got != testCase.expected {
				t.Fatalf("IsRollingUpdate returned %v, expected %v", got, testCase.expected)
			}
		})
	}
}

func TestMaxUnavailable(t *testing.T) {
	testCases := []struct {
		label          string
		replicas       int32
		strategy       apps.DeploymentStrategyType
		maxSurge       *intstrutil.IntOrString
		maxUnavailable *intstrutil.IntOrString
		expected       int32
	}{
		{
			label:          "Not a rolling update returns zero",
			replicas:       5,
			strategy:       apps.RecreateDeploymentStrategyType,
			maxSurge:       intOrStringPtr(intstrutil.FromInt(1)),
			maxUnavailable: intOrStringPtr(intstrutil.FromInt(1)),
			expected:       0,
		},
		{
			label:          "Zero replicas returns zero",
			replicas:       0,
			strategy:       apps.RollingUpdateDeploymentStrategyType,
			maxSurge:       intOrStringPtr(intstrutil.FromInt(1)),
			maxUnavailable: intOrStringPtr(intstrutil.FromInt(1)),
			expected:       0,
		},
		{
			label:          "Integer maxUnavailable",
			replicas:       5,
			strategy:       apps.RollingUpdateDeploymentStrategyType,
			maxSurge:       intOrStringPtr(intstrutil.FromInt(2)),
			maxUnavailable: intOrStringPtr(intstrutil.FromInt(2)),
			expected:       2,
		},
		{
			label:          "maxUnavailable capped at replicas",
			replicas:       3,
			strategy:       apps.RollingUpdateDeploymentStrategyType,
			maxSurge:       intOrStringPtr(intstrutil.FromInt(0)),
			maxUnavailable: intOrStringPtr(intstrutil.FromInt(10)),
			expected:       3,
		},
		{
			label:          "Percentage maxUnavailable",
			replicas:       10,
			strategy:       apps.RollingUpdateDeploymentStrategyType,
			maxSurge:       intOrStringPtr(intstrutil.FromString("0%")),
			maxUnavailable: intOrStringPtr(intstrutil.FromString("30%")),
			expected:       3,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			d := newTestDeployment(testCase.replicas, testCase.strategy, testCase.maxSurge, testCase.maxUnavailable)
			if got := MaxUnavailable(*d); got != testCase.expected {
				t.Fatalf("MaxUnavailable returned %v, expected %v", got, testCase.expected)
			}
		})
	}
}

// intOrStringPtr returns a pointer to the given IntOrString.
func intOrStringPtr(i intstrutil.IntOrString) *intstrutil.IntOrString {
	return &i
}

func TestResolveFenceposts(t *testing.T) {
	testCases := []struct {
		label               string
		maxSurge            *intstrutil.IntOrString
		maxUnavailable      *intstrutil.IntOrString
		desired             int32
		expectedSurge       int32
		expectedUnavailable int32
		expectError         bool
	}{
		{
			label:               "Integer surge and unavailable",
			maxSurge:            intOrStringPtr(intstrutil.FromInt(1)),
			maxUnavailable:      intOrStringPtr(intstrutil.FromInt(1)),
			desired:             2,
			expectedSurge:       1,
			expectedUnavailable: 1,
		},
		{
			label:               "Percentage surge and unavailable",
			maxSurge:            intOrStringPtr(intstrutil.FromString("25%")),
			maxUnavailable:      intOrStringPtr(intstrutil.FromString("25%")),
			desired:             8,
			expectedSurge:       2, // rounds up
			expectedUnavailable: 2, // rounds down
		},
		{
			label:               "Both zero forces unavailable to one",
			maxSurge:            intOrStringPtr(intstrutil.FromInt(0)),
			maxUnavailable:      intOrStringPtr(intstrutil.FromInt(0)),
			desired:             2,
			expectedSurge:       0,
			expectedUnavailable: 1,
		},
		{
			label:               "Nil surge and unavailable defaults to zero, forces unavailable one",
			maxSurge:            nil,
			maxUnavailable:      nil,
			desired:             3,
			expectedSurge:       0,
			expectedUnavailable: 1,
		},
		{
			label:          "Invalid percentage string returns error",
			maxSurge:       intOrStringPtr(intstrutil.FromString("not-a-percent")),
			maxUnavailable: intOrStringPtr(intstrutil.FromInt(1)),
			desired:        3,
			expectError:    true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			surge, unavailable, err := ResolveFenceposts(testCase.maxSurge, testCase.maxUnavailable, testCase.desired)
			if testCase.expectError {
				if err == nil {
					t.Fatalf("ResolveFenceposts expected an error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveFenceposts returned unexpected error: %v", err)
			}
			if surge != testCase.expectedSurge {
				t.Fatalf("ResolveFenceposts surge returned %v, expected %v", surge, testCase.expectedSurge)
			}
			if unavailable != testCase.expectedUnavailable {
				t.Fatalf("ResolveFenceposts unavailable returned %v, expected %v", unavailable, testCase.expectedUnavailable)
			}
		})
	}
}

func TestEqualIgnoreHash(t *testing.T) {
	baseTemplate := func() *v1.PodTemplateSpec {
		return &v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "test"},
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Name: "c1", Image: "image:v1"},
				},
			},
		}
	}

	t.Run("Equal templates return true", func(t *testing.T) {
		t1 := baseTemplate()
		t2 := baseTemplate()
		if !EqualIgnoreHash(t1, t2) {
			t.Fatalf("EqualIgnoreHash returned false for equal templates")
		}
	})

	t.Run("Templates differing only by hash label are equal", func(t *testing.T) {
		t1 := baseTemplate()
		t2 := baseTemplate()
		t2.Labels[apps.DefaultDeploymentUniqueLabelKey] = "abc123"
		if !EqualIgnoreHash(t1, t2) {
			t.Fatalf("EqualIgnoreHash returned false when only the hash label differs")
		}
	})

	t.Run("Templates differing by a real label are not equal", func(t *testing.T) {
		t1 := baseTemplate()
		t2 := baseTemplate()
		t2.Labels["extra"] = "value"
		if EqualIgnoreHash(t1, t2) {
			t.Fatalf("EqualIgnoreHash returned true when a real label differs")
		}
	})

	t.Run("Templates differing by container image are not equal", func(t *testing.T) {
		t1 := baseTemplate()
		t2 := baseTemplate()
		t2.Spec.Containers[0].Image = "image:v2"
		if EqualIgnoreHash(t1, t2) {
			t.Fatalf("EqualIgnoreHash returned true when the container image differs")
		}
	})

	t.Run("Original templates are not mutated", func(t *testing.T) {
		t1 := baseTemplate()
		t1.Labels[apps.DefaultDeploymentUniqueLabelKey] = "hash1"
		t2 := baseTemplate()
		t2.Labels[apps.DefaultDeploymentUniqueLabelKey] = "hash2"
		EqualIgnoreHash(t1, t2)
		if _, ok := t1.Labels[apps.DefaultDeploymentUniqueLabelKey]; !ok {
			t.Fatalf("EqualIgnoreHash mutated the original template labels")
		}
	})
}

// deploymentWithTemplate builds a Deployment whose pod template carries the
// given labels and container image.
func deploymentWithTemplate(labels map[string]string, image string) *apps.Deployment {
	return &apps.Deployment{
		Spec: apps.DeploymentSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "c1", Image: image}},
				},
			},
		},
	}
}

// replicaSetWithTemplate builds a ReplicaSet with the given name, creation
// order and pod template.
func replicaSetWithTemplate(name string, order int64, labels map[string]string, image string) *apps.ReplicaSet {
	return &apps.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              name,
			CreationTimestamp: metav1.Unix(order, 0),
		},
		Spec: apps.ReplicaSetSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "c1", Image: image}},
				},
			},
		},
	}
}

func TestFindNewReplicaSet(t *testing.T) {
	deploymentLabels := map[string]string{"app": "test"}

	t.Run("Returns matching ReplicaSet", func(t *testing.T) {
		d := deploymentWithTemplate(deploymentLabels, "image:v2")
		matching := replicaSetWithTemplate("rs-new", 20, map[string]string{
			"app":                                "test",
			apps.DefaultDeploymentUniqueLabelKey: "hashnew",
		}, "image:v2")
		old := replicaSetWithTemplate("rs-old", 10, map[string]string{
			"app":                                "test",
			apps.DefaultDeploymentUniqueLabelKey: "hashold",
		}, "image:v1")

		got := FindNewReplicaSet(d, []*apps.ReplicaSet{matching, old})
		if got == nil {
			t.Fatalf("FindNewReplicaSet returned nil, expected rs-new")
		}
		if got.Name != "rs-new" {
			t.Fatalf("FindNewReplicaSet returned %s, expected rs-new", got.Name)
		}
	})

	t.Run("Returns nil when no ReplicaSet matches", func(t *testing.T) {
		d := deploymentWithTemplate(deploymentLabels, "image:v3")
		old := replicaSetWithTemplate("rs-old", 10, deploymentLabels, "image:v1")
		if got := FindNewReplicaSet(d, []*apps.ReplicaSet{old}); got != nil {
			t.Fatalf("FindNewReplicaSet returned %v, expected nil", got)
		}
	})

	t.Run("Chooses oldest ReplicaSet when multiple match", func(t *testing.T) {
		d := deploymentWithTemplate(deploymentLabels, "image:v2")
		newer := replicaSetWithTemplate("rs-newer", 30, deploymentLabels, "image:v2")
		older := replicaSetWithTemplate("rs-older", 15, deploymentLabels, "image:v2")

		got := FindNewReplicaSet(d, []*apps.ReplicaSet{newer, older})
		if got == nil {
			t.Fatalf("FindNewReplicaSet returned nil, expected rs-older")
		}
		if got.Name != "rs-older" {
			t.Fatalf("FindNewReplicaSet returned %s, expected the oldest matching rs-older", got.Name)
		}
	})
}

func TestListReplicaSets(t *testing.T) {
	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			UID:       "deployment-uid",
		},
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
		},
	}

	t.Run("Only owned ReplicaSets are returned", func(t *testing.T) {
		owned := &apps.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "owned",
				Namespace: "default",
				OwnerReferences: []metav1.OwnerReference{
					*metav1.NewControllerRef(deployment,
						apps.SchemeGroupVersion.WithKind("Deployment")),
				},
			},
		}
		orphan := &apps.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "orphan",
				Namespace: "default",
			},
		}

		listFunc := func(namespace string, options metav1.ListOptions) ([]*apps.ReplicaSet, error) {
			return []*apps.ReplicaSet{owned, orphan}, nil
		}

		result, err := ListReplicaSets(deployment, listFunc)
		if err != nil {
			t.Fatalf("ListReplicaSets returned unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].Name != "owned" {
			t.Fatalf("ListReplicaSets returned %v, expected only the owned ReplicaSet", result)
		}
	})

	t.Run("Invalid selector returns error", func(t *testing.T) {
		badDeployment := &apps.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: "default"},
			Spec: apps.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      "app",
							Operator: metav1.LabelSelectorOperator("BadOperator"),
						},
					},
				},
			},
		}
		listFunc := func(namespace string, options metav1.ListOptions) ([]*apps.ReplicaSet, error) {
			return nil, nil
		}
		if _, err := ListReplicaSets(badDeployment, listFunc); err == nil {
			t.Fatalf("ListReplicaSets expected an error for an invalid selector")
		}
	})
}

func TestRsListFromClientAndGetNewReplicaSet(t *testing.T) {
	deploymentLabels := map[string]string{"app": "test"}
	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
			UID:       "deployment-uid",
		},
		Spec: apps.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: deploymentLabels},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: deploymentLabels},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "c1", Image: "image:v2"}},
				},
			},
		},
	}

	matchingRS := &apps.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "matching-rs",
			Namespace: "default",
			Labels:    deploymentLabels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(deployment,
					apps.SchemeGroupVersion.WithKind("Deployment")),
			},
		},
		Spec: apps.ReplicaSetSpec{
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: deploymentLabels},
				Spec: v1.PodSpec{
					Containers: []v1.Container{{Name: "c1", Image: "image:v2"}},
				},
			},
		},
	}

	client := fakeclient.NewSimpleClientset(matchingRS)

	t.Run("RsListFromClient lists ReplicaSets", func(t *testing.T) {
		listFunc := RsListFromClient(client.AppsV1())
		result, err := listFunc("default", metav1.ListOptions{})
		if err != nil {
			t.Fatalf("RsListFromClient returned unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].Name != "matching-rs" {
			t.Fatalf("RsListFromClient returned %v, expected the matching ReplicaSet", result)
		}
	})

	t.Run("GetNewReplicaSet returns matching ReplicaSet", func(t *testing.T) {
		rs, err := GetNewReplicaSet(deployment, client.AppsV1())
		if err != nil {
			t.Fatalf("GetNewReplicaSet returned unexpected error: %v", err)
		}
		if rs == nil || rs.Name != "matching-rs" {
			t.Fatalf("GetNewReplicaSet returned %v, expected matching-rs", rs)
		}
	})

	_ = context.Background()
}
