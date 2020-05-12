/*
 * Copyright 2020 Intel Corporation, Inc
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

package controller

import (
	"reflect"
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/types"

	pkgerrors "github.com/pkg/errors"
)

func TestCreateController(t *testing.T) {
	testCases := []struct {
		label         string
		inp           Controller
		expectedError string
		mockdb        *db.MockDB
		expected      Controller
	}{
		{
			label: "Create Controller",
			inp: Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expected: Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expectedError: "",
			mockdb:        &db.MockDB{},
		},
		{
			label:         "Failed Create Controller",
			expectedError: "Error Creating Controller",
			mockdb: &db.MockDB{
				Err: pkgerrors.New("Error Creating Controller"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			db.DBconn = testCase.mockdb
			impl := NewControllerClient()
			got, err := impl.CreateController(testCase.inp, false)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Create returned an unexpected error %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Create returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestGetController(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
		inp           string
		expected      Controller
	}{
		{
			label: "Get Controller",
			name:  "testController",
			expected: Controller{
				Metadata: types.Metadata{
					Name: "testController",
				},
				Spec: ControllerSpec{
					Host: "132.156.0.10",
					Port: 8080,
				},
			},
			expectedError: "",
			mockdb: &db.MockDB{
				Items: map[string]map[string][]byte{
					ControllerKey{ControllerName: "testController"}.String(): {
						"controllermetadata": []byte(
							"{\"metadata\":{" +
								"\"name\":\"testController\"" +
								"}," +
								"\"spec\":{" +
								"\"host\":\"132.156.0.10\"," +
								"\"port\": 8080 }}"),
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
			impl := NewControllerClient()
			got, err := impl.GetController(testCase.name)
			if err != nil {
				if testCase.expectedError == "" {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
				if strings.Contains(err.Error(), testCase.expectedError) == false {
					t.Fatalf("Get returned an unexpected error: %s", err)
				}
			} else {
				if reflect.DeepEqual(testCase.expected, got) == false {
					t.Errorf("Get returned unexpected body: got %v;"+
						" expected %v", got, testCase.expected)
				}
			}
		})
	}
}

func TestDeleteController(t *testing.T) {

	testCases := []struct {
		label         string
		name          string
		expectedError string
		mockdb        *db.MockDB
	}{
		{
			label:  "Delete Controller",
			name:   "testController",
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
			impl := NewControllerClient()
			err := impl.DeleteController(testCase.name)
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
