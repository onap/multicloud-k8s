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

	utils "k8splugin/internal"

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
		input          *utils.ResourceData
		clientOutput   *coreV1.Service
		expectedResult string
		expectedError  string
	}{
		{
			label: "Fail to create a service with invalid type",
			input: &utils.ResourceData{
				YamlFilePath: "../../mock_files/mock_yamls/deployment.yaml",
			},
			clientOutput:  &coreV1.Service{},
			expectedError: "contains another resource different than Service",
		},
		{
			label: "Successfully create a service",
			input: &utils.ResourceData{
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
		{
			label: "Sucessfully display a list of existing services in default namespace",
			input: "",
			clientOutput: &coreV1.ServiceList{
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
		label        string
		input        map[string]string
		clientOutput *coreV1.Service
	}{
		{
			label: "Sucessfully to delete an existing service",
			input: map[string]string{"name": "test-service", "namespace": "test-namespace"},
			clientOutput: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "test-service",
					Namespace: "test-namespace",
				},
			},
		},
		{
			label: "Sucessfully delete an existing service in default namespace",
			input: map[string]string{"name": "test-service", "namespace": ""},
			clientOutput: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
			},
		},
	}

	for _, testCase := range testCases {
		client := testclient.NewSimpleClientset(testCase.clientOutput)
		t.Run(testCase.label, func(t *testing.T) {
			err := Delete(testCase.input["name"], testCase.input["namespace"], client)
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
		clientOutput   *coreV1.Service
		expectedResult string
		expectedError  string
	}{
		{
			label: "Sucessfully to get an existing service",
			input: map[string]string{"name": "test-service", "namespace": "test-namespace"},
			clientOutput: &coreV1.Service{
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
			clientOutput: &coreV1.Service{
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
			clientOutput: &coreV1.Service{
				ObjectMeta: metaV1.ObjectMeta{
					Name:      "test-service",
					Namespace: "default",
				},
			},
			expectedError: "not found",
		},
	}

	for _, testCase := range testCases {
		client := testclient.NewSimpleClientset(testCase.clientOutput)
		t.Run(testCase.label, func(t *testing.T) {
			result, err := Get(testCase.input["name"], testCase.input["namespace"], client)
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
