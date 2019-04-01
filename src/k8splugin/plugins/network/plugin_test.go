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
	utils "k8splugin/internal"
	"os"
	"plugin"
	"reflect"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

func LoadMockNetworkPlugins(krdLoadedPlugins *map[string]*plugin.Plugin, networkName, errMsg string) error {
	if _, err := os.Stat("../../mock_files/mock_plugins/mocknetworkplugin.so"); os.IsNotExist(err) {
		return pkgerrors.New("mocknetworkplugin.so does not exist. Please compile mocknetworkplugin.go to generate")
	}

	mockNetworkPlugin, err := plugin.Open("../../mock_files/mock_plugins/mocknetworkplugin.so")
	if err != nil {
		return pkgerrors.Cause(err)
	}

	symErrVar, err := mockNetworkPlugin.Lookup("Err")
	if err != nil {
		return err
	}
	symNetworkNameVar, err := mockNetworkPlugin.Lookup("NetworkName")
	if err != nil {
		return err
	}

	*symErrVar.(*string) = errMsg
	*symNetworkNameVar.(*string) = networkName
	(*krdLoadedPlugins)["ovn4nfvk8s-network"] = mockNetworkPlugin

	return nil
}

func TestCreateNetwork(t *testing.T) {
	internalVNFID := "1"
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	testCases := []struct {
		label          string
		input          *utils.ResourceData
		mockError      string
		mockOutput     string
		expectedResult string
		expectedError  string
	}{
		{
			label: "Fail to decode a network object",
			input: &utils.ResourceData{
				YamlFilePath: "../../mock_files/mock_yamls/service.yaml",
			},
			expectedError: "Fail to decode network's configuration: Invalid configuration value",
		},
		{
			label: "Fail to create a network",
			input: &utils.ResourceData{
				YamlFilePath: "../../mock_files/mock_yamls/ovn4nfvk8s.yaml",
			},
			mockError:     "Internal error",
			expectedError: "Error during the creation for ovn4nfvk8s plugin: Internal error",
		},
		{
			label: "Successfully create a ovn4nfv network",
			input: &utils.ResourceData{
				VnfId:        internalVNFID,
				YamlFilePath: "../../mock_files/mock_yamls/ovn4nfvk8s.yaml",
			},
			expectedResult: internalVNFID + "_ovn4nfvk8s_myNetwork",
			mockOutput:     "myNetwork",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			err := LoadMockNetworkPlugins(&utils.LoadedPlugins, testCase.mockOutput, testCase.mockError)
			if err != nil {
				t.Fatalf("TestCreateNetwork returned an error (%s)", err)
			}
			result, err := Create(testCase.input, nil)
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
				if !reflect.DeepEqual(testCase.expectedResult, result) {

					t.Fatalf("Create method returned: \n%v\n and it was expected: \n%v", result, testCase.expectedResult)
				}
			}
		})
	}
}

func TestDeleteNetwork(t *testing.T) {
	oldkrdPluginData := utils.LoadedPlugins

	defer func() {
		utils.LoadedPlugins = oldkrdPluginData
	}()

	testCases := []struct {
		label          string
		input          string
		mockError      string
		mockOutput     string
		expectedResult string
		expectedError  string
	}{
		{
			label:         "Fail to load non-existing plugin",
			input:         "test",
			expectedError: "No plugin for resource",
		},
		{
			label:         "Fail to delete a network",
			input:         "1_ovn4nfvk8s_test",
			mockError:     "Internal error",
			expectedError: "Error during the deletion for ovn4nfvk8s plugin: Internal error",
		},
		{
			label: "Successfully delete a ovn4nfv network",
			input: "1_ovn4nfvk8s_test",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			err := LoadMockNetworkPlugins(&utils.LoadedPlugins, testCase.mockOutput, testCase.mockError)
			if err != nil {
				t.Fatalf("TestDeleteNetwork returned an error (%s)", err)
			}
			err = Delete(testCase.input, "", nil)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create method return an un-expected (%s)", err)
				}
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Create method returned an error (%s)", err)
				}
			}
		})
	}
}
