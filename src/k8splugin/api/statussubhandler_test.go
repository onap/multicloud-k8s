/*
Copyright © 2022 Orange
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
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"

	pkgerrors "github.com/pkg/errors"
)

type mockInstanceStatusSubClient struct {
	app.InstanceStatusSubManager
	item  app.StatusSubscription
	items []app.StatusSubscription
	err   error
}

func (m *mockInstanceStatusSubClient) Create(ctx context.Context, instanceId string, subDetails app.SubscriptionRequest) (app.StatusSubscription, error) {
	if m.err != nil {
		return app.StatusSubscription{}, m.err
	}
	return m.item, nil
}

func (m *mockInstanceStatusSubClient) Get(ctx context.Context, instanceId, subId string) (app.StatusSubscription, error) {
	if m.err != nil {
		return app.StatusSubscription{}, m.err
	}
	return m.item, nil
}

func (m *mockInstanceStatusSubClient) Update(ctx context.Context, instanceId, subId string, subDetails app.SubscriptionRequest) (app.StatusSubscription, error) {
	if m.err != nil {
		return app.StatusSubscription{}, m.err
	}
	return m.item, nil
}

func (m *mockInstanceStatusSubClient) List(ctx context.Context, instanceId string) ([]app.StatusSubscription, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.items, nil
}

func (m *mockInstanceStatusSubClient) Delete(ctx context.Context, instanceId, subId string) error {
	return m.err
}

func (m *mockInstanceStatusSubClient) Cleanup(ctx context.Context, instanceId string) error {
	return m.err
}

func (m *mockInstanceStatusSubClient) RestoreWatchers(ctx context.Context) {}

func TestStatusSubCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expectedCode int
		subClient    *mockInstanceStatusSubClient
	}{
		{
			label:        "Empty body",
			input:        nil,
			expectedCode: http.StatusBadRequest,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Invalid JSON request format",
			input:        bytes.NewBuffer([]byte("invalid")),
			expectedCode: http.StatusUnprocessableEntity,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Missing name",
			input:        bytes.NewBuffer([]byte(`{"callback-url": "http://example.com"}`)),
			expectedCode: http.StatusBadRequest,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Negative min notify interval",
			input:        bytes.NewBuffer([]byte(`{"name": "sub1", "min-notify-interval": -1, "callback-url": "http://example.com"}`)),
			expectedCode: http.StatusBadRequest,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Missing callback url",
			input:        bytes.NewBuffer([]byte(`{"name": "sub1"}`)),
			expectedCode: http.StatusBadRequest,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Manager error",
			input:        bytes.NewBuffer([]byte(`{"name": "sub1", "callback-url": "http://example.com"}`)),
			expectedCode: http.StatusInternalServerError,
			subClient: &mockInstanceStatusSubClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful create",
			input:        bytes.NewBuffer([]byte(`{"name": "sub1", "callback-url": "http://example.com"}`)),
			expectedCode: http.StatusCreated,
			subClient: &mockInstanceStatusSubClient{
				item: app.StatusSubscription{Name: "sub1", CallbackUrl: "http://example.com"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/instance/HaKpys8e/status/subscription", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, testCase.subClient, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestStatusSubGetHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		subClient    *mockInstanceStatusSubClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			subClient: &mockInstanceStatusSubClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful get",
			expectedCode: http.StatusOK,
			subClient: &mockInstanceStatusSubClient{
				item: app.StatusSubscription{Name: "sub1"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/HaKpys8e/status/subscription/sub1", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, testCase.subClient, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestStatusSubUpdateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expectedCode int
		subClient    *mockInstanceStatusSubClient
	}{
		{
			label:        "Empty body",
			input:        nil,
			expectedCode: http.StatusBadRequest,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Invalid JSON request format",
			input:        bytes.NewBuffer([]byte("invalid")),
			expectedCode: http.StatusUnprocessableEntity,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Negative min notify interval",
			input:        bytes.NewBuffer([]byte(`{"min-notify-interval": -1, "callback-url": "http://example.com"}`)),
			expectedCode: http.StatusBadRequest,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Missing callback url",
			input:        bytes.NewBuffer([]byte(`{"name": "sub1"}`)),
			expectedCode: http.StatusBadRequest,
			subClient:    &mockInstanceStatusSubClient{},
		},
		{
			label:        "Manager error",
			input:        bytes.NewBuffer([]byte(`{"name": "sub1", "callback-url": "http://example.com"}`)),
			expectedCode: http.StatusInternalServerError,
			subClient: &mockInstanceStatusSubClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful update",
			input:        bytes.NewBuffer([]byte(`{"name": "sub1", "callback-url": "http://example.com"}`)),
			expectedCode: http.StatusOK,
			subClient: &mockInstanceStatusSubClient{
				item: app.StatusSubscription{Name: "sub1", CallbackUrl: "http://example.com"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v1/instance/HaKpys8e/status/subscription/sub1", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, testCase.subClient, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestStatusSubDeleteHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		subClient    *mockInstanceStatusSubClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			subClient: &mockInstanceStatusSubClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful delete",
			expectedCode: http.StatusAccepted,
			subClient:    &mockInstanceStatusSubClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v1/instance/HaKpys8e/status/subscription/sub1", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, testCase.subClient, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestStatusSubListHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		subClient    *mockInstanceStatusSubClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			subClient: &mockInstanceStatusSubClient{
				err: pkgerrors.New("Internal error"),
			},
		},
		{
			label:        "Successful list",
			expectedCode: http.StatusOK,
			subClient: &mockInstanceStatusSubClient{
				items: []app.StatusSubscription{{Name: "sub1"}},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/HaKpys8e/status/subscription", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, testCase.subClient, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}
