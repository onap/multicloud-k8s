// +build unit

/*
 * Copyright 2018 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rb

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func generateTestData() (string, string, string, error) {

	testDir, err := ioutil.TempDir("", "k8splugin-test-pr-")
	if err != nil {
		return "", "", "", err
	}

	//Create sample override_values.yaml
	valueOverrideFile, err := ioutil.TempFile(testDir, "values-*.yaml")
	if err != nil {
		return "", "", "", err
	}

	//Create a folder structure for override yaml files
	err = os.MkdirAll(filepath.Join(testDir, "testfol", "subdir"), 0755)
	if err != nil {
		return "", "", "", err
	}

	ov1, err := ioutil.TempFile(filepath.Join(testDir, "testfol", "subdir"), "ov1-*")
	if err != nil {
		return "", "", "", err
	}

	//Relative path
	ov1path := filepath.Join("testfol", "subdir", filepath.Base(ov1.Name()))

	//Create a dummy chart folder
	testChDir, err := ioutil.TempDir("", "k8splugin-test-chart-")
	if err != nil {
		return "", "", "", err
	}

	ch1path := "dest.yaml"

	//Create the manifest file
	manifestFile, err := ioutil.TempFile(testDir, "manifest-*.yaml")
	if err != nil {
		return "", "", "", err
	}

	manifestContent := []byte("---\n" +
		"version: v1\n" +
		"type:\n" +
		"  values: " + valueOverrideFile.Name() + "\n" +
		"  configresource:\n" +
		"    - filepath: " + ov1path + "\n" +
		"      chartpath: " + ch1path + "\n")

	_, err = manifestFile.Write(manifestContent)
	if err != nil {
		return "", "", "", err
	}

	return testDir, testChDir, filepath.Base(manifestFile.Name()), nil
}

func TestCopyConfigurationOverrides(t *testing.T) {

	testDir, chDir, manifest, err := generateTestData()
	if err != nil {
		t.Errorf("Error creating testdata %s", err)
	}

	defer os.RemoveAll(testDir)
	defer os.RemoveAll(chDir)

	testCases := []struct {
		label                  string
		prDir, chDir, manifest string
		expectedError          string
	}{
		{
			label:         "Process Profile Yaml",
			prDir:         testDir,
			chDir:         chDir,
			manifest:      manifest,
			expectedError: "",
		},
		{
			label:         "Non existent manifest file",
			prDir:         testDir,
			chDir:         chDir,
			manifest:      "non-existant-file.yaml",
			expectedError: "Reading manifest file",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_, err = ProcessProfileYaml(testCase.prDir, testCase.manifest)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got an error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}
			}
		})
	}

	copyConfiguration_testCases := []struct {
		label                  string
		prDir, chDir, manifest string
		expectedError          string
	}{
		{
			label:         "Copy Configuration Override Files",
			prDir:         testDir,
			chDir:         chDir,
			manifest:      manifest,
			expectedError: "",
		},
	}

	for _, testCase := range copyConfiguration_testCases {
		t.Run(testCase.label, func(t *testing.T) {
			p, err := ProcessProfileYaml(testCase.prDir, testCase.manifest)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got an error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}
			}

			err = p.CopyConfigurationOverrides(chDir)
			if err != nil {
				t.Fatalf("CopyConfigurationOverrides returned error %s", err)
			}
		})
	}
}
