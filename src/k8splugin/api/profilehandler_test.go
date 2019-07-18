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

package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	pkgerrors "github.com/pkg/errors"
)

//Creating an embedded interface via anonymous variable
//This allows us to make mockDB satisfy the DatabaseConnection
//interface even if we are not implementing all the methods in it
type mockRBProfile struct {
	rb.ProfileManager
	// Items and err will be used to customize each test
	// via a localized instantiation of mockRBProfile
	Items []rb.Profile
	Err   error
}

func (m *mockRBProfile) Create(inp rb.Profile) (rb.Profile, error) {
	if m.Err != nil {
		return rb.Profile{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockRBProfile) Get(rbname, rbversion, prname string) (rb.Profile, error) {
	if m.Err != nil {
		return rb.Profile{}, m.Err
	}

	return m.Items[0], nil
}

func (m *mockRBProfile) List(rbname, rbversion string) ([]rb.Profile, error) {
	if m.Err != nil {
		return []rb.Profile{}, m.Err
	}

	return m.Items, nil
}

func (m *mockRBProfile) Delete(rbname, rbversion, prname string) error {
	return m.Err
}

func (m *mockRBProfile) Upload(rbname, rbversion, prname string, inp []byte) error {
	return m.Err
}

func TestRBProfileCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		reader       io.Reader
		expected     rb.Profile
		expectedCode int
		rbProClient  *mockRBProfile
	}{
		{
			label:        "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			rbProClient:  &mockRBProfile{},
			reader:       nil,
		},
		{
			label:        "Create New Profile for Definition",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"rb-name":"test-rbdef",
				"rb-version":"v1",
				"profile-name":"profile1",
				"release-name":"testprofilereleasename",
				"namespace":"default",
				"kubernetes-version":"1.12.3"
				}`)),
			expected: rb.Profile{
				RBName:            "test-rbdef",
				RBVersion:         "v1",
				ProfileName:       "profile1",
				ReleaseName:       "testprofilereleasename",
				Namespace:         "default",
				KubernetesVersion: "1.12.3",
			},
			rbProClient: &mockRBProfile{
				//Items that will be returned by the mocked Client
				Items: []rb.Profile{
					{
						RBName:            "test-rbdef",
						RBVersion:         "v1",
						ProfileName:       "profile1",
						ReleaseName:       "testprofilereleasename",
						Namespace:         "default",
						KubernetesVersion: "1.12.3",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/rb/definition/test-rbdef/v1/profile",
				testCase.reader)
			resp := executeRequest(request, NewRouter(nil, testCase.rbProClient, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusCreated
			if resp.StatusCode == http.StatusCreated {
				got := rb.Profile{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("createHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestRBProfileGetHandler(t *testing.T) {

	testCases := []struct {
		label        string
		expected     rb.Profile
		prname       string
		expectedCode int
		rbProClient  *mockRBProfile
	}{
		{
			label:        "Get Bundle Profile",
			expectedCode: http.StatusOK,
			expected: rb.Profile{
				RBName:            "test-rbdef",
				RBVersion:         "v1",
				ProfileName:       "profile1",
				ReleaseName:       "testprofilereleasename",
				Namespace:         "default",
				KubernetesVersion: "1.12.3",
			},
			prname: "profile1",
			rbProClient: &mockRBProfile{
				// Profile that will be returned by the mockclient
				Items: []rb.Profile{
					{
						RBName:            "test-rbdef",
						RBVersion:         "v1",
						ProfileName:       "profile1",
						ReleaseName:       "testprofilereleasename",
						Namespace:         "default",
						KubernetesVersion: "1.12.3",
					},
				},
			},
		},
		{
			label:        "Get Non-Exiting Bundle Profile",
			expectedCode: http.StatusInternalServerError,
			prname:       "non-existing-profile",
			rbProClient: &mockRBProfile{
				// list of Profiles that will be returned by the mockclient
				Items: []rb.Profile{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/rb/definition/test-rbdef/v1/profile/"+testCase.prname, nil)
			resp := executeRequest(request, NewRouter(nil, testCase.rbProClient, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := rb.Profile{}
				json.NewDecoder(resp.Body).Decode(&got)

				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestRBProfileListHandler(t *testing.T) {

	testCases := []struct {
		def          string
		version      string
		label        string
		expected     []rb.Profile
		expectedCode int
		rbProClient  *mockRBProfile
	}{
		{
			def:          "test-rbdef",
			version:      "v1",
			label:        "List Profiles",
			expectedCode: http.StatusOK,
			expected: []rb.Profile{
				{
					RBName:            "test-rbdef",
					RBVersion:         "v1",
					ProfileName:       "profile1",
					ReleaseName:       "testprofilereleasename",
					Namespace:         "ns1",
					KubernetesVersion: "1.12.3",
				},
				{
					RBName:            "test-rbdef",
					RBVersion:         "v1",
					ProfileName:       "profile2",
					ReleaseName:       "testprofilereleasename",
					Namespace:         "ns2",
					KubernetesVersion: "1.12.3",
				},
			},
			rbProClient: &mockRBProfile{
				// list of Profiles that will be returned by the mockclient
				Items: []rb.Profile{
					{
						RBName:            "test-rbdef",
						RBVersion:         "v1",
						ProfileName:       "profile1",
						ReleaseName:       "testprofilereleasename",
						Namespace:         "ns1",
						KubernetesVersion: "1.12.3",
					},
					{
						RBName:            "test-rbdef",
						RBVersion:         "v1",
						ProfileName:       "profile2",
						ReleaseName:       "testprofilereleasename",
						Namespace:         "ns2",
						KubernetesVersion: "1.12.3",
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/rb/definition/"+testCase.def+"/"+testCase.version+"/profile", nil)
			resp := executeRequest(request, NewRouter(nil, testCase.rbProClient, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			//Check returned body only if statusOK
			if resp.StatusCode == http.StatusOK {
				got := []rb.Profile{}
				json.NewDecoder(resp.Body).Decode(&got)

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
					t.Errorf("listHandler returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestRBProfileDeleteHandler(t *testing.T) {

	testCases := []struct {
		label        string
		prname       string
		expectedCode int
		rbProClient  *mockRBProfile
	}{
		{
			label:        "Delete Bundle Profile",
			expectedCode: http.StatusNoContent,
			prname:       "profile1",
			rbProClient:  &mockRBProfile{},
		},
		{
			label:        "Delete Non-Exiting Bundle Profile",
			expectedCode: http.StatusInternalServerError,
			prname:       "non-existing",
			rbProClient: &mockRBProfile{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v1/rb/definition/test-rbdef/v1/profile/"+testCase.prname, nil)
			resp := executeRequest(request, NewRouter(nil, testCase.rbProClient, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestRBProfileUploadHandler(t *testing.T) {

	testCases := []struct {
		label        string
		prname       string
		body         io.Reader
		expectedCode int
		rbProClient  *mockRBProfile
	}{
		{
			label:        "Upload Bundle Profile Content",
			expectedCode: http.StatusOK,
			prname:       "profile1",
			body: bytes.NewBuffer([]byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			}),
			rbProClient: &mockRBProfile{},
		},
		{
			label:        "Upload Invalid Bundle Profile Content",
			expectedCode: http.StatusInternalServerError,
			prname:       "profile1",
			body: bytes.NewBuffer([]byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			}),
			rbProClient: &mockRBProfile{
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Upload Empty Body Content",
			expectedCode: http.StatusBadRequest,
			prname:       "profile1",
			rbProClient:  &mockRBProfile{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST",
				"/v1/rb/definition/test-rbdef/v1/profile/"+testCase.prname+"/content", testCase.body)
			resp := executeRequest(request, NewRouter(nil, testCase.rbProClient, nil, nil, nil, nil))

			//Check returned code
			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
