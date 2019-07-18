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
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"

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
				ProfileName:       "testprofile1",
				ReleaseName:       "testprofilereleasename",
				Namespace:         "testnamespace",
				KubernetesVersion: "1.12.3",
				RBName:            "testresourcebundle",
				RBVersion:         "v1",
			},
			expected: Profile{
				ProfileName:       "testprofile1",
				ReleaseName:       "testprofilereleasename",
				Namespace:         "testnamespace",
				KubernetesVersion: "1.12.3",
				RBName:            "testresourcebundle",
				RBVersion:         "v1",
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					DefinitionKey{RBName: "testresourcebundle", RBVersion: "v1"}.String(): {
						"defmetadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"testchart\"}"),
					},
				},
			},
		},
		{
			label: "Create Resource Bundle Profile With Non-Existing Definition",
			inp: Profile{
				ProfileName:       "testprofile1",
				ReleaseName:       "testprofilereleasename",
				Namespace:         "testnamespace",
				KubernetesVersion: "1.12.3",
				RBName:            "testresourcebundle",
				RBVersion:         "v1",
			},
			expectedError: "Error getting Resource Bundle Definition",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					DefinitionKey{RBName: "testresourcebundle", RBVersion: "v2"}.String(): {
						"defmetadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"description\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"testchart\"}"),
					},
				},
			},
		},
		{
			label:         "Failed Create Resource Bundle Profile",
			expectedError: "Name is required",
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

func TestGetProfile(t *testing.T) {

	testCases := []struct {
		label                     string
		rbname, rbversion, prname string
		expectedError             string
		mockdb                    *db.MockDB
		expected                  Profile
	}{
		{
			label:     "Get Resource Bundle Profile",
			rbname:    "testresourcebundle",
			rbversion: "v1",
			prname:    "testprofile1",
			expected: Profile{
				ProfileName:       "testprofile1",
				ReleaseName:       "testprofilereleasename",
				Namespace:         "testnamespace",
				KubernetesVersion: "1.12.3",
				RBName:            "testresourcebundle",
				RBVersion:         "v1",
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1", ProfileName: "testprofile1"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"testprofile1\"," +
								"\"release-name\":\"testprofilereleasename\"," +
								"\"namespace\":\"testnamespace\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"kubernetes-version\":\"1.12.3\"}"),
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
			got, err := impl.Get(testCase.rbname, testCase.rbversion, testCase.prname)
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

func TestListProfile(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		rbdef         string
		version       string
		expectedError string
		mockdb        *db.MockDB
		expected      []Profile
	}{
		{
			label:   "List Resource Bundle Profile",
			name:    "testresourcebundle",
			rbdef:   "testresourcebundle",
			version: "v1",
			expected: []Profile{
				{
					ProfileName:       "testprofile1",
					ReleaseName:       "testprofilereleasename",
					Namespace:         "testnamespace",
					KubernetesVersion: "1.12.3",
					RBName:            "testresourcebundle",
					RBVersion:         "v1",
				},
				{
					ProfileName:       "testprofile2",
					ReleaseName:       "testprofilereleasename2",
					Namespace:         "testnamespace2",
					KubernetesVersion: "1.12.3",
					RBName:            "testresourcebundle",
					RBVersion:         "v1",
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1", ProfileName: "testprofile1"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"testprofile1\"," +
								"\"release-name\":\"testprofilereleasename\"," +
								"\"namespace\":\"testnamespace\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"kubernetes-version\":\"1.12.3\"}"),
					},
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1", ProfileName: "testprofile2"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"testprofile2\"," +
								"\"release-name\":\"testprofilereleasename2\"," +
								"\"namespace\":\"testnamespace2\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"kubernetes-version\":\"1.12.3\"}"),
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
			got, err := impl.List(testCase.rbdef, testCase.version)
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
					return got[i].ProfileName < got[j].ProfileName
				})
				// Sort both as it is not expected that testCase.expected
				// is sorted
				sort.Slice(testCase.expected, func(i, j int) bool {
					return testCase.expected[i].ProfileName < testCase.expected[j].ProfileName
				})

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("List Resource Bundle returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteProfile(t *testing.T) {

	testCases := []struct {
		label                     string
		rbname, rbversion, prname string
		expectedError             string
		mockdb                    *db.MockDB
	}{
		{
			label:     "Delete Resource Bundle Profile",
			rbname:    "testresourcebundle",
			rbversion: "v1",
			prname:    "testprofile1",
			mockdb:    &db.MockDB{},
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
			err := impl.Delete(testCase.rbname, testCase.rbversion, testCase.prname)
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
		label                     string
		rbname, rbversion, prname string
		content                   []byte
		expectedError             string
		mockdb                    *db.MockDB
	}{
		{
			label:     "Upload Resource Bundle Profile",
			rbname:    "testresourcebundle",
			rbversion: "v1",
			prname:    "testprofile1",
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
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1", ProfileName: "testprofile1"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"testprofile1\"," +
								"\"release-name\":\"testprofilereleasename\"," +
								"\"namespace\":\"testnamespace\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
					},
				},
			},
		},
		{
			label:         "Upload with an Invalid Resource Bundle Profile",
			rbname:        "testresourcebundle",
			rbversion:     "v1",
			prname:        "testprofile1",
			expectedError: "Invalid Profile Name provided",
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
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1", ProfileName: "testprofile2"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"testprofile1\"," +
								"\"release-name\":\"testprofilereleasename\"," +
								"\"namespace\":\"testnamespace\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
					},
				},
			},
		},
		{
			label:         "Invalid File Format Error",
			rbname:        "testresourcebundle",
			rbversion:     "v1",
			prname:        "testprofile1",
			expectedError: "Error in file format",
			content: []byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1", ProfileName: "testprofile1"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"testprofile1\"," +
								"\"release-name\":\"testprofilereleasename\"," +
								"\"namespace\":\"testnamespace\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
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
			err := impl.Upload(testCase.rbname, testCase.rbversion, testCase.prname, testCase.content)
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

func TestDownloadProfile(t *testing.T) {
	testCases := []struct {
		label                     string
		rbname, rbversion, prname string
		expected                  []byte
		expectedError             string
		mockdb                    *db.MockDB
	}{
		{
			label:     "Download Resource Bundle Profile",
			rbname:    "testresourcebundle",
			rbversion: "v1",
			prname:    "testprofile1",
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
				Items: map[string]map[string][]byte{
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1", ProfileName: "testprofile1"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"testprofile1\"," +
								"\"release-name\":\"testprofilereleasename\"," +
								"\"namespace\":\"testnamespace\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
						"profilecontent": []byte("H4sICLBr9FsAA3Rlc3QudGFyAO3OQQrCMBCF4aw9RU5" +
							"QEtLE40igAUtSC+2IHt9IEVwIpYtShP/bvGFmFk/SLI08Re3IVCG077Rn" +
							"b75zYZ2yztVV8N7XP9vWSWmzZ6mP+yxx0lrF7pJzjkN/Sz//1u5/6ppKG" +
							"R/jVLrT0VUAAAAAAAAAAAAAAAAAABu8ALXoSvkAKAAA"),
					},
				},
			},
		},
		{
			label:         "Download with an Invalid Resource Bundle Profile",
			rbname:        "testresourcebundle",
			rbversion:     "v1",
			prname:        "testprofile1",
			expectedError: "Invalid Profile Name provided",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1", ProfileName: "testprofile2"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"testprofile1\"," +
								"\"release-name\":\"testprofilereleasename\"," +
								"\"namespace\":\"testnamespace\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
					},
				},
			},
		},
		{
			label:         "Download Error",
			expectedError: "DB Error",
			rbname:        "testresourcebundle",
			rbversion:     "v1",
			prname:        "testprofile1",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("DB Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewProfileClient()
			data, err := impl.Download(testCase.rbname, testCase.rbversion, testCase.prname)
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

func TestResolveProfile(t *testing.T) {
	testCases := []struct {
		label                     string
		rbname, rbversion, prname string
		expected                  map[string][]string
		expectedError             string
		mockdb                    *db.MockDB
	}{
		{
			label:     "Resolve Resource Bundle Profile",
			rbname:    "testresourcebundle",
			rbversion: "v1",
			prname:    "profile1",
			expected:  map[string][]string{},
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ProfileKey{RBName: "testresourcebundle", RBVersion: "v1",
						ProfileName: "profile1"}.String(): {
						"profilemetadata": []byte(
							"{\"profile-name\":\"profile1\"," +
								"\"release-name\":\"testprofilereleasename\"," +
								"\"namespace\":\"testnamespace\"," +
								"\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"kubernetesversion\":\"1.12.3\"}"),
						// base64 encoding of vagrant/tests/vnfs/testrb/helm/profile
						"profilecontent": []byte("H4sICLmjT1wAA3Byb2ZpbGUudGFyAO1Y32/bNhD2s/6Kg/KyYZZsy" +
							"78K78lLMsxY5gRxmqIYhoKWaJsYJWokZdfo+r/vSFmunCZNBtQJ1vF7sXX36e54vDN5T" +
							"knGFlTpcEtS3jgO2ohBr2c/EXc/29Gg1+h0e1F32Ol1B1Gj3Ymifr8B7SPFc4BCaSIBG" +
							"lII/SXeY/r/KIIg8NZUKiayEaw7nt7mdOQBrAkvqBqBL1ArWULflRJbJz4SYpEt2FJSJ" +
							"QoZ21cAAlgwTnOiVyPQWFQLwVuqmCdMthKac7FNaVZWmqWjkRWRuuSvScF1gFZVwYOEr" +
							"luapjknaOazd186Z98S7tver+3j0f5v1/q/18f+7w56bdf/zwFF5ZqV/WtbH6YioVdCa" +
							"hRkJEVBVSFBvUNRmyNpesgwors0lmkqM8KNzRG8iqLIWN45GUGv57l+fkFUP9PH9GF6f" +
							"IgH+kP9b76b/o+GUb9r5J1O1I0a0D9mUBX+5/1/55g+io9/sf+DnuF1sA4Gbv+fA1++p" +
							"n0dH4+c/92oPaztv+n/fn84dOf/c+AETkW+lWy50hC1O69gguc1R6HEw5xoHAuaKIq9E" +
							"+8ELvCikCmaQJElVIJeURjnJMaPnaYJt+UoAVHYhu8Mwd+p/O9/RAtbUUBKtnj+aygUR" +
							"RNM2ZkB6PuY5hpvCzhY4L2fkSymsGF6Zd3sjIRo4u3OhJhrgmyC/ByfFnUeEG0DLrHSO" +
							"h+1WpvNJiQ23FDIZYuXVNW6mJyeT2fnAYZsX3qdcaoUSPpXwSQudr4FkmNEMZljnJxsQ" +
							"EggOPmgTgsT8UYyzbJlE5RY6A2RFK0kTGnJ5oU+SFcVH666TsCEkQz88QwmMx9+Gs8ms" +
							"ybaeDO5+eXy9Q28GV9fj6c3k/MZXF7D6eX0bHIzuZzi088wnr6FXyfTsyZQTBa6oe9za" +
							"eLHIJlJJE1M1maUHgSwEGVAKqcxW7AY15UtC7KksDS3uQyXAzmVKVNmOxWGl6AVzlKmb" +
							"VGozxcVeh7J2W01S2LOVAsHyj9ZlozgbP+74qVUk4RoMtrfMD98wCzGvEiwXHD3U5GFi" +
							"4Jzo/QhhI8fd0yFu3c/fa/d8zmZU67KsRRDefCt/Qu7YdQSw1PzNTS3W1QGnyRVef+N5" +
							"YHDKZao/4MP/ju/siEpp0SVQYbX5UNlxxJwizCFyzuMWXkLNySzIyZs4wBrTpXE23I62" +
							"wlPRZHp0qJCC7EWslxpSnS8uqgt/YmLr2btnZXaDhnwA4NPzueT8lEt126AyExPY44rS" +
							"YA1bJPl15JgRaEdM9CKv/f1YDHdE5e1cYVFdiUwoduDJC+5mBMe5nstbndCF9Zfxakpa" +
							"1aNP2LK/Xffhuc3fTNfUYlfzH8a/h97qhmVaikNPi2+nItq8exGtLA+SdW9rgUvUvqbq" +
							"YkDi6mRXNk/V1pUxy0uYsI1S+meU+XsPo2kJLnMOKZGy4J6Xt3XgZuHTayEKv3XZLjy+" +
							"yJ66WPQwcHBwcHBwcHBwcHBwcHBwcHhm8Q/mTHqWgAoAAA="),
					},
					DefinitionKey{RBName: "testresourcebundle", RBVersion: "v1"}.String(): {
						"defmetadata": []byte(
							"{\"rb-name\":\"testresourcebundle\"," +
								"\"rb-version\":\"v1\"," +
								"\"chart-name\":\"vault-consul-dev\"," +
								"\"description\":\"testresourcebundle\"}"),
						// base64 encoding of vagrant/tests/vnfs/testrb/helm/vault-consul-dev
						"defcontent": []byte("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" +
							"cOHDgwIEDB18K/AcxEDJDAHgAAA=="),
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewProfileClient()
			data, err := impl.Resolve(testCase.rbname, testCase.rbversion, testCase.prname,
				[]string{})
			if err != nil {
				if testCase.expectedError == "" {
					t.Errorf("Resolve returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Errorf("Resolve returned an unexpected error %s", err)
				}
			}
			t.Log(data)
		})
	}
}
