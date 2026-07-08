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

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/rb"

	pkgerrors "github.com/pkg/errors"
)

type mockConfigTemplateClient struct {
	rb.ConfigTemplateManager
	template rb.ConfigTemplate
	list     []rb.ConfigTemplateList
	err      error
}

func (m *mockConfigTemplateClient) CreateOrUpdate(rbName, rbVersion string, p rb.ConfigTemplate, update bool) error {
	return m.err
}

func (m *mockConfigTemplateClient) Get(rbName, rbVersion, templateName string) (rb.ConfigTemplate, error) {
	if m.err != nil {
		return rb.ConfigTemplate{}, m.err
	}
	return m.template, nil
}

func (m *mockConfigTemplateClient) List(rbName, rbVersion string) ([]rb.ConfigTemplateList, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.list, nil
}

func (m *mockConfigTemplateClient) Delete(rbName, rbVersion, templateName string) error {
	return m.err
}

func (m *mockConfigTemplateClient) Upload(rbName, rbVersion, templateName string, inp []byte) error {
	return m.err
}

const cfgTmplBasePath = "/v1/rb/definition/test-rbdef/v1/config-template"

func TestConfigTemplateCreateHandler(t *testing.T) {
	testCases := []struct {
		label          string
		input          io.Reader
		expectedCode   int
		templateClient *mockConfigTemplateClient
	}{
		{
			label:          "Empty body",
			input:          nil,
			expectedCode:   http.StatusBadRequest,
			templateClient: &mockConfigTemplateClient{},
		},
		{
			label:          "Invalid JSON request format",
			input:          bytes.NewBuffer([]byte("invalid")),
			expectedCode:   http.StatusUnprocessableEntity,
			templateClient: &mockConfigTemplateClient{},
		},
		{
			label:          "Missing template name",
			input:          bytes.NewBuffer([]byte(`{"description": "d"}`)),
			expectedCode:   http.StatusBadRequest,
			templateClient: &mockConfigTemplateClient{},
		},
		{
			label:          "Manager error",
			input:          bytes.NewBuffer([]byte(`{"template-name": "t1"}`)),
			expectedCode:   http.StatusInternalServerError,
			templateClient: &mockConfigTemplateClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:          "Successful create",
			input:          bytes.NewBuffer([]byte(`{"template-name": "t1"}`)),
			expectedCode:   http.StatusCreated,
			templateClient: &mockConfigTemplateClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("POST", cfgTmplBasePath, testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, testCase.templateClient, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigTemplateGetHandler(t *testing.T) {
	testCases := []struct {
		label          string
		expectedCode   int
		templateClient *mockConfigTemplateClient
	}{
		{
			label:          "Manager error",
			expectedCode:   http.StatusInternalServerError,
			templateClient: &mockConfigTemplateClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:          "Successful get",
			expectedCode:   http.StatusOK,
			templateClient: &mockConfigTemplateClient{template: rb.ConfigTemplate{TemplateName: "t1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", cfgTmplBasePath+"/t1", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, testCase.templateClient, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigTemplateListHandler(t *testing.T) {
	testCases := []struct {
		label          string
		expectedCode   int
		templateClient *mockConfigTemplateClient
	}{
		{
			label:          "Manager error",
			expectedCode:   http.StatusInternalServerError,
			templateClient: &mockConfigTemplateClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:          "Successful list",
			expectedCode:   http.StatusOK,
			templateClient: &mockConfigTemplateClient{list: []rb.ConfigTemplateList{{TemplateName: "t1"}}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", cfgTmplBasePath, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, testCase.templateClient, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigTemplateUpdateHandler(t *testing.T) {
	testCases := []struct {
		label          string
		input          io.Reader
		expectedCode   int
		templateClient *mockConfigTemplateClient
	}{
		{
			label:          "Empty body",
			input:          nil,
			expectedCode:   http.StatusBadRequest,
			templateClient: &mockConfigTemplateClient{},
		},
		{
			label:          "Invalid JSON request format",
			input:          bytes.NewBuffer([]byte("invalid")),
			expectedCode:   http.StatusUnprocessableEntity,
			templateClient: &mockConfigTemplateClient{},
		},
		{
			label:          "Missing template name",
			input:          bytes.NewBuffer([]byte(`{"description": "d"}`)),
			expectedCode:   http.StatusBadRequest,
			templateClient: &mockConfigTemplateClient{},
		},
		{
			label:          "Successful update",
			input:          bytes.NewBuffer([]byte(`{"template-name": "t1"}`)),
			expectedCode:   http.StatusCreated,
			templateClient: &mockConfigTemplateClient{template: rb.ConfigTemplate{TemplateName: "t1"}},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("PUT", cfgTmplBasePath+"/t1", testCase.input)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, testCase.templateClient, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}

func TestConfigTemplateDeleteHandler(t *testing.T) {
	testCases := []struct {
		label          string
		expectedCode   int
		templateClient *mockConfigTemplateClient
	}{
		{
			label:          "Manager error",
			expectedCode:   http.StatusInternalServerError,
			templateClient: &mockConfigTemplateClient{err: pkgerrors.New("Internal error")},
		},
		{
			label:          "Successful delete",
			expectedCode:   http.StatusNoContent,
			templateClient: &mockConfigTemplateClient{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", cfgTmplBasePath+"/t1", nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, testCase.templateClient, nil, nil))
			if testCase.expectedCode != resp.StatusCode {
				t.Fatalf("Request method returned: %v and it was expected: %v", resp.StatusCode, testCase.expectedCode)
			}
		})
	}
}
