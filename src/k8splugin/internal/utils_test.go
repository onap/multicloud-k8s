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

package utils

import (
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"net/http/httptest"
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

func TestGetESRInfo(t *testing.T) {
	testCases := []struct {
		label          string
		input          map[string]string
		mockResponse   map[string]interface{}
		expectedError  string
		expectedResult *Vim
	}{
		{
			label:         "Empty cloudOwner value failure",
			expectedError: "cloudOwner empty value",
		},
		{
			label: "Empty cloudRegionID value failure",
			input: map[string]string{
				"cloudOwner": "VIM-test1",
			},
			expectedError: "cloudRegionID empty value",
		},
		{
			label: "Fail to parse the body AAI information",
			mockResponse: map[string]interface{}{
				"code": 500,
			},
			input: map[string]string{
				"cloudOwner":    "VIM-test1",
				"cloudRegionID": "RegionOne1",
			},
			expectedError: "The AAI body response is a invalid VIM information",
		},
		{
			label: "Sucess to retrieve ESR VIM information",
			input: map[string]string{
				"cloudOwner":    "VIM-test1",
				"cloudRegionID": "RegionOne1",
			},
			mockResponse: map[string]interface{}{
				"body": []byte(`{
					"cloud-owner":"VIM-test1",
					"cloud-region-id":"RegionOne1",
					"cloud-type":"kubernetes",
					"owner-defined-type":"owner-defined-type",
					"cloud-region-version":"v1.0",
					"cloud-zone":"cloud zone",
					"complex-name":"complex name",
					"cloud-extra-info":"{}",
					"resource-version":"1548686509260"
				}`),
			},
			expectedResult: &Vim{
				CloudOwner:         "VIM-test1",
				CloudRegionID:      "RegionOne1",
				CloudType:          "kubernetes",
				OwnerDefinedType:   "owner-defined-type",
				CloudRegionVersion: "v1.0",
				CloudZone:          "cloud zone",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				if val, ok := testCase.mockResponse["code"]; ok {
					if code, ok := val.(int); ok {
						res.WriteHeader(code)
					}
				}
				if val, ok := testCase.mockResponse["body"]; ok {
					if body, ok := val.([]byte); ok {
						res.Write(body)
					}
				}
			}))
			defer func() { testServer.Close() }()
			os.Setenv("AAI_SERVICE_URL", testServer.URL)

			result, err := GetESRInfo(testCase.input["cloudOwner"], testCase.input["cloudRegionID"])
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("ESR information get method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("ESR information get method returned an error (%s)", err)
				}
			} else {
				if testCase.expectedError != "" && testCase.expectedResult == nil {
					t.Fatalf("ESR information get method was expecting \"%s\" error message", testCase.expectedError)
				}
				if result == nil {
					t.Fatal("ESR information get method returned nil result")
				}
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("ESR information get method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}

func TestGetESRAuthInfo(t *testing.T) {
	testCases := []struct {
		label          string
		input          map[string]string
		mockResponse   map[string]interface{}
		expectedError  string
		expectedResult *VimAuth
	}{
		{
			label:         "Empty cloudOwner value failure",
			expectedError: "cloudOwner empty value",
		},
		{
			label: "Empty cloudRegionID value failure",
			input: map[string]string{
				"cloudOwner": "VIM-test1",
			},
			expectedError: "cloudRegionID empty value",
		},
		{
			label: "Fail to parse the body AAI information",
			mockResponse: map[string]interface{}{
				"code": 500,
			},
			input: map[string]string{
				"cloudOwner":    "VIM-test1",
				"cloudRegionID": "RegionOne1",
			},
			expectedError: "The AAI body response is a invalid VIM authentication data",
		},
		{
			label: "Sucess to retrieve ESR Authentication information",
			input: map[string]string{
				"cloudOwner":    "VIM-test1",
				"cloudRegionID": "RegionOne1",
			},
			mockResponse: map[string]interface{}{
				"body": []byte(`{
					"esr-system-info-id":"vim-esr-system-info-id",
					"service-url":"http://10.12.25.2:5000/v3",
					"user-name":"demo",
					"password":"onapdemo",
					"system-type":"VIM",
					"ssl-insecure":true,
					"cloud-domain":"Default",
					"resource-version":"1548686509260"
					}`),
			},
			expectedResult: &VimAuth{
				ESRSystemInfoID: "vim-esr-system-info-id",
				ServiceURL:      "http://10.12.25.2:5000/v3",
				Username:        "demo",
				Password:        "onapdemo",
				SystemType:      "VIM",
				SslInsecure:     true,
				CloudDomain:     "Default",
				ResourceVersion: "1548686509260",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
				if val, ok := testCase.mockResponse["code"]; ok {
					if code, ok := val.(int); ok {
						res.WriteHeader(code)
					}
				}
				if val, ok := testCase.mockResponse["body"]; ok {
					if body, ok := val.([]byte); ok {
						res.Write(body)
					}
				}
			}))
			defer func() { testServer.Close() }()
			os.Setenv("AAI_SERVICE_URL", testServer.URL)
			os.Setenv("AAI_USERNAME", "AAI")
			os.Setenv("AAI_PASSWORD", "AAI")

			result, err := GetESRAuthInfo(testCase.input["cloudOwner"], testCase.input["cloudRegionID"])
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("The get method that retrieves ESR Authentication data return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("The get method that retrieves ESR Authentication data returned an error (%s)", err)
				}
			} else {
				if testCase.expectedError != "" && testCase.expectedResult == nil {
					t.Fatalf("The get method that retrieves ESR Authentication data was expecting \"%s\" error message", testCase.expectedError)
				}
				if result == nil {
					t.Fatal("The get method that retrieves ESR Authentication data returned nil result")
				}
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("The get method that retrieves ESR Authentication data method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}
