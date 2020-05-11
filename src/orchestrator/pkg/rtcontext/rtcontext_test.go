/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *	   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rtcontext

import (
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
	pkgerrors "github.com/pkg/errors"
	"strings"
	"testing"
)

// MockContextDb for mocking contextdb
type MockContextDb struct {
	Items map[string]interface{}
	Err   error
}

// Put function
func (c *MockContextDb) Put(key string, val interface{}) error {
	if c.Items == nil {
		c.Items = make(map[string]interface{})
	}
	c.Items[key] = val
	return c.Err
}

// Get function
func (c *MockContextDb) Get(key string, val interface{}) error {
	var s *string
	s = val.(*string)
	for kvKey, kvValue := range c.Items {
		if kvKey == key {
			*s = kvValue.(string)
			return c.Err
		}
	}
	return c.Err
}

// Delete function
func (c *MockContextDb) Delete(key string) error {
	delete(c.Items, key)
	return c.Err
}

// Delete all function
func (c *MockContextDb) DeleteAll(key string) error {
	for kvKey := range c.Items {
		delete(c.Items, kvKey)
	}
	return c.Err
}

// GetAllKeys function
func (c *MockContextDb) GetAllKeys(path string) ([]string, error) {
	var keys []string

	for k := range c.Items {
		keys = append(keys, string(k))
	}
	return keys, c.Err
}

func (c *MockContextDb) HealthCheck() error {
	return nil
}

func TestRtcInit(t *testing.T) {
	var rtc = RunTimeContext{}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		expectedError string
	}{
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
		},
		{
			label:         "Init returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Client not intialized")},
			expectedError: "Error, context already initialized",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			_, err := rtc.RtcInit()
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}

		})
	}
}

func TestRtcLoad(t *testing.T) {
	var rtc = RunTimeContext{"", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		id            string
		expectedError string
	}{
		{
			label:         "Success case",
			id:            "5345674458787728",
			mockContextDb: &MockContextDb{},
		},
		{
			label:         "reinit returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Client not intialized")},
			id:            "8885674458787728",
			expectedError: "Error finding the context id:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			_, err := rtc.RtcLoad("5345674458787728")
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}

		})
	}
}

func TestRtcCreate(t *testing.T) {
	var rtc = RunTimeContext{"/context/5345674458787728/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		expectedError string
	}{
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
		},
		{
			label:         "Create returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Client not intialized")},
			expectedError: "Error creating run time context:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			_, err := rtc.RtcCreate()
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}

		})
	}
}

func TestRtcGet(t *testing.T) {
	var rtc = RunTimeContext{"/context/5345674458787728/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		expectedError string
	}{
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
		},
		{
			label:         "Get returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Client not intialized")},
			expectedError: "Error getting run time context metadata:",
		},
		{
			label:         "Context handle does not match",
			mockContextDb: &MockContextDb{Err: nil},
			expectedError: "Error matching run time context metadata",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			switch testCase.label {
			case "Success case":
				contextdb.Db = testCase.mockContextDb
				chandle, err := rtc.RtcCreate()
				if err != nil {
					t.Fatalf("Create returned an error (%s)", err)
				}
				ghandle, err := rtc.RtcGet()
				if err != nil {
					t.Fatalf("Get returned an error (%s)", err)
				}
				if chandle != ghandle {
					t.Fatalf("Create and Get does not match")
				}
			case "Get returns error case":
				contextdb.Db = testCase.mockContextDb
				_, err := rtc.RtcGet()
				if err != nil {
					if !strings.Contains(string(err.Error()), testCase.expectedError) {
						t.Fatalf("Method returned an error (%s)", err)
					}
				}
			case "Context handle does not match":
				contextdb.Db = testCase.mockContextDb
				contextdb.Db.Put("/context/5345674458787728/", "6345674458787728")
				_, err := rtc.RtcGet()
				if err != nil {
					if !strings.Contains(string(err.Error()), testCase.expectedError) {
						t.Fatalf("Method returned an error (%s)", err)
					}
				}
			}
		})
	}
}

func TestRtcAddLevel(t *testing.T) {
	var rtc = RunTimeContext{"/context/3528435435454354/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		handle        interface{}
		level         string
		value         string
		expectedError string
	}{
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/",
			level:         "app",
			value:         "testapp1",
		},
		{
			label:         "Not a valid rtc handle",
			mockContextDb: &MockContextDb{},
			handle:        "/context/9528435435454354/",
			level:         "app",
			value:         "testapp1",
			expectedError: "Not a valid run time context handle",
		},
		{
			label:         "Not a valid rtc level",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/",
			level:         "",
			value:         "testapp1",
			expectedError: "Not a valid run time context level",
		},
		{
			label:         "Not a valid rtc value",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/",
			level:         "app",
			value:         "",
			expectedError: "Not a valid run time context level value",
		},
		{
			label:         "Put returns error",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Client not intialized")},
			handle:        "/context/3528435435454354/",
			level:         "app",
			value:         "testapp1",
			expectedError: "Error adding run time context level:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			_, err := rtc.RtcAddLevel(testCase.handle, testCase.level, testCase.value)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestRtcAddResource(t *testing.T) {
	var rtc = RunTimeContext{"/context/3528435435454354/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		handle        interface{}
		resname       string
		value         interface{}
		expectedError string
	}{
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			resname:       "R1",
			value:         "res1",
		},
		{
			label:         "Not a valid rtc handle",
			mockContextDb: &MockContextDb{},
			handle:        "/context/9528435435454354/app/apptest1/cluster/cluster1/",
			resname:       "R1",
			value:         "res1",
			expectedError: "Not a valid run time context handle",
		},
		{
			label:         "Not a valid rtc resource name",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			resname:       "",
			value:         "res1",
			expectedError: "Not a valid run time context resource name",
		},
		{
			label:         "Not a valid rtc value",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			resname:       "R1",
			value:         nil,
			expectedError: "Not a valid run time context resource value",
		},
		{
			label:         "Put returns error",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Client not intialized")},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			resname:       "R1",
			value:         "res1",
			expectedError: "Error adding run time context resource:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			_, err := rtc.RtcAddResource(testCase.handle, testCase.resname, testCase.value)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestRtcAddInstruction(t *testing.T) {
	var rtc = RunTimeContext{"/context/3528435435454354/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		handle        interface{}
		level         string
		insttype      string
		value         interface{}
		expectedError string
	}{
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			level:         "resource",
			insttype:      "order",
			value:         "{resorder: [R3, R1, R2]}",
		},
		{
			label:         "Not a valid rtc handle",
			mockContextDb: &MockContextDb{},
			handle:        "/context/9528435435454354/app/apptest1/cluster/cluster1/",
			level:         "resource",
			insttype:      "order",
			value:         "{resorder: [R3, R1, R2]}",
			expectedError: "Not a valid run time context handle",
		},
		{
			label:         "Not a valid rtc level name",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			level:         "",
			insttype:      "order",
			value:         "{resorder: [R3, R1, R2]}",
			expectedError: "Not a valid run time context level",
		},
		{
			label:         "Not a valid rtc instruction type",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			level:         "resource",
			insttype:      "",
			value:         "{resorder: [R3, R1, R2]}",
			expectedError: "Not a valid run time context instruction type",
		},
		{
			label:         "Not a valid rtc value",
			mockContextDb: &MockContextDb{},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			level:         "resource",
			insttype:      "order",
			value:         nil,
			expectedError: "Not a valid run time context instruction value",
		},
		{
			label:         "Put returns error",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Client not intialized")},
			handle:        "/context/3528435435454354/app/apptest1/cluster/cluster1/",
			level:         "resource",
			insttype:      "order",
			value:         "{resorder: [R3, R1, R2]}",
			expectedError: "Error adding run time context instruction:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			_, err := rtc.RtcAddInstruction(testCase.handle, testCase.level, testCase.insttype, testCase.value)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestRtcGetHandles(t *testing.T) {
	var rtc = RunTimeContext{"/context/5345674458787728/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		key           interface{}
		expectedError string
	}{
		{
			label:         "Not valid input handle case",
			mockContextDb: &MockContextDb{},
			key:           "/context/3528435435454354/",
			expectedError: "Not a valid run time context handle",
		},
		{
			label:         "Contextdb call returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Key does not exist")},
			key:           "/context/5345674458787728/",
			expectedError: "Error getting run time context handles:",
		},
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
			key:           "/context/5345674458787728/",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			if testCase.label == "Success case" {
				contextdb.Db.Put("/context/5345674458787728/", 5345674458787728)
			}
			_, err := rtc.RtcGetHandles(testCase.key)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestRtcGetValue(t *testing.T) {
	var rtc = RunTimeContext{"/context/5345674458787728/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		key           interface{}
		expectedError string
	}{
		{
			label:         "Not valid input handle case",
			mockContextDb: &MockContextDb{},
			key:           "/context/3528435435454354/",
			expectedError: "Not a valid run time context handle",
		},
		{
			label:         "Contextdb call returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Key does not exist")},
			key:           "/context/5345674458787728/",
			expectedError: "Error getting run time context value:",
		},
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
			key:           "/context/5345674458787728/",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			if testCase.label == "Success case" {
				contextdb.Db.Put("/context/5345674458787728/", "5345674458787728")
			}
			var val string
			err := rtc.RtcGetValue(testCase.key, &val)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestRtcUpdateValue(t *testing.T) {
	var rtc = RunTimeContext{"/context/5345674458787728/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		key           interface{}
		value         interface{}
		expectedError string
	}{
		{
			label:         "Not valid input handle case",
			mockContextDb: &MockContextDb{},
			key:           "/context/3528435435454354/",
			value:         "{apporder: [app1, app2, app3]}",
			expectedError: "Not a valid run time context handle",
		},
		{
			label:         "Contextdb call returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Key does not exist")},
			key:           "/context/5345674458787728/",
			value:         "{apporder: [app1, app2, app3]}",
			expectedError: "Error updating run time context value:",
		},
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
			key:           "/context/5345674458787728/",
			value:         "{apporder: [app2, app3, app1]}",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			if testCase.label == "Success case" {
				contextdb.Db.Put("/context/5345674458787728/", "5345674458787728")
			}
			err := rtc.RtcUpdateValue(testCase.key, testCase.value)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestRtcDeletePair(t *testing.T) {
	var rtc = RunTimeContext{"/context/5345674458787728/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		key           interface{}
		expectedError string
	}{
		{
			label:         "Not valid input handle case",
			mockContextDb: &MockContextDb{},
			key:           "/context/3528435435454354/",
			expectedError: "Not a valid run time context handle",
		},
		{
			label:         "Contextdb call returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Key does not exist")},
			key:           "/context/5345674458787728/",
			expectedError: "Error deleting run time context pair:",
		},
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
			key:           "/context/5345674458787728/",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			err := rtc.RtcDeletePair(testCase.key)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}

func TestRtcDeletePrefix(t *testing.T) {
	var rtc = RunTimeContext{"/context/5345674458787728/", ""}
	testCases := []struct {
		label         string
		mockContextDb *MockContextDb
		key           interface{}
		expectedError string
	}{
		{
			label:         "Not valid input handle case",
			mockContextDb: &MockContextDb{},
			key:           "/context/3528435435454354/",
			expectedError: "Not a valid run time context handle",
		},
		{
			label:         "Contextdb call returns error case",
			mockContextDb: &MockContextDb{Err: pkgerrors.Errorf("Key does not exist")},
			key:           "/context/5345674458787728/",
			expectedError: "Error deleting run time context with prefix:",
		},
		{
			label:         "Success case",
			mockContextDb: &MockContextDb{},
			key:           "/context/5345674458787728/",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			contextdb.Db = testCase.mockContextDb
			err := rtc.RtcDeletePrefix(testCase.key)
			if err != nil {
				if !strings.Contains(string(err.Error()), testCase.expectedError) {
					t.Fatalf("Method returned an error (%s)", err)
				}
			}
		})
	}
}
