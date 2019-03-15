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
	"k8splugin/internal/rb"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

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
		rbDefClient  *mockRBProfile
	}{
		{
			label:        "Missing Body Failure",
			expectedCode: http.StatusBadRequest,
			rbDefClient:  &mockRBProfile{},
			reader:       nil,
		},
		{
			label:        "Create New Profile for Definition",
			expectedCode: http.StatusCreated,
			reader: bytes.NewBuffer([]byte(`{
				"rb-name":"testresource_bundle_definition",
				"rb-version":"v1",
				"profile-name":"profile1",
				"release-name":"testprofilereleasename",
				"namespace":"default",
				"kubernetes-version":"1.12.3"
				}`)),
			expected: rb.Profile{
				RBName:            "testresource_bundle_definition",
				RBVersion:         "v1",
				Name:              "profile1",
				ReleaseName:       "testprofilereleasename",
				Namespace:         "default",
				KubernetesVersion: "1.12.3",
			},
			rbDefClient: &mockRBProfile{
				//Items that will be returned by the mocked Client
				Items: []rb.Profile{
					{
						RBName:            "testresource_bundle_definition",
						RBVersion:         "v1",
						Name:              "profile1",
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
			vh := rbProfileHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("POST", "/v1/rb/profile/testresource_bundle_definition/v1/profile",
				testCase.reader)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.createHandler)
			hr.ServeHTTP(rr, req)

			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}

			//Check returned body only if statusCreated
			if rr.Code == http.StatusCreated {
				got := rb.Profile{}
				json.NewDecoder(rr.Body).Decode(&got)

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
		inpUUID      string
		expectedCode int
		rbDefClient  *mockRBProfile
	}{
		{
			label:        "Get Bundle Profile",
			expectedCode: http.StatusOK,
			expected: rb.Profile{
				RBName:            "testresource_bundle_definition",
				RBVersion:         "v1",
				Name:              "profile1",
				ReleaseName:       "testprofilereleasename",
				Namespace:         "default",
				KubernetesVersion: "1.12.3",
			},
			inpUUID: "123e4567-e89b-12d3-a456-426655441111",
			rbDefClient: &mockRBProfile{
				// Profile that will be returned by the mockclient
				Items: []rb.Profile{
					{
						RBName:            "testresource_bundle_definition",
						RBVersion:         "v1",
						Name:              "profile1",
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
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			rbDefClient: &mockRBProfile{
				// list of Profiles that will be returned by the mockclient
				Items: []rb.Profile{},
				Err:   pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := rbProfileHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("GET", "/v1/rb/profile/"+testCase.inpUUID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.getHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}

			//Check returned body only if statusOK
			if rr.Code == http.StatusOK {
				got := rb.Profile{}
				json.NewDecoder(rr.Body).Decode(&got)

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
		inpUUID      string
		expectedCode int
		rbDefClient  *mockRBProfile
	}{
		{
			label:        "Delete Bundle Profile",
			expectedCode: http.StatusNoContent,
			inpUUID:      "123e4567-e89b-12d3-a456-426655441111",
			rbDefClient:  &mockRBProfile{},
		},
		{
			label:        "Delete Non-Exiting Bundle Profile",
			expectedCode: http.StatusInternalServerError,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			rbDefClient: &mockRBProfile{
				Err: pkgerrors.New("Internal Error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := rbProfileHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("GET", "/v1/rb/profile/"+testCase.inpUUID, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.deleteHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}
		})
	}
}

func TestRBProfileUploadHandler(t *testing.T) {

	testCases := []struct {
		label        string
		inpUUID      string
		body         io.Reader
		expectedCode int
		rbDefClient  *mockRBProfile
	}{
		{
			label:        "Upload Bundle Profile Content",
			expectedCode: http.StatusOK,
			inpUUID:      "123e4567-e89b-12d3-a456-426655441111",
			body: bytes.NewBuffer([]byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			}),
			rbDefClient: &mockRBProfile{},
		},
		{
			label:        "Upload Invalid Bundle Profile Content",
			expectedCode: http.StatusInternalServerError,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			body: bytes.NewBuffer([]byte{
				0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0xff, 0xf2, 0x48, 0xcd,
			}),
			rbDefClient: &mockRBProfile{
				Err: pkgerrors.New("Internal Error"),
			},
		},
		{
			label:        "Upload Empty Body Content",
			expectedCode: http.StatusBadRequest,
			inpUUID:      "123e4567-e89b-12d3-a456-426655440000",
			rbDefClient:  &mockRBProfile{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			vh := rbProfileHandler{client: testCase.rbDefClient}
			req, err := http.NewRequest("POST",
				"/v1/rb/profile/"+testCase.inpUUID+"/content", testCase.body)

			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			hr := http.HandlerFunc(vh.uploadHandler)

			hr.ServeHTTP(rr, req)
			//Check returned code
			if rr.Code != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, rr.Code)
			}
		})
	}
}
