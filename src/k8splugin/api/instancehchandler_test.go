/*
Copyright © 2021 Samsung Electronics
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/healthcheck"

	pkgerrors "github.com/pkg/errors"
)

type mockInstanceHCClient struct {
	healthcheck.InstanceHCManager
	miniStatus healthcheck.InstanceMiniHCStatus
	status     healthcheck.InstanceHCStatus
	overview   healthcheck.InstanceHCOverview
	err        error
}

func (m *mockInstanceHCClient) Create(instanceId string) (healthcheck.InstanceMiniHCStatus, error) {
	if m.err != nil {
		return healthcheck.InstanceMiniHCStatus{}, m.err
	}
	return m.miniStatus, nil
}

func (m *mockInstanceHCClient) Get(instanceId, healthcheckId string) (healthcheck.InstanceHCStatus, error) {
	if m.err != nil {
		return healthcheck.InstanceHCStatus{}, m.err
	}
	return m.status, nil
}

func (m *mockInstanceHCClient) List(instanceId string) (healthcheck.InstanceHCOverview, error) {
	if m.err != nil {
		return healthcheck.InstanceHCOverview{}, m.err
	}
	return m.overview, nil
}

func (m *mockInstanceHCClient) Delete(instanceId, healthcheckId string) error {
	return m.err
}

func TestInstanceHCCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		instID       string
		expectedCode int
		hcClient     *mockInstanceHCClient
	}{
		{
			label:        "Fail to create healthcheck",
			instID:       "HaKpys8e",
			expectedCode: http.StatusInternalServerError,
			hcClient: &mockInstanceHCClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful create healthcheck",
			instID:       "HaKpys8e",
			expectedCode: http.StatusCreated,
			hcClient: &mockInstanceHCClient{
				miniStatus: healthcheck.InstanceMiniHCStatus{
					HealthcheckId: "hc1",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/instance/"+testCase.instID+"/healthcheck", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, testCase.hcClient))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestInstanceHCGetHandler(t *testing.T) {
	testCases := []struct {
		label        string
		instID       string
		hcID         string
		expectedCode int
		hcClient     *mockInstanceHCClient
	}{
		{
			label:        "Fail to get healthcheck",
			instID:       "HaKpys8e",
			hcID:         "hc1",
			expectedCode: http.StatusInternalServerError,
			hcClient: &mockInstanceHCClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful get healthcheck",
			instID:       "HaKpys8e",
			hcID:         "hc1",
			expectedCode: http.StatusOK,
			hcClient: &mockInstanceHCClient{
				status: healthcheck.InstanceHCStatus{},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/"+testCase.instID+"/healthcheck/"+testCase.hcID, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, testCase.hcClient))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestInstanceHCListHandler(t *testing.T) {
	testCases := []struct {
		label        string
		instID       string
		expectedCode int
		hcClient     *mockInstanceHCClient
	}{
		{
			label:        "Fail to list healthchecks",
			instID:       "HaKpys8e",
			expectedCode: http.StatusInternalServerError,
			hcClient: &mockInstanceHCClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful list healthchecks",
			instID:       "HaKpys8e",
			expectedCode: http.StatusOK,
			hcClient: &mockInstanceHCClient{
				overview: healthcheck.InstanceHCOverview{
					InstanceId: "HaKpys8e",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/"+testCase.instID+"/healthcheck", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, testCase.hcClient))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestInstanceHCDeleteHandler(t *testing.T) {
	testCases := []struct {
		label        string
		instID       string
		hcID         string
		expectedCode int
		hcClient     *mockInstanceHCClient
	}{
		{
			label:        "Fail to delete healthcheck",
			instID:       "HaKpys8e",
			hcID:         "hc1",
			expectedCode: http.StatusInternalServerError,
			hcClient: &mockInstanceHCClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful delete healthcheck",
			instID:       "HaKpys8e",
			hcID:         "hc1",
			expectedCode: http.StatusAccepted,
			hcClient:     &mockInstanceHCClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v1/instance/"+testCase.instID+"/healthcheck/"+testCase.hcID, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, testCase.hcClient))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}
