/*
Copyright © 2021 Orange
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
	"context"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
)

type mockQueryClient struct {
	app.QueryManager
	status app.QueryStatus
	err    error
}

func (m *mockQueryClient) Query(ctx context.Context, namespace, cloudRegion, apiVersion, kind, name, labels string) (app.QueryStatus, error) {
	if m.err != nil {
		return app.QueryStatus{}, m.err
	}
	return m.status, nil
}

func TestQueryHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        map[string]string
		expectedCode int
		queryClient  *mockQueryClient
	}{
		{
			label: "Missing CloudRegion mandatory parameter",
			input: map[string]string{
				"ApiVersion": "v1",
				"Kind":       "Pod",
			},
			expectedCode: http.StatusBadRequest,
			queryClient:  &mockQueryClient{},
		},
		{
			label: "Missing ApiVersion mandatory parameter",
			input: map[string]string{
				"CloudRegion": "kud",
				"Kind":        "Pod",
			},
			expectedCode: http.StatusBadRequest,
			queryClient:  &mockQueryClient{},
		},
		{
			label: "Missing Kind mandatory parameter",
			input: map[string]string{
				"CloudRegion": "kud",
				"ApiVersion":  "v1",
			},
			expectedCode: http.StatusBadRequest,
			queryClient:  &mockQueryClient{},
		},
		{
			label: "Successful query",
			input: map[string]string{
				"CloudRegion": "kud",
				"ApiVersion":  "v1",
				"Kind":        "Pod",
				"Name":        "test",
				"Namespace":   "default",
			},
			expectedCode: http.StatusOK,
			queryClient: &mockQueryClient{
				status: app.QueryStatus{
					ResourceCount: 1,
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			params := neturl.Values{}
			for k, v := range testCase.input {
				params.Add(k, v)
			}
			url := "/v1/query?" + params.Encode()
			request := httptest.NewRequest("GET", url, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, testCase.queryClient, nil, nil, nil, nil, nil))

			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v",
					resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}
