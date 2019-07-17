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

package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

type TestKubernetesConnector struct {
	object runtime.Object
}

func (t TestKubernetesConnector) GetMapper() meta.RESTMapper {
	return nil
}

func (t TestKubernetesConnector) GetDynamicClient() dynamic.Interface {
	return nil
}

func (t TestKubernetesConnector) GetStandardClient() kubernetes.Interface {
	return fake.NewSimpleClientset(t.object)
}

func (t TestKubernetesConnector) GetInstanceID() string {
	return ""
}

func TestCreateService(t *testing.T) {
	name := "mock-service"
	testCases := []struct {
		label          string
		input          string
		namespace      string
		object         *coreV1.Service
		expectedResult string
		expectedError  string
	}{
		{
			label:         "Fail to create a service with invalid type",
			input:         "../../mock_files/mock_yamls/deployment.yaml",
			namespace:     "test1",
			object:        &coreV1.Service{},
			expectedError: "contains another resource different than Service",
		},
		{
			label:          "Successfully create a service",
			input:          "../../mock_files/mock_yamls/service.yaml",
			namespace:      "test1",
			object:         &coreV1.Service{},
			expectedResult: name,
		},
	}

	for _, testCase := range testCases {
		client := TestKubernetesConnector{testCase.object}
		t.Run(testCase.label, func(t *testing.T) {
			result, err := servicePlugin{}.Create(testCase.input, testCase.namespace, client)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Create method returned an error (%s)", err)
				}
			} else {
				if testCase.expectedError != "" && testCase.expectedResult == "" {
					t.Fatalf("Create method was expecting \"%s\" error message", testCase.expectedError)
				}
				if result == "" {
					t.Fatal("Create method returned nil result")
				}
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("Create method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}

func TestListService(t *testing.T) {
	testCases := []struct {
		label          string
		namespace      string
		object         *coreV1.ServiceList
		expectedResult []helm.KubernetesResource
	}{
		{
			label:          "Sucessfully to display an empty service list",
			namespace:      "test1",
			object:         &coreV1.ServiceList{},
			expectedResult: []helm.KubernetesResource{},
		},
		{
			label:     "Sucessfully to display a list of existing services",
			namespace: "test1",
			object: &coreV1.ServiceList{
				Items: []coreV1.Service{
					coreV1.Service{
						ObjectMeta: metaV1.ObjectMeta{
							Name:      "test",
							Namespace: "test1",
						},
					},
				},
			},
			expectedResult: []helm.KubernetesResource{
				{
					Name: "test",
					GVK:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
				},
			},
		},
		{
			label:     "Sucessfully display a list of existing services in default namespace",
			namespace: "",
			object: &coreV1.ServiceList{
				Items: []coreV1.Service{
					coreV1.Service{
						ObjectMeta: metaV1.ObjectMeta{
							Name:      "test",
							Namespace: "default",
						},
					},
					coreV1.Service{
						ObjectMeta: metaV1.ObjectMeta{
							Name:      "test2",
							Namespace: "test1",
						},
					},
				},
			},
			expectedResult: []helm.KubernetesResource{
				{
					Name: "test",
					GVK:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
				},
			},
		},
	}

	for _, testCase := range testCases {
		client := TestKubernetesConnector{testCase.object}
		t.Run(testCase.label, func(t *testing.T) {
			result, err := servicePlugin{}.List(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Service"}, testCase.namespace, client)
			if err != nil {
				t.Fatalf("List method returned an error (%s)", err)
			} else {
				if result == nil {
					t.Fatal("List method returned nil result")
				}
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("List method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}

func TestDeleteService(t *testing.T) {
	testCases := []struct {
		label  string
		input  map[string]string
		object *coreV1.Service
	}{
		{
			label: "Sucessfully to delete an existing service",
			input: map[string]string{"name": "test-service", "namespace": "test-namespace"},
			object: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "test-service",
					Namespace: "test-namespace",
				},
			},
		},
		{
			label: "Sucessfully delete an existing service in default namespace",
			input: map[string]string{"name": "test-service", "namespace": ""},
			object: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
			},
		},
	}

	for _, testCase := range testCases {
		client := TestKubernetesConnector{testCase.object}
		t.Run(testCase.label, func(t *testing.T) {
			err := servicePlugin{}.Delete(helm.KubernetesResource{
				GVK:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
				Name: testCase.input["name"],
			}, testCase.input["namespace"], client)
			if err != nil {
				t.Fatalf("Delete method returned an error (%s)", err)
			}
		})
	}
}

func TestGetService(t *testing.T) {
	testCases := []struct {
		label          string
		input          map[string]string
		object         *coreV1.Service
		expectedResult string
		expectedError  string
	}{
		{
			label: "Sucessfully to get an existing service",
			input: map[string]string{"name": "test-service", "namespace": "test-namespace"},
			object: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "test-service",
					Namespace: "test-namespace",
				},
			},
			expectedResult: "test-service",
		},
		{
			label: "Sucessfully get an existing service from default namespaces",
			input: map[string]string{"name": "test-service", "namespace": ""},
			object: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
			},
			expectedResult: "test-service",
		},
		{
			label: "Fail to get an non-existing namespace",
			input: map[string]string{"name": "test-name", "namespace": "test-namespace"},
			object: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
			},
			expectedError: "not found",
		},
	}

	for _, testCase := range testCases {
		client := TestKubernetesConnector{testCase.object}
		t.Run(testCase.label, func(t *testing.T) {
			result, err := servicePlugin{}.Get(helm.KubernetesResource{
				GVK:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
				Name: testCase.input["name"],
			}, testCase.input["namespace"], client)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Get method returned an error (%s)", err)
				}
			} else {
				if testCase.expectedError != "" && testCase.expectedResult == "" {
					t.Fatalf("Get method was expecting \"%s\" error message", testCase.expectedError)
				}
				if result == "" {
					t.Fatal("Get method returned nil result")
				}
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("Get method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}
