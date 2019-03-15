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
	"bytes"
	"k8splugin/internal/db"
	"reflect"
	"sort"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

func TestCreateDefinition(t *testing.T) {
	testCases := []struct {
		label         string
		inp           Definition
		expectedError string
		mockdb        *db.MockDB
		expected      Definition
	}{
		{
			label: "Create Resource Bundle Definition",
			inp: Definition{
				Name:        "testresourcebundle",
				Version:     "v1",
				Description: "testresourcebundle",
				ChartName:   "",
			},
			expected: Definition{
				Name:        "testresourcebundle",
				Version:     "v1",
				Description: "testresourcebundle",
				ChartName:   "",
			},
			expectedError: "",
			mockdb:        &db.MockDB{},
		},
		{
			label:         "Failed Create Resource Bundle Definition",
			expectedError: "Error Creating Definition",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Definition"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewDefinitionClient()
			got, err := impl.Create(testCase.inp)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Create Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestListDefinition(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		expected      []Definition
	}{
		{
			label: "List Resource Bundle Definition",
			name:  "testresourcebundle",
			expected: []Definition{
				{
					Name:        "testresourcebundle",
					Version:     "v1",
					Description: "testresourcebundle",
					ChartName:   "testchart",
				},
				{
					Name:        "testresourcebundle",
					Version:     "v2",
					Description: "testresourcebundle_version2",
					ChartName:   "testchart",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"testchart\"}"),
					},
					definitionKey{Name: "testresourcebundle", Version: "v2"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle_version2\"," +
								"\"rb-version\":\"v2\"," +
								"\"chart-name\":\"testchart\"}"),
					},
				},
			},
		},
		{
			label:         "List Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewDefinitionClient()
			got, err := impl.List(testCase.name)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("List returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("List returned an unexpected error %s", err)
				}
			} else {
				// Since the order of returned slice is not guaranteed
				// Check both and return error if both don't match
				sort.Slice(got, func(i, j int) bool {
					return got[i].Version < got[j].Version
				})
				// Sort both as it is not expected that testCase.expected
				// is sorted
				sort.Slice(testCase.expected, func(i, j int) bool {
					return testCase.expected[i].Version < testCase.expected[j].Version
				})

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("List Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetDefinition(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		version       string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      Definition
	}{
		{
			label:   "Get Resource Bundle Definition",
			name:    "testresourcebundle",
			version: "v1",
			expected: Definition{
				Name:        "testresourcebundle",
				Version:     "v1",
				Description: "testresourcebundle",
				ChartName:   "testchart",
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"testchart\"}"),
					},
				},
			},
		},
		{
			label:         "Get Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewDefinitionClient()
			got, err := impl.Get(testCase.name, testCase.version)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Get returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteDefinition(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		version       string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:   "Delete Resource Bundle Definition",
			name:    "testresourcebundle",
			version: "v1",
			mockdb:  &db.MockDB{},
		},
		{
			label:         "Delete Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewDefinitionClient()
			err := impl.Delete(testCase.name, testCase.version)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Delete returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Delete returned an unexpected error %s", err)
				}
			}
		})
	}
}

func TestUploadDefinition(t *testing.T) {
	testCases := []struct {
		label         string
		name, version string
		content       []byte
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:   "Upload With Chart Name Detection",
			name:    "testresourcebundle",
			version: "v1",
			//Binary format for testchart/Chart.yaml
			content: []byte{
				0x1f, 0x8b, 0x08, 0x08, 0xb3, 0xeb, 0x86, 0x5c,
				0x00, 0x03, 0x74, 0x65, 0x73, 0x74, 0x63, 0x68,
				0x61, 0x72, 0x74, 0x2e, 0x74, 0x61, 0x72, 0x00,
				0xed, 0xd2, 0x41, 0x4b, 0xc3, 0x30, 0x18, 0xc6,
				0xf1, 0x9c, 0xfb, 0x29, 0xde, 0x4f, 0x50, 0x93,
				0x36, 0x69, 0x60, 0x37, 0xd9, 0x45, 0xf0, 0xee,
				0x55, 0xe2, 0x16, 0xb1, 0x74, 0xed, 0x46, 0x9a,
				0x0d, 0xfc, 0xf6, 0xae, 0x83, 0x4d, 0x91, 0x89,
				0x97, 0x0d, 0x91, 0xfd, 0x7f, 0x87, 0x84, 0x90,
				0x90, 0xbc, 0xe1, 0x79, 0x73, 0x1c, 0xf3, 0xe2,
				0x2d, 0xa4, 0x7c, 0xa7, 0xae, 0x46, 0xef, 0x79,
				0xef, 0xa6, 0xd9, 0x78, 0xa7, 0xbf, 0xce, 0x47,
				0xca, 0xd4, 0xd6, 0x1a, 0xd7, 0xb8, 0xa6, 0xb6,
				0x4a, 0x1b, 0x5b, 0xbb, 0x4a, 0x89, 0xbb, 0x5e,
				0x49, 0x9f, 0xb6, 0x63, 0x0e, 0x49, 0x44, 0x85,
				0xe5, 0x73, 0xd7, 0x75, 0xa1, 0x6f, 0x87, 0x78,
				0xf6, 0xdc, 0x6f, 0xfb, 0xff, 0x54, 0x3e, 0xe5,
				0x3f, 0x9f, 0xc6, 0xf2, 0x3d, 0xf4, 0xab, 0x4b,
				0xbf, 0x31, 0x05, 0xdc, 0x34, 0xf6, 0xc7, 0xfc,
				0x4d, 0xe5, 0xbf, 0xe5, 0xdf, 0x54, 0xde, 0x2b,
				0xd1, 0x97, 0x2e, 0xe4, 0x9c, 0x1b, 0xcf, 0x3f,
				0x6c, 0xda, 0xa7, 0x98, 0xc6, 0x76, 0x3d, 0xcc,
				0x64, 0x67, 0x8a, 0x65, 0x1c, 0x17, 0xa9, 0xdd,
				0xe4, 0xc3, 0xfa, 0x5e, 0x1e, 0xe2, 0xaa, 0x97,
				0x43, 0x7b, 0xc8, 0xeb, 0x3a, 0xc9, 0xe3, 0xf6,
				0x25, 0xa6, 0x21, 0xee, 0x7b, 0xa6, 0x18, 0x42,
				0x1f, 0x67, 0x72, 0xea, 0x9e, 0x62, 0x77, 0xbc,
				0x44, 0x97, 0xa6, 0xd4, 0xc5, 0x5f, 0x7f, 0x0b,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0xb8, 0x09, 0x1f, 0xae,
				0x48, 0xfe, 0xe8, 0x00, 0x28, 0x00, 0x00,
			},
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"}"),
					},
				},
			},
		},
		{
			label:   "Upload With Chart Name",
			name:    "testresourcebundle",
			version: "v1",
			content: []byte{
				0x1f, 0x8b, 0x08, 0x08, 0xb0, 0x6b, 0xf4, 0x5b,
				0x00, 0x03, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x74,
				0x61, 0x72, 0x00, 0xed, 0xce, 0x41, 0x0a, 0xc2,
				0x30, 0x10, 0x85, 0xe1, 0xac, 0x3d, 0x45, 0x4e,
				0x50, 0x12, 0xd2, 0xc4, 0xe3, 0x48, 0xa0, 0x01,
				0x4b, 0x52, 0x0b, 0xed, 0x88, 0x1e, 0xdf, 0x48,
				0x11, 0x5c, 0x08, 0xa5, 0x8b, 0x52, 0x84, 0xff,
				0xdb, 0xbc, 0x61, 0x66, 0x16, 0x4f, 0xd2, 0x2c,
				0x8d, 0x3c, 0x45, 0xed, 0xc8, 0x54, 0x21, 0xb4,
				0xef, 0xb4, 0x67, 0x6f, 0xbe, 0x73, 0x61, 0x9d,
				0xb2, 0xce, 0xd5, 0x55, 0xf0, 0xde, 0xd7, 0x3f,
				0xdb, 0xd6, 0x49, 0x69, 0xb3, 0x67, 0xa9, 0x8f,
				0xfb, 0x2c, 0x71, 0xd2, 0x5a, 0xc5, 0xee, 0x92,
				0x73, 0x8e, 0x43, 0x7f, 0x4b, 0x3f, 0xff, 0xd6,
				0xee, 0x7f, 0xea, 0x9a, 0x4a, 0x19, 0x1f, 0xe3,
				0x54, 0xba, 0xd3, 0xd1, 0x55, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x1b, 0xbc, 0x00, 0xb5, 0xe8,
				0x4a, 0xf9, 0x00, 0x28, 0x00, 0x00,
			},
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"firewall\"}"),
					},
				},
			},
		},
		{
			label:         "Upload Without Chart.yaml",
			name:          "testresourcebundle",
			version:       "v1",
			expectedError: "Unable to detect chart name",
			content: []byte{
				0x1f, 0x8b, 0x08, 0x08, 0xb0, 0x6b, 0xf4, 0x5b,
				0x00, 0x03, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x74,
				0x61, 0x72, 0x00, 0xed, 0xce, 0x41, 0x0a, 0xc2,
				0x30, 0x10, 0x85, 0xe1, 0xac, 0x3d, 0x45, 0x4e,
				0x50, 0x12, 0xd2, 0xc4, 0xe3, 0x48, 0xa0, 0x01,
				0x4b, 0x52, 0x0b, 0xed, 0x88, 0x1e, 0xdf, 0x48,
				0x11, 0x5c, 0x08, 0xa5, 0x8b, 0x52, 0x84, 0xff,
				0xdb, 0xbc, 0x61, 0x66, 0x16, 0x4f, 0xd2, 0x2c,
				0x8d, 0x3c, 0x45, 0xed, 0xc8, 0x54, 0x21, 0xb4,
				0xef, 0xb4, 0x67, 0x6f, 0xbe, 0x73, 0x61, 0x9d,
				0xb2, 0xce, 0xd5, 0x55, 0xf0, 0xde, 0xd7, 0x3f,
				0xdb, 0xd6, 0x49, 0x69, 0xb3, 0x67, 0xa9, 0x8f,
				0xfb, 0x2c, 0x71, 0xd2, 0x5a, 0xc5, 0xee, 0x92,
				0x73, 0x8e, 0x43, 0x7f, 0x4b, 0x3f, 0xff, 0xd6,
				0xee, 0x7f, 0xea, 0x9a, 0x4a, 0x19, 0x1f, 0xe3,
				0x54, 0xba, 0xd3, 0xd1, 0x55, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x1b, 0xbc, 0x00, 0xb5, 0xe8,
				0x4a, 0xf9, 0x00, 0x28, 0x00, 0x00,
			},
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"firewall\"}"),
					},
				},
			},
		},
		{
			label:         "Upload with an Invalid Resource Bundle Definition",
			name:          "testresourcebundle",
			version:       "v1",
			expectedError: "Invalid Definition ID provided",
			content: []byte{
				0x1f, 0x8b, 0x08, 0x08, 0xb0, 0x6b, 0xf4, 0x5b,
				0x00, 0x03, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x74,
				0x61, 0x72, 0x00, 0xed, 0xce, 0x41, 0x0a, 0xc2,
				0x30, 0x10, 0x85, 0xe1, 0xac, 0x3d, 0x45, 0x4e,
				0x50, 0x12, 0xd2, 0xc4, 0xe3, 0x48, 0xa0, 0x01,
				0x4b, 0x52, 0x0b, 0xed, 0x88, 0x1e, 0xdf, 0x48,
				0x11, 0x5c, 0x08, 0xa5, 0x8b, 0x52, 0x84, 0xff,
				0xdb, 0xbc, 0x61, 0x66, 0x16, 0x4f, 0xd2, 0x2c,
				0x8d, 0x3c, 0x45, 0xed, 0xc8, 0x54, 0x21, 0xb4,
				0xef, 0xb4, 0x67, 0x6f, 0xbe, 0x73, 0x61, 0x9d,
				0xb2, 0xce, 0xd5, 0x55, 0xf0, 0xde, 0xd7, 0x3f,
				0xdb, 0xd6, 0x49, 0x69, 0xb3, 0x67, 0xa9, 0x8f,
				0xfb, 0x2c, 0x71, 0xd2, 0x5a, 0xc5, 0xee, 0x92,
				0x73, 0x8e, 0x43, 0x7f, 0x4b, 0x3f, 0xff, 0xd6,
				0xee, 0x7f, 0xea, 0x9a, 0x4a, 0x19, 0x1f, 0xe3,
				0x54, 0xba, 0xd3, 0xd1, 0x55, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x1b, 0xbc, 0x00, 0xb5, 0xe8,
				0x4a, 0xf9, 0x00, 0x28, 0x00, 0x00,
			},
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"firewall\"}"),
					},
				},
			},
		},
		{
			label:         "Invalid File Format Error",
			name:          "testresourcebundle",
			version:       "v1",
			expectedError: "Error in file format",
			content: []byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			},
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"firewall\"}"),
					},
				},
			},
		},
		{
			label:         "Upload Error",
			expectedError: "DB Error",
			content: []byte{
				0x1f, 0x8b, 0x08, 0x08, 0xb0, 0x6b, 0xf4, 0x5b,
				0x00, 0x03, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x74,
				0x61, 0x72, 0x00, 0xed, 0xce, 0x41, 0x0a, 0xc2,
				0x30, 0x10, 0x85, 0xe1, 0xac, 0x3d, 0x45, 0x4e,
				0x50, 0x12, 0xd2, 0xc4, 0xe3, 0x48, 0xa0, 0x01,
				0x4b, 0x52, 0x0b, 0xed, 0x88, 0x1e, 0xdf, 0x48,
				0x11, 0x5c, 0x08, 0xa5, 0x8b, 0x52, 0x84, 0xff,
				0xdb, 0xbc, 0x61, 0x66, 0x16, 0x4f, 0xd2, 0x2c,
				0x8d, 0x3c, 0x45, 0xed, 0xc8, 0x54, 0x21, 0xb4,
				0xef, 0xb4, 0x67, 0x6f, 0xbe, 0x73, 0x61, 0x9d,
				0xb2, 0xce, 0xd5, 0x55, 0xf0, 0xde, 0xd7, 0x3f,
				0xdb, 0xd6, 0x49, 0x69, 0xb3, 0x67, 0xa9, 0x8f,
				0xfb, 0x2c, 0x71, 0xd2, 0x5a, 0xc5, 0xee, 0x92,
				0x73, 0x8e, 0x43, 0x7f, 0x4b, 0x3f, 0xff, 0xd6,
				0xee, 0x7f, 0xea, 0x9a, 0x4a, 0x19, 0x1f, 0xe3,
				0x54, 0xba, 0xd3, 0xd1, 0x55, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x1b, 0xbc, 0x00, 0xb5, 0xe8,
				0x4a, 0xf9, 0x00, 0x28, 0x00, 0x00,
			},
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewDefinitionClient()
			err := impl.Upload(testCase.name, testCase.version, testCase.content)
			if err != nil {
				if testCase.expectedError == "" {
					t.Errorf("Upload returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Errorf("Upload returned an unexpected error %s", err)
				}
			}
		})
	}
}

func TestDownloadDefinition(t *testing.T) {
	testCases := []struct {
		label         string
		name, version string
		expected      []byte
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:   "Download Resource Bundle Definition",
			name:    "testresourcebundle",
			version: "v1",
			expected: []byte{
				0x1f, 0x8b, 0x08, 0x08, 0xb0, 0x6b, 0xf4, 0x5b,
				0x00, 0x03, 0x74, 0x65, 0x73, 0x74, 0x2e, 0x74,
				0x61, 0x72, 0x00, 0xed, 0xce, 0x41, 0x0a, 0xc2,
				0x30, 0x10, 0x85, 0xe1, 0xac, 0x3d, 0x45, 0x4e,
				0x50, 0x12, 0xd2, 0xc4, 0xe3, 0x48, 0xa0, 0x01,
				0x4b, 0x52, 0x0b, 0xed, 0x88, 0x1e, 0xdf, 0x48,
				0x11, 0x5c, 0x08, 0xa5, 0x8b, 0x52, 0x84, 0xff,
				0xdb, 0xbc, 0x61, 0x66, 0x16, 0x4f, 0xd2, 0x2c,
				0x8d, 0x3c, 0x45, 0xed, 0xc8, 0x54, 0x21, 0xb4,
				0xef, 0xb4, 0x67, 0x6f, 0xbe, 0x73, 0x61, 0x9d,
				0xb2, 0xce, 0xd5, 0x55, 0xf0, 0xde, 0xd7, 0x3f,
				0xdb, 0xd6, 0x49, 0x69, 0xb3, 0x67, 0xa9, 0x8f,
				0xfb, 0x2c, 0x71, 0xd2, 0x5a, 0xc5, 0xee, 0x92,
				0x73, 0x8e, 0x43, 0x7f, 0x4b, 0x3f, 0xff, 0xd6,
				0xee, 0x7f, 0xea, 0x9a, 0x4a, 0x19, 0x1f, 0xe3,
				0x54, 0xba, 0xd3, 0xd1, 0x55, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x1b, 0xbc, 0x00, 0xb5, 0xe8,
				0x4a, 0xf9, 0x00, 0x28, 0x00, 0x00,
			},
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"firewall\"}"),
						"content": []byte("H4sICLBr9FsAA3Rlc3QudGFyAO3OQQrCMBCF4aw9RU5" +
							"QEtLE40igAUtSC+2IHt9IEVwIpYtShP/bvGFmFk/SLI08Re3IVCG077Rn" +
							"b75zYZ2yztVV8N7XP9vWSWmzZ6mP+yxx0lrF7pJzjkN/Sz//1u5/6ppKG" +
							"R/jVLrT0VUAAAAAAAAAAAAAAAAAABu8ALXoSvkAKAAA"),
					},
				},
			},
		},
		{
			label:         "Download with an Invalid Resource Bundle Definition",
			name:          "testresourcebundle",
			version:       "v2",
			expectedError: "Invalid Definition ID provided",
			mockdb: &db.MockDB{
				Items: map[db.DBKey]map[string][]byte{
					definitionKey{Name: "testresourcebundle", Version: "v1"}: {
						"metadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"firewall\"}"),
					},
				},
			},
		},
		{
			label:         "Download Error",
			expectedError: "DB Error",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewDefinitionClient()
			data, err := impl.Download(testCase.name, testCase.version)
			if err != nil {
				if testCase.expectedError == "" {
					t.Errorf("Download returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Errorf("Download returned an unexpected error %s", err)
				}
			} else {
				if bytes.Equal(testCase.expected, data) == false {
					t.Errorf("Download returned unexpected data: got %v - expected %v",
						data, testCase.expected)
				}
			}
		})
	}
}
