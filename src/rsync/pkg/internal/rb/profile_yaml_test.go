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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessProfileYaml(t *testing.T) {

	profileDir := "../../mock_files/mock_profiles/profile1"
	manifestFile := "manifest.yaml"
	faultymanifestfile := "faulty-manifest.yaml"

	testCases := []struct {
		label           string
		prDir, manifest string
		expectedError   string
	}{
		{
			label:         "Process Profile Yaml",
			prDir:         profileDir,
			manifest:      manifestFile,
			expectedError: "",
		},
		{
			label:         "Non existent manifest file",
			prDir:         profileDir,
			manifest:      "non-existant-file.yaml",
			expectedError: "Reading manifest file",
		},
		{
			label:         "Faulty manifest file",
			prDir:         profileDir,
			manifest:      faultymanifestfile,
			expectedError: "Marshaling manifest yaml file",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			_, err := ProcessProfileYaml(testCase.prDir, testCase.manifest)
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
}

func TestCopyConfigurationOverrides(t *testing.T) {

	profileDir := "../../mock_files/mock_profiles/profile1"
	profileFileName := "p1.yaml"
	manifestFile := "manifest.yaml"
	faultySrcManifestFile := "faulty-src-manifest.yaml"
	faultyDestManifestFile := "faulty-dest-manifest.yaml"
	chartBaseDir := "../../mock_files/mock_charts"

	//Remove the testchart1/templates/p1.yaml file that gets copied over
	defer os.Remove(filepath.Join(chartBaseDir, "testchart1", "templates", profileFileName))

	testCases := []struct {
		label                  string
		prDir, chDir, manifest string
		expectedError          string
	}{
		{
			label:         "Copy Configuration Overrides",
			prDir:         profileDir,
			manifest:      manifestFile,
			chDir:         chartBaseDir,
			expectedError: "",
		},
		{
			label:         "Copy Configuration Overrides Faulty Source",
			prDir:         profileDir,
			manifest:      faultySrcManifestFile,
			chDir:         chartBaseDir,
			expectedError: "Reading configuration file",
		},
		{
			label:         "Copy Configuration Overrides Faulty Destination",
			prDir:         profileDir,
			manifest:      faultyDestManifestFile,
			chDir:         chartBaseDir,
			expectedError: "Writing configuration file",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			p, err := ProcessProfileYaml(testCase.prDir, testCase.manifest)
			if err != nil {
				t.Fatalf("Got unexpected error processing yaml %s", err)
			}

			err = p.CopyConfigurationOverrides(testCase.chDir)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}

			} else {
				//Check if the file got copied over
				if _, err = os.Stat(filepath.Join(testCase.chDir, "testchart1",
					"templates", profileFileName)); os.IsNotExist(err) {
					t.Fatalf("Failed to copy override file: %s", profileFileName)
				}
			}
		})
	}
}
