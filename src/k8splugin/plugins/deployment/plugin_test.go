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

	appsV1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	testclient "k8s.io/client-go/kubernetes/fake"
)

func TestCreateDeployment(t *testing.T) {
	namespace := "test1"
	name := "mock-deployment"
	internalVNFID := "1"
	testCases := []struct {
		label          string
		input          *krd.ResourceData
		clientOutput   *appsV1.Deployment
		expectedResult string
		expectedError  string
	}{
		{
			label: "Fail to create a deployment with non-existing file",
			input: &krd.ResourceData{
				YamlFilePath: "non-existing_test_file.yaml",
			},
			clientOutput:  &appsV1.Deployment{},
			expectedError: "not found",
		},
		{
			label: "Fail to create a deployment with invalid type",
			input: &krd.ResourceData{
				YamlFilePath: "../../mock_files/mock_yamls/service.yaml",
			},
			clientOutput:  &appsV1.Deployment{},
			expectedError: "contains another resource different than Deployment",
		},
		{
			label: "Successfully create a deployment",
			input: &krd.ResourceData{
				VnfId:        internalVNFID,
				YamlFilePath: "../../mock_files/mock_yamls/deployment.yaml",
			},
			clientOutput: &appsV1.Deployment{
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

func TestListDeployment(t *testing.T) {
	namespace := "test1"
	testCases := []struct {
		label          string
		input          string
		clientOutput   *appsV1.DeploymentList
		expectedResult []string
	}{
		{
			label:          "Sucessfully display an empty deployment list",
			input:          namespace,
			clientOutput:   &appsV1.DeploymentList{},
			expectedResult: []string{},
		},
		{
			label: "Sucessfully display a list of existing deployments",
			input: namespace,
			clientOutput: &appsV1.DeploymentList{
				Items: []appsV1.Deployment{
					appsV1.Deployment{
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

func TestDeleteDeployment(t *testing.T) {
	namespace := "test1"
	name := "mock-deployment"
	testCases := []struct {
		label        string
		input        string
		clientOutput *appsV1.Deployment
	}{
		{
			label: "Sucessfully delete an existing deployment",
			input: name,
			clientOutput: &appsV1.Deployment{
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

func TestGetDeployment(t *testing.T) {
	namespace := "test1"
	name := "mock-deployment"
	testCases := []struct {
		label          string
		input          string
		clientOutput   *appsV1.Deployment
		expectedResult string
	}{
		{
			label: "Sucessfully get an existing deployment",
			input: name,
			clientOutput: &appsV1.Deployment{
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
