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

func TestCreateService(t *testing.T) {
	namespace := "test1"
	name := "mock-service"
	internalVNFID := "1"
	testCases := []struct {
		label          string
		input          *krd.ResourceData
		clientOutput   *coreV1.Service
		expectedResult string
		expectedError  string
	}{
		{
			label: "Fail to create a service with non-existing file",
			input: &krd.ResourceData{
				YamlFilePath: "non-existing_test_file.yaml",
			},
			clientOutput:  &coreV1.Service{},
			expectedError: "not found",
		},
		{
			label: "Fail to create a service with invalid type",
			input: &krd.ResourceData{
				YamlFilePath: "../../mock_files/mock_yamls/deployment.yaml",
			},
			clientOutput:  &coreV1.Service{},
			expectedError: "contains another resource different than Service",
		},
		{
			label: "Successfully create a service",
			input: &krd.ResourceData{
				VnfId:        internalVNFID,
				YamlFilePath: "../../mock_files/mock_yamls/service.yaml",
			},
			clientOutput: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			},
			expectedResult: internalVNFID + "-" + name,
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

func TestListService(t *testing.T) {
	namespace := "test1"
	testCases := []struct {
		label          string
		input          string
		clientOutput   *coreV1.ServiceList
		expectedResult []string
	}{
		{
			label:          "Sucessfully to display an empty service list",
			input:          namespace,
			clientOutput:   &coreV1.ServiceList{},
			expectedResult: []string{},
		},
		{
			label: "Sucessfully to display a list of existing services",
			input: namespace,
			clientOutput: &coreV1.ServiceList{
				Items: []coreV1.Service{
					coreV1.Service{
						ObjectMeta: metaV1.ObjectMeta{
							Name:      "test",
							Namespace: namespace,
						},
					},
				},
			},
			expectedResult: []string{"test"},
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

func TestDeleteService(t *testing.T) {
	namespace := "test1"
	name := "mock-service"
	testCases := []struct {
		label        string
		input        string
		clientOutput *coreV1.Service
	}{
		{
			label: "Sucessfully to delete an existing service",
			input: name,
			clientOutput: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
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

func TestGetService(t *testing.T) {
	namespace := "test1"
	name := "mock-service"
	testCases := []struct {
		label          string
		input          string
		clientOutput   *coreV1.Service
		expectedResult string
	}{
		{
			label: "Sucessfully to get an existing service",
			input: name,
			clientOutput: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
			},
			expectedResult: name,
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
