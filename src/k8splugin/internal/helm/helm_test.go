/*
 * Copyright 2018 Intel Corporation, Inc
 * Copyright 2020,2021 Samsung Electronics, Modifications
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

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
			expectedHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
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
			expectedHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
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
			expectedHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			expectedError: "",
		},
		{
			label: "Process complex Pair Override",
			values: []string{
				"name={a,b,c}",
				"servers[0].port=80",
			},
			expectedError: "",
			expectedHash:  "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
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
				gotHash := fmt.Sprintf("%x", h.Sum(nil))
				h.Reset()
				if gotHash != testCase.expectedHash {
					mout, _ := yaml.Marshal(&out)
					t.Fatalf("Got unexpected hash '%s' of values.yaml:\n%v", gotHash, string(mout))
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
				"manifest-0": "fcc1083ace82b633e3a0a687d50f532c07e1212b7a42b2c178b65e5768fffcfe",
				"manifest-2": "eefeac6ff5430a16a32ae3974857cbe5ff516a1a68566e5edcddd410d60397c0",
				"manifest-1": "b88aa963ee3afb9676e9930519d7caa103df1251da48a9351ab4ac0c5730d2af",
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
				"manifest-0": "fcc1083ace82b633e3a0a687d50f532c07e1212b7a42b2c178b65e5768fffcfe",
				"manifest-2": "03ae530e49071d005be78f581b7c06c59119f91f572b28c0c0c06ced8e37bf6e",
				"manifest-1": "b88aa963ee3afb9676e9930519d7caa103df1251da48a9351ab4ac0c5730d2af",
			},
			expectedError: "",
		},
		{
			label:      "Generate artifacts from multi-template and empty files v1",
			chartPath:  "../../mock_files/mock_charts/testchart3",
			valueFiles: []string{},
			values: []string{
				"goingEmpty=false",
			},
			expectedHashMap: map[string]string{
				"manifest-0": "666e8d114981a4b5d13fb799be060aa57e0e48904bba4a410f87a2e827a57ddb",
				"manifest-2": "6a5af22538c273b9d4a3156e3b6bb538c655041eae31e93db21a9e178f73ecf0",
			},
			expectedError: "",
		},
		{
			label:      "Generate artifacts from multi-template and empty files v2",
			chartPath:  "../../mock_files/mock_charts/testchart3",
			valueFiles: []string{},
			values: []string{
				"goingEmpty=true",
			},
			expectedHashMap: map[string]string{
				"manifest-0": "666e8d114981a4b5d13fb799be060aa57e0e48904bba4a410f87a2e827a57ddb",
				"manifest-1": "8613e7e7cc0186516b13be37ec7fc321ff89e3abaed0a841773a4eba2d77ce2a",
				"manifest-2": "3543ae9563fe62ce4a7446d72e1cd23140d8cc5495f0221430d70e94845c1408",
			},
			expectedError: "",
		},
		{
			label:         "Test simple v3 helm charts support",
			chartPath:     "../../mock_files/mock_charts/mockv3",
			valueFiles:    []string{},
			values:        []string{},
			expectedError: "",
			expectedHashMap: map[string]string{
				"manifest-0": "94975ff704b9cc00a7988fe7fc865665495655ec2584d3e9de2f7e5294c7eb0d",
				"dummy-test": "b50bb5f818fe0be332f09401104ae9cea59442e2dabe1a16b4ce21b753177a80",
			},
		},
	}

	h := sha256.New()

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			tc := NewTemplateClient("1.12.3", "testnamespace", "testreleasename")
			out, _, hooks, err := tc.GenerateKubernetesArtifacts(testCase.chartPath, testCase.valueFiles,
				testCase.values)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Got an error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Got unexpected error message %s", err)
				}
			} else if len(testCase.expectedHashMap) != len(out)+len(hooks) {
				t.Fatalf("Mismatch of expected files (%d) and returned resources (%d)",
					len(testCase.expectedHashMap), len(out)+len(hooks))
			} else {
				//Compute the hash of returned data and compare
				for _, v := range out {
					f := v.FilePath
					data, err := ioutil.ReadFile(f)
					if err != nil {
						t.Fatalf("Unable to read file %s", v)
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
						t.Fatalf("Got unexpected hash for %s: '%s'; expected: '%s'", f, gotHash, expectedHash)
					}
				}
				for _, v := range hooks {
					f := v.KRT.FilePath
					data, err := ioutil.ReadFile(f)
					if err != nil {
						t.Fatalf("Unable to read file %+v", v)
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
						t.Fatalf("Got unexpected hash for %s: '%s'; expected: '%s'", f, gotHash, expectedHash)
					}
				}
			}
		})
	}
}

func TestReverseResources(t *testing.T) {

	t.Run("Successfully reverse resources", func(t *testing.T) {
		data := []KubernetesResource{
			{
				GVK: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment"},
				Name: "deployment-1",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "apps",
					Version: "v1",
					Kind:    "Deployment"},
				Name: "deployment-2",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Service"},
				Name: "service-1",
			},
			{
				GVK: schema.GroupVersionKind{
					Group:   "",
					Version: "v1",
					Kind:    "Service"},
				Name: "service-2",
			},
		}

		reversed := GetReverseK8sResources(data)

		if reversed[0] != data[len(data)-1] {
			t.Fatalf("Unexpected k8s resource at position 0 %s", reversed[0])
		}
	})
}
