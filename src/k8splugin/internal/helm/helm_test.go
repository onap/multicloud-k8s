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

package helm

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestProcessValues(t *testing.T) {

	chartDir := "../../mock_files/mock_charts/testchart2"
	profileDir := "../../mock_files/mock_profiles/profile1"

	testCases := []struct {
		label         string
		valueFiles    []string
		values        []string
		expectedHash  string
		expectedError string
	}{
		{
			label: "Process Values with Value Files Override",
			valueFiles: []string{
				filepath.Join(chartDir, "values.yaml"),
				filepath.Join(profileDir, "override_values.yaml"),
			},
			//Hash of a combined values.yaml file that is expected
			expectedHash:  "c18a70f426933de3c051c996dc34fd537d0131b2d13a2112a2ecff674db6c2f9",
			expectedError: "",
		},
		{
			label: "Process Values with Values Pair Override",
			valueFiles: []string{
				filepath.Join(chartDir, "values.yaml"),
			},
			//Use the same convention as specified in helm template --set
			values: []string{
				"service.externalPort=82",
			},
			//Hash of a combined values.yaml file that is expected
			expectedHash:  "028a3521fc9f8777ea7e67a6de0c51f2c875b88ca91734999657f0ca924ddb7a",
			expectedError: "",
		},
		{
			label: "Process Values with Both Overrides",
			valueFiles: []string{
				filepath.Join(chartDir, "values.yaml"),
				filepath.Join(profileDir, "override_values.yaml"),
			},
			//Use the same convention as specified in helm template --set
			//Key takes precedence over the value from override_values.yaml
			values: []string{
				"service.externalPort=82",
			},
			//Hash of a combined values.yaml file that is expected
			expectedHash:  "516fab4ab7b76ba2ff35a97c2a79b74302543f532857b945f2fe25e717e755be",
			expectedError: "",
		},
	}

	h := sha256.New()

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			tc := NewTemplateClient("1.12.3", "testnamespace", "testreleasename")
			out, err := tc.processValues(testCase.valueFiles, testCase.values)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got an error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}
			} else {
				//Compute the hash of returned data and compare
				h.Write(out)
				gotHash := fmt.Sprintf("%x", h.Sum(nil))
				h.Reset()
				if gotHash != testCase.expectedHash {
					t.Fatalf("Got unexpected values.yaml %s", out)
				}
			}
		})
	}
}

func TestGenerateKubernetesArtifacts(t *testing.T) {

	chartDir := "../../mock_files/mock_charts/testchart2"
	profileDir := "../../mock_files/mock_profiles/profile1"

	testCases := []struct {
		label           string
		chartPath       string
		valueFiles      []string
		values          []string
		expectedHashMap map[string]string
		expectedError   string
	}{
		{
			label:      "Generate artifacts without any overrides",
			chartPath:  chartDir,
			valueFiles: []string{},
			values:     []string{},
			//sha256 hash of the evaluated templates in each chart
			expectedHashMap: map[string]string{
				"testchart2/templates/service.yaml": "fdd6a2b6795486f0dd1d8c44379afb5ffe4072c09f9cf6594738e8ded4dd872d",
				"subcharta/templates/service.yaml":  "570389588fffdb7193ab265888d781f3d751f3a40362533344f9aa7bb93a8bb0",
				"subchartb/templates/service.yaml":  "5654e03d922e8ec49649b4bbda9dfc9e643b3b7c9c18b602cc7e26fd36a39c2a",
			},
			expectedError: "",
		},
		{
			label:     "Generate artifacts with overrides",
			chartPath: chartDir,
			valueFiles: []string{
				filepath.Join(profileDir, "override_values.yaml"),
			},
			values: []string{
				"service.externalPort=82",
			},
			//sha256 hash of the evaluated templates in each chart
			expectedHashMap: map[string]string{
				"testchart2/templates/service.yaml": "2bb96e791ecb6a3404bc5de3f6c4182aed881630269e2aa6766df38b0f852724",
				"subcharta/templates/service.yaml":  "570389588fffdb7193ab265888d781f3d751f3a40362533344f9aa7bb93a8bb0",
				"subchartb/templates/service.yaml":  "5654e03d922e8ec49649b4bbda9dfc9e643b3b7c9c18b602cc7e26fd36a39c2a",
			},
			expectedError: "",
		},
	}

	h := sha256.New()

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			tc := NewTemplateClient("1.12.3", "testnamespace", "testreleasename")
			out, err := tc.GenerateKubernetesArtifacts(testCase.chartPath, testCase.valueFiles,
				testCase.values)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got an error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}
			} else {
				//Compute the hash of returned data and compare
				for _, v := range out {
					f := v.FilePath
					data, err := ioutil.ReadFile(f)
					if err != nil {
						t.Errorf("Unable to read file %s", v)
					}
					h.Write(data)
					gotHash := fmt.Sprintf("%x", h.Sum(nil))
					h.Reset()

					//Find the right hash from expectedHashMap
					expectedHash := ""
					for k1, v1 := range testCase.expectedHashMap {
						if strings.Contains(f, k1) == true {
							expectedHash = v1
							break
						}
					}
					if gotHash != expectedHash {
						t.Fatalf("Got unexpected hash for %s", f)
					}
				}
			}
		})
	}
}
