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
	"k8splugin/db"
	"reflect"
	"sort"
	"strings"
	"testing"

	pkgerrors "github.com/pkg/errors"
)

func TestCreateProfile(t *testing.T) {
	testCases := []struct {
		label         string
		inp           Profile
		expectedError string
		mockdb        *db.MockDB
		expected      Profile
	}{
		{
			label: "Create Resource Bundle Profile",
			inp: Profile{
				UUID:              "123e4567-e89b-12d3-a456-426655440000",
				RBDID:             "abcde123-e89b-8888-a456-986655447236",
				Name:              "testresourcebundle",
				Namespace:         "default",
				KubernetesVersion: "1.12.3",
			},
			expected: Profile{
				UUID:              "123e4567-e89b-12d3-a456-426655440000",
				RBDID:             "abcde123-e89b-8888-a456-986655447236",
				Name:              "testresourcebundle",
				Namespace:         "default",
				KubernetesVersion: "1.12.3",
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					"abcde123-e89b-8888-a456-986655447236": {
						"metadata": []byte(
							"{\"name\":\"testresourcebundle\"," +
								"\"namespace\":\"default\"," +
								"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
					},
				},
			},
		},
		{
			label:         "Failed Create Resource Bundle Profile",
			expectedError: "Error Creating Profile",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Profile"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewProfileClient()
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

func TestListProfiles(t *testing.T) {

	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		expected      []Profile
	}{
		{
			label: "List Resource Bundle Profile",
			expected: []Profile{
				{
					UUID:              "123e4567-e89b-12d3-a456-426655440000",
					RBDID:             "abcde123-e89b-8888-a456-986655447236",
					Name:              "testresourcebundle",
					Namespace:         "default",
					KubernetesVersion: "1.12.3",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					"123e4567-e89b-12d3-a456-426655440000": {
						"metadata": []byte(
							"{\"name\":\"testresourcebundle\"," +
								"\"namespace\":\"default\"," +
								"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
								"\"rbdid\":\"abcde123-e89b-8888-a456-986655447236\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
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
			impl := NewProfileClient()
			got, err := impl.List()
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
					return got[i].UUID < got[j].UUID
				})
				// Sort both as it is not expected that testCase.expected
				// is sorted
				sort.Slice(testCase.expected, func(i, j int) bool {
					return testCase.expected[i].UUID < testCase.expected[j].UUID
				})

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("List Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetProfile(t *testing.T) {

	testCases := []struct {
		label         string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      Profile
	}{
		{
			label: "Get Resource Bundle Profile",
			inp:   "123e4567-e89b-12d3-a456-426655440000",
			expected: Profile{
				UUID:              "123e4567-e89b-12d3-a456-426655440000",
				RBDID:             "abcde123-e89b-8888-a456-986655447236",
				Name:              "testresourcebundle",
				Namespace:         "default",
				KubernetesVersion: "1.12.3",
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					"123e4567-e89b-12d3-a456-426655440000": {
						"metadata": []byte(
							"{\"name\":\"testresourcebundle\"," +
								"\"namespace\":\"default\"," +
								"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
								"\"rbdid\":\"abcde123-e89b-8888-a456-986655447236\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
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
			impl := NewProfileClient()
			got, err := impl.Get(testCase.inp)
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

func TestDeleteProfile(t *testing.T) {

	testCases := []struct {
		label         string
		inp           string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:  "Delete Resource Bundle Profile",
			inp:    "123e4567-e89b-12d3-a456-426655440000",
			mockdb: &db.MockDB{},
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
			impl := NewProfileClient()
			err := impl.Delete(testCase.inp)
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

func TestUploadProfile(t *testing.T) {
	testCases := []struct {
		label         string
		inp           string
		content       []byte
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label: "Upload Resource Bundle Profile",
			inp:   "123e4567-e89b-12d3-a456-426655440000",
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
				Items: map[string]map[string][]byte{
					"123e4567-e89b-12d3-a456-426655440000": {
						"metadata": []byte(
							"{\"name\":\"testresourcebundle\"," +
								"\"namespace\":\"default\"," +
								"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
								"\"rbdid\":\"abcde123-e89b-8888-a456-986655447236\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
					},
				},
			},
		},
		{
			label:         "Upload with an Invalid Resource Bundle Profile",
			inp:           "123e4567-e89b-12d3-a456-426655440000",
			expectedError: "Invalid Profile ID provided",
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
				Items: map[string]map[string][]byte{
					"123e4567-e89b-12d3-a456-426655441111": {
						"metadata": []byte(
							"{\"name\":\"testresourcebundle\"," +
								"\"namespace\":\"default\"," +
								"\"uuid\":\"123e4567-e89b-12d3-a456-426655441111\"," +
								"\"rbdid\":\"abcde123-e89b-8888-a456-986655447236\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
					},
				},
			},
		},
		{
			label:         "Invalid File Format Error",
			inp:           "123e4567-e89b-12d3-a456-426655440000",
			expectedError: "Error in file format",
			content: []byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					"123e4567-e89b-12d3-a456-426655440000": {
						"metadata": []byte(
							"{\"name\":\"testresourcebundle\"," +
								"\"namespace\":\"default\"," +
								"\"uuid\":\"123e4567-e89b-12d3-a456-426655440000\"," +
								"\"rbdid\":\"abcde123-e89b-8888-a456-986655447236\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
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
			impl := NewProfileClient()
			err := impl.Upload(testCase.inp, testCase.content)
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
