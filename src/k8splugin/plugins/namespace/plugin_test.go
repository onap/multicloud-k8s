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

func TestCreateNamespace(t *testing.T) {
	testCases := []struct {
		label          string
		input          string
		object         *coreV1.Namespace
		expectedResult string
		expectedError  string
	}{
		{
			label:          "Successfully create a namespace",
			input:          "test1",
			object:         &coreV1.Namespace{},
			expectedResult: "test1",
		},
	}

	for _, testCase := range testCases {
		client := TestKubernetesConnector{testCase.object}
		t.Run(testCase.label, func(t *testing.T) {
			result, err := namespacePlugin{}.Create("", testCase.input, client)
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

func TestListNamespace(t *testing.T) {
	testCases := []struct {
		label          string
		input          string
		object         *coreV1.NamespaceList
		expectedResult []helm.KubernetesResource
	}{
		{
			label:          "Sucessfully to display an empty namespace list",
			input:          "",
			object:         &coreV1.NamespaceList{},
			expectedResult: []helm.KubernetesResource{},
		},
		{
			label: "Sucessfully to display a list of existing namespaces",
			input: "test1",
			object: &coreV1.NamespaceList{
				Items: []coreV1.Namespace{
					coreV1.Namespace{
						ObjectMeta: metaV1.ObjectMeta{
							Name: "test1",
						},
					},
				},
			},
			expectedResult: []helm.KubernetesResource{
				{
					Name: "test1",
					GVK:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
				},
			},
		},
	}

	for _, testCase := range testCases {
		client := TestKubernetesConnector{testCase.object}
		t.Run(testCase.label, func(t *testing.T) {
			result, err := namespacePlugin{}.List(schema.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "Namespace",
			}, testCase.input, client)
			if err != nil {
				t.Fatalf("List method returned an error (%s)", err)
			} else {
				if result == nil {
					t.Fatal("List method returned nil result")
				}
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("List method returned: \n%+v\n and it was expected: \n%+v", result, testCase.expectedResult)
				}
			}
		})
	}
}

func TestDeleteNamespace(t *testing.T) {
	testCases := []struct {
		label  string
		input  map[string]string
		object *coreV1.Namespace
	}{
		{
			label: "Sucessfully to delete an existing namespace",
			input: map[string]string{"name": "test-name", "namespace": "test-namespace"},
			object: &coreV1.Namespace{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test-name",
				},
			},
		},
	}

	for _, testCase := range testCases {
		client := TestKubernetesConnector{testCase.object}
		t.Run(testCase.label, func(t *testing.T) {
			err := namespacePlugin{}.Delete(helm.KubernetesResource{
				GVK:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
				Name: testCase.input["name"],
			}, testCase.input["namespace"], client)
			if err != nil {
				t.Fatalf("Delete method returned an error (%s)", err)
			}
		})
	}
}

func TestGetNamespace(t *testing.T) {
	testCases := []struct {
		label          string
		input          map[string]string
		object         *coreV1.Namespace
		expectedResult string
		expectedError  string
	}{
		{
			label: "Sucessfully to get an existing namespace",
			input: map[string]string{"name": "test-name", "namespace": "test-namespace"},
			object: &coreV1.Namespace{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test-name",
				},
			},
			expectedResult: "test-name",
		},
		{
			label: "Fail to get an non-existing namespace",
			input: map[string]string{"name": "test-name", "namespace": "test-namespace"},
			object: &coreV1.Namespace{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test-name2",
				},
			},
			expectedError: "not found",
		},
	}

	for _, testCase := range testCases {
		client := TestKubernetesConnector{testCase.object}
		t.Run(testCase.label, func(t *testing.T) {
			result, err := namespacePlugin{}.Get(helm.KubernetesResource{
				GVK:  schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
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
