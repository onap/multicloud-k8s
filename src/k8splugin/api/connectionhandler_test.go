/*
 * Copyright 2026 Deutsche Telekom
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
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/connection"
	pkgerrors "github.com/pkg/errors"
)

// mockConnection implements connection.ConnectionManager for testing.
type mockConnection struct {
	connection.ConnectionManager
	Items []connection.Connection
	Err   error
}

func (m *mockConnection) Create(_ context.Context, c connection.Connection) (connection.Connection, error) {
	if m.Err != nil {
		return connection.Connection{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockConnection) Get(_ context.Context, name string) (connection.Connection, error) {
	if m.Err != nil {
		return connection.Connection{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockConnection) Delete(_ context.Context, name string) error {
	return m.Err
}

func (m *mockConnection) GetConnectivityRecordByName(_ context.Context, connname, name string) (map[string]string, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return map[string]string{}, nil
}

// buildMultipartRequest builds a multipart/form-data POST request for /v1/connectivity-info.
func buildMultipartRequest(metadata, fileContent string) (*http.Request, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("metadata", metadata); err != nil {
		return nil, err
	}

	part, err := writer.CreateFormFile("file", "kubeconfig")
	if err != nil {
		return nil, err
	}
	if _, err = io.WriteString(part, fileContent); err != nil {
		return nil, err
	}

	writer.Close()

	req := httptest.NewRequest("POST", "/v1/connectivity-info", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

func TestConnectionCreateHandler(t *testing.T) {
	testCases := []struct {
		label            string
		useMultipart     bool
		metadata         string
		fileContent      string
		reader           io.Reader
		contentType      string
		expected         connection.Connection
		expectedCode     int
		connectionClient *mockConnection
	}{
		{
			label:            "Missing Multipart Body",
			reader:           bytes.NewBuffer([]byte("")),
			contentType:      "application/json",
			expectedCode:     http.StatusUnprocessableEntity,
			connectionClient: &mockConnection{},
		},
		{
			label:            "Missing Cloud Region",
			useMultipart:     true,
			metadata:         `{"cloud-owner":"test-owner"}`,
			fileContent:      "kubeconfig-content",
			expectedCode:     http.StatusBadRequest,
			connectionClient: &mockConnection{},
		},
		{
			label:            "Missing Cloud Owner",
			useMultipart:     true,
			metadata:         `{"cloud-region":"test-region"}`,
			fileContent:      "kubeconfig-content",
			expectedCode:     http.StatusBadRequest,
			connectionClient: &mockConnection{},
		},
		{
			label:        "Successful Connection Creation",
			useMultipart: true,
			metadata:     `{"cloud-region":"test-region","cloud-owner":"test-owner"}`,
			fileContent:  "kubeconfig-content",
			expectedCode: http.StatusCreated,
			expected: connection.Connection{
				CloudRegion: "test-region",
				CloudOwner:  "test-owner",
			},
			connectionClient: &mockConnection{
				Items: []connection.Connection{
					{
						CloudRegion: "test-region",
						CloudOwner:  "test-owner",
					},
				},
			},
		},
		{
			label:        "Backend Create Error",
			useMultipart: true,
			metadata:     `{"cloud-region":"test-region","cloud-owner":"test-owner"}`,
			fileContent:  "kubeconfig-content",
			expectedCode: http.StatusInternalServerError,
			connectionClient: &mockConnection{
				Err: pkgerrors.New("DB error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			var request *http.Request
			if testCase.useMultipart {
				req, err := buildMultipartRequest(testCase.metadata, testCase.fileContent)
				if err != nil {
					t.Fatalf("Failed to build multipart request: %s", err)
				}
				request = req
			} else {
				request = httptest.NewRequest("POST", "/v1/connectivity-info", testCase.reader)
				request.Header.Set("Content-Type", testCase.contentType)
			}

			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, testCase.connectionClient, nil, nil, nil))

			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusCreated {
				got := connection.Connection{}
				json.NewDecoder(resp.Body).Decode(&got)
				if !reflect.DeepEqual(testCase.expected, got) {
					t.Errorf("createHandler returned unexpected body: got %v; expected %v",
						got, testCase.expected)
				}
			}
		})
	}
}

func TestConnectionGetHandler(t *testing.T) {
	testCases := []struct {
		label            string
		name             string
		expected         connection.Connection
		expectedCode     int
		connectionClient *mockConnection
	}{
		{
			label:        "Get Existing Connection",
			name:         "test-region",
			expectedCode: http.StatusOK,
			expected: connection.Connection{
				CloudRegion: "test-region",
				CloudOwner:  "test-owner",
			},
			connectionClient: &mockConnection{
				Items: []connection.Connection{
					{
						CloudRegion: "test-region",
						CloudOwner:  "test-owner",
					},
				},
			},
		},
		{
			label:        "Get Non-Existing Connection",
			name:         "non-existing",
			expectedCode: http.StatusNotFound,
			connectionClient: &mockConnection{
				Err: pkgerrors.New("Get Connection: Error finding master table: mongo: no documents in result"),
			},
		},
		{
			label:        "Backend Get Error",
			name:         "test-region",
			expectedCode: http.StatusInternalServerError,
			connectionClient: &mockConnection{
				Err: pkgerrors.New("DB error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/v1/connectivity-info/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, testCase.connectionClient, nil, nil, nil))

			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}

			if resp.StatusCode == http.StatusOK {
				got := connection.Connection{}
				json.NewDecoder(resp.Body).Decode(&got)
				if !reflect.DeepEqual(testCase.expected, got) {
					t.Errorf("getHandler returned unexpected body: got %v; expected %v",
						got, testCase.expected)
				}
			}
		})
	}
}

func TestConnectionDeleteHandler(t *testing.T) {
	testCases := []struct {
		label            string
		name             string
		expectedCode     int
		connectionClient *mockConnection
	}{
		{
			label:            "Delete Existing Connection",
			name:             "test-region",
			expectedCode:     http.StatusNoContent,
			connectionClient: &mockConnection{},
		},
		{
			label:        "Delete Non-Existing Connection",
			name:         "non-existing",
			expectedCode: http.StatusNotFound,
			connectionClient: &mockConnection{
				Err: pkgerrors.New("Delete Connection: Error finding master table: mongo: no documents in result"),
			},
		},
		{
			label:        "Backend Delete Error",
			name:         "test-region",
			expectedCode: http.StatusInternalServerError,
			connectionClient: &mockConnection{
				Err: pkgerrors.New("DB error"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			request := httptest.NewRequest("DELETE", "/v1/connectivity-info/"+testCase.name, nil)
			resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, testCase.connectionClient, nil, nil, nil))

			if resp.StatusCode != testCase.expectedCode {
				t.Fatalf("Expected %d; Got: %d", testCase.expectedCode, resp.StatusCode)
			}
		})
	}
}
