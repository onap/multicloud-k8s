// +build unit

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

	"k8splugin/krd"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestCreateNamespace(t *testing.T) {
	namespace := "test1"
	testCases := []struct {
		label          string
		input          *krd.ResourceData
		clientOutput   *coreV1.Namespace
		expectedResult string
		expectedError  string
	}{
		{
			label: "Successfully create a namespace",
			input: &krd.ResourceData{
				Namespace: namespace,
			},
			clientOutput:   &coreV1.Namespace{},
			expectedResult: namespace,
		},
	}

	for _, testCase := range testCases {
		client := testclient.NewSimpleClientset(testCase.clientOutput)
		t.Run(testCase.label, func(t *testing.T) {
			result, err := Create(testCase.input, client)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Create method returned an error (%s)", err)
				}
			}
			if !reflect.DeepEqual(testCase.expectedResult, result) {
				t.Fatalf("Create method returned %v and it was expected (%v)", result, testCase.expectedResult)
			}
		})
	}
}

func TestListNamespace(t *testing.T) {
	namespace := "test1"
	testCases := []struct {
		label          string
		input          string
		clientOutput   *coreV1.NamespaceList
		expectedResult []string
	}{
		{
			label:          "Sucessfully to display an empty namespace list",
			input:          namespace,
			clientOutput:   &coreV1.NamespaceList{},
			expectedResult: []string{},
		},
		{
			label: "Sucessfully to display a list of existing namespaces",
			input: namespace,
			clientOutput: &coreV1.NamespaceList{
				Items: []coreV1.Namespace{
					coreV1.Namespace{
						ObjectMeta: metaV1.ObjectMeta{
							Name: namespace,
						},
					},
				},
			},
			expectedResult: []string{namespace},
		},
	}

	for _, testCase := range testCases {
		client := testclient.NewSimpleClientset(testCase.clientOutput)
		t.Run(testCase.label, func(t *testing.T) {
			result, err := List(testCase.input, client)
			if err != nil {
				t.Fatalf("List method returned an error (%s)", err)
			}
			if !reflect.DeepEqual(testCase.expectedResult, result) {
				t.Fatalf("List method returned %v and it was expected (%v)", result, testCase.expectedResult)
			}
		})
	}
}

func TestDeleteNamespace(t *testing.T) {
	namespace := "test1"
	testCases := []struct {
		label        string
		input        string
		clientOutput *coreV1.Namespace
	}{
		{
			label: "Sucessfully to delete an existing namespace",
			input: namespace,
			clientOutput: &coreV1.Namespace{
				ObjectMeta: metaV1.ObjectMeta{
					Name: namespace,
				},
			},
		},
	}

	for _, testCase := range testCases {
		client := testclient.NewSimpleClientset(testCase.clientOutput)
		t.Run(testCase.label, func(t *testing.T) {
			err := Delete(testCase.input, namespace, client)
			if err != nil {
				t.Fatalf("Delete method returned an error (%s)", err)
			}
		})
	}
}

func TestGetNamespace(t *testing.T) {
	namespace := "test1"
	testCases := []struct {
		label          string
		input          string
		clientOutput   *coreV1.Namespace
		expectedResult string
	}{
		{
			label: "Sucessfully to get an existing namespace",
			input: namespace,
			clientOutput: &coreV1.Namespace{
				ObjectMeta: metaV1.ObjectMeta{
					Name: namespace,
				},
			},
			expectedResult: namespace,
		},
	}

	for _, testCase := range testCases {
		client := testclient.NewSimpleClientset(testCase.clientOutput)
		t.Run(testCase.label, func(t *testing.T) {
			result, err := Get(testCase.input, namespace, client)
			if err != nil {
				t.Fatalf("Get method returned an error (%s)", err)
			}
			if !reflect.DeepEqual(testCase.expectedResult, result) {
				t.Fatalf("Get method returned %v and it was expected (%v)", result, testCase.expectedResult)
			}
		})
	}
}
