/*
 * Copyright © 2021 Orange
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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"

	pkgerrors "github.com/pkg/errors"
)

type mockConfigClient struct {
	app.ConfigManager
	result  app.ConfigResult
	config  app.Config
	configs []app.Config
	tag     app.ConfigTag
	tags    []app.ConfigTag
	err     error
}

func (m *mockConfigClient) Create(instanceID string, p app.Config) (app.ConfigResult, error) {
	if m.err != nil {
		return app.ConfigResult{}, m.err
	}
	return m.result, nil
}

func (m *mockConfigClient) Get(instanceID, configName string) (app.Config, error) {
	if m.err != nil {
		return app.Config{}, m.err
	}
	return m.config, nil
}

func (m *mockConfigClient) GetVersion(instanceID, configName, version string) (app.Config, error) {
	if m.err != nil {
		return app.Config{}, m.err
	}
	return m.config, nil
}

func (m *mockConfigClient) GetTag(instanceID, configName, tagName string) (app.Config, error) {
	if m.err != nil {
		return app.Config{}, m.err
	}
	return m.config, nil
}

func (m *mockConfigClient) List(instanceID string) ([]app.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.configs, nil
}

func (m *mockConfigClient) VersionList(instanceID, configName string) ([]app.Config, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.configs, nil
}

func (m *mockConfigClient) Help() map[string]string {
	return map[string]string{}
}

func (m *mockConfigClient) Update(instanceID, configName string, p app.Config) (app.ConfigResult, error) {
	if m.err != nil {
		return app.ConfigResult{}, m.err
	}
	return m.result, nil
}

func (m *mockConfigClient) Delete(instanceID, configName string) (app.ConfigResult, error) {
	if m.err != nil {
		return app.ConfigResult{}, m.err
	}
	return m.result, nil
}

func (m *mockConfigClient) DeleteAll(instanceID, configName string, deleteConfigOnly bool) error {
	return m.err
}

func (m *mockConfigClient) Rollback(instanceID string, configName string, p app.ConfigRollback, acceptRevert bool) (app.ConfigResult, error) {
	if m.err != nil {
		return app.ConfigResult{}, m.err
	}
	return m.result, nil
}

func (m *mockConfigClient) Cleanup(instanceID string) error {
	return m.err
}

func (m *mockConfigClient) Tagit(instanceID string, configName string, p app.ConfigTagit) (app.ConfigTag, error) {
	if m.err != nil {
		return app.ConfigTag{}, m.err
	}
	return m.tag, nil
}

func (m *mockConfigClient) TagList(instanceID, configName string) ([]app.ConfigTag, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tags, nil
}

func TestConfigCreateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Invalid JSON request format",
			input:        bytes.NewBuffer([]byte("invalid")),
			expectedCode: http.StatusUnprocessableEntity,
			configClient: &mockConfigClient{},
		},
		{
			label:        "Missing config name",
			input:        bytes.NewBuffer([]byte(`{"template-name": "t1"}`)),
			expectedCode: http.StatusBadRequest,
			configClient: &mockConfigClient{},
		},
		{
			label:        "Manager error",
			input:        bytes.NewBuffer([]byte(`{"config-name": "c1", "template-name": "t1"}`)),
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful create",
			input:        bytes.NewBuffer([]byte(`{"config-name": "c1", "template-name": "t1"}`)),
			expectedCode: http.StatusCreated,
			configClient: &mockConfigClient{result: app.ConfigResult{ConfigName: "c1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/instance/HaKpys8e/config", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigGetHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful get",
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{config: app.Config{ConfigName: "c1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/HaKpys8e/config/c1", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigListHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful list",
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{configs: []app.Config{{ConfigName: "c1"}}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/HaKpys8e/config", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigUpdateHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Invalid JSON request format",
			input:        bytes.NewBuffer([]byte("invalid")),
			expectedCode: http.StatusUnprocessableEntity,
			configClient: &mockConfigClient{},
		},
		{
			label:        "Manager error",
			input:        bytes.NewBuffer([]byte(`{"config-name": "c1"}`)),
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful update",
			input:        bytes.NewBuffer([]byte(`{"config-name": "c1"}`)),
			expectedCode: http.StatusCreated,
			configClient: &mockConfigClient{result: app.ConfigResult{ConfigName: "c1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", "/v1/instance/HaKpys8e/config/c1", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigDeleteAllHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful delete all",
			expectedCode: http.StatusAccepted,
			configClient: &mockConfigClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v1/instance/HaKpys8e/config/c1", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigDeleteHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful delete",
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{result: app.ConfigResult{ConfigName: "c1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/instance/HaKpys8e/config/c1/delete", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigRollbackHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Invalid JSON request format",
			input:        bytes.NewBuffer([]byte("invalid")),
			expectedCode: http.StatusUnprocessableEntity,
			configClient: &mockConfigClient{},
		},
		{
			label:        "Manager error",
			input:        bytes.NewBuffer([]byte(`{"config-version": "1"}`)),
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful rollback",
			input:        bytes.NewBuffer([]byte(`{"config-version": "1"}`)),
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{result: app.ConfigResult{ConfigName: "c1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/instance/HaKpys8e/config/c1/rollback", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigTagitHandler(t *testing.T) {
	testCases := []struct {
		label        string
		input        io.Reader
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Invalid JSON request format",
			input:        bytes.NewBuffer([]byte("invalid")),
			expectedCode: http.StatusUnprocessableEntity,
			configClient: &mockConfigClient{},
		},
		{
			label:        "Manager error",
			input:        bytes.NewBuffer([]byte(`{"tag-name": "tag1"}`)),
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful tagit",
			input:        bytes.NewBuffer([]byte(`{"tag-name": "tag1"}`)),
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{tag: app.ConfigTag{ConfigTag: "tag1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", "/v1/instance/HaKpys8e/config/c1/tagit", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigTagListHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful tag list",
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{tags: []app.ConfigTag{{ConfigTag: "tag1"}}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/HaKpys8e/config/c1/tag", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigGetTagHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful get tag",
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{config: app.Config{ConfigName: "c1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/HaKpys8e/config/c1/tag/tag1", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigVersionListHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful version list",
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{configs: []app.Config{{ConfigName: "c1"}}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/HaKpys8e/config/c1/version", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigGetVersionHandler(t *testing.T) {
	testCases := []struct {
		label        string
		expectedCode int
		configClient *mockConfigClient
	}{
		{
			label:        "Manager error",
			expectedCode: http.StatusInternalServerError,
			configClient: &mockConfigClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:        "Successful get version",
			expectedCode: http.StatusOK,
			configClient: &mockConfigClient{config: app.Config{ConfigName: "c1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/instance/HaKpys8e/config/c1/version/1", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, testCase.configClient, nil, nil, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}
