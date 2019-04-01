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
	pkgerrors "github.com/pkg/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8splugin/plugins/network/v1"
	"reflect"
	"strings"
	"testing"
)

type mockOVNCmd struct {
	StdOut string
	StdErr string
	Err    error
}

func (cmd *mockOVNCmd) Run(args ...string) (string, string, error) {
	return cmd.StdOut, cmd.StdErr, cmd.Err
}

func TestCreateOVN4NFVK8SNetwork(t *testing.T) {
	testCases := []struct {
		label          string
		input          *v1.OnapNetwork
		mock           *mockOVNCmd
		expectedResult string
		expectedError  string
	}{
		{
			label:         "Fail to decode a network",
			input:         &v1.OnapNetwork{},
			expectedError: "Invalid configuration value",
		},
		{
			label: "Fail to create a network",
			input: &v1.OnapNetwork{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test",
				},
				Spec: v1.OnapNetworkSpec{
					Config: "{\"cnitype\": \"ovn4nfvk8s\",\"name\": \"mynet\",\"subnet\": \"172.16.33.0/24\",\"gateway\": \"172.16.33.1\",\"routes\": [{\"dst\": \"172.16.29.1/24\",\"gw\": \"100.64.1.1\"}]}",
				},
			},
			expectedError: "Failed to get logical router",
			mock: &mockOVNCmd{
				Err: pkgerrors.New("Internal error"),
			},
		},
		{
			label: "Successfully create a ovn4nfv network",
			input: &v1.OnapNetwork{
				ObjectMeta: metaV1.ObjectMeta{
					Name: "test",
				},
				Spec: v1.OnapNetworkSpec{
					Config: "{\"cnitype\": \"ovn4nfvk8s\",\"name\": \"mynet\",\"subnet\": \"172.16.33.0/24\",\"gateway\": \"172.16.33.1\",\"routes\": [{\"dst\": \"172.16.29.1/24\",\"gw\": \"100.64.1.1\"}]}",
				},
			},
			expectedResult: "mynet",
			mock:           &mockOVNCmd{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			if testCase.mock != nil {
				ovnCmd = testCase.mock
			}
			result, err := CreateNetwork(testCase.input)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("CreateNetwork method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("CreateNetwork method returned an error (%s)", err)
				}
			} else {
				if testCase.expectedError != "" && testCase.expectedResult == "" {
					t.Fatalf("CreateNetwork method was expecting \"%s\" error message", testCase.expectedError)
				}
				if result == "" {
					t.Fatal("CreateNetwork method returned nil result")
				}
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("CreateNetwork method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}

func TestDeleteOVN4NFVK8SNetwork(t *testing.T) {
	testCases := []struct {
		label         string
		input         string
		mock          *mockOVNCmd
		expectedError string
	}{
		{
			label:         "Fail to delete a network",
			input:         "test",
			expectedError: "Failed to delete switch test",
			mock: &mockOVNCmd{
				Err: pkgerrors.New("Internal error"),
			},
		},
		{
			label: "Successfully delete a ovn4nfv network",
			input: "test",
			mock:  &mockOVNCmd{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			if testCase.mock != nil {
				ovnCmd = testCase.mock
			}
			err := DeleteNetwork(testCase.input)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("DeleteNetwork method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("DeleteNetwork method returned an error (%s)", err)
				}
			}
		})
	}
}
