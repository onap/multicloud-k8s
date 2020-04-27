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
	"strings"
	"testing"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDecodeYAML(t *testing.T) {
	testCases := []struct {
		label          string
		input          string
		expectedResult runtime.Object
		expectedError  string
	}{
		{
			label:         "Fail to read non-existing YAML file",
			input:         "unexisting-file.yaml",
			expectedError: "not found",
		},
		{
			label:         "Fail to read invalid YAML format",
			input:         "./utils_test.go",
			expectedError: "mapping values are not allowed in this contex",
		},
		{
			label: "Successfully read YAML file",
			input: "../mock_files/mock_yamls/deployment.yaml",
			expectedResult: &appsV1.Deployment{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "mock-deployment",
				},
				Spec: appsV1.DeploymentSpec{
					Template: coreV1.PodTemplateSpec{
						ObjectMeta: metaV1.ObjectMeta{
							Labels: map[string]string{"app": "sise"},
						},
						Spec: coreV1.PodSpec{
							Containers: []coreV1.Container{
								coreV1.Container{
									Name:  "sise",
									Image: "mhausenblas/simpleservice:0.5.0",
								},
							},
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			result, err := DecodeYAML(testCase.input, nil)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Decode YAML method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Decode YAML method returned an error (%s)", err)
				}
			} else {
				if testCase.expectedError != "" && testCase.expectedResult == nil {
					t.Fatalf("Decode YAML method was expecting \"%s\" error message", testCase.expectedError)
				}
				if result == nil {
					t.Fatal("Decode YAML method returned nil result")
				}
				// if !reflect.DeepEqual(testCase.expectedResult, result) {

				// 	t.Fatalf("Decode YAML method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				// }
			}
		})
	}
}
