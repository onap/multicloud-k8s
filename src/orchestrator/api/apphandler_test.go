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

package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"

	pkgerrors "github.com/pkg/errors"
)

type mockAppManager struct {
	Items        []moduleLib.App
	ItemsContent moduleLib.AppContent
	Err          error
}

func (m *mockAppManager) CreateApp(a moduleLib.App, ac moduleLib.AppContent, p, cN, cV string) (moduleLib.App, error) {
	if m.Err != nil {
		return moduleLib.App{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockAppManager) GetApp(name, p, cN, cV string) (moduleLib.App, error) {
	if m.Err != nil {
		return moduleLib.App{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockAppManager) GetAppContent(name, p, cN, cV string) (moduleLib.AppContent, error) {
	if m.Err != nil {
		return moduleLib.AppContent{}, m.Err
	}
	return m.ItemsContent, nil
}

func (m *mockAppManager) GetApps(p, cN, cV string) ([]moduleLib.App, error) {
	if m.Err != nil {
		return []moduleLib.App{}, m.Err
	}
	return m.Items, nil
}

func (m *mockAppManager) DeleteApp(name, p, cN, cV string) error {
	return m.Err
}

func init() {
	appJSONFile = "../json-schemas/metadata.json"
}

// makeTarGz produces a minimal valid gzip-compressed tar archive.
func makeTarGz(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)
	content := []byte("hello")
	hdr := &tar.Header{Name: "file.txt", Mode: 0600, Size: int64(len(content))}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("failed writing tar header: %s", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("failed writing tar content: %s", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("failed closing tar writer: %s", err)
	}
	if err := gzw.Close(); err != nil {
		t.Fatalf("failed closing gzip writer: %s", err)
	}
	return buf.Bytes()
}

// buildAppMultipart builds a multipart body carrying a metadata JSON part and a
// file part with the given content.
func buildAppMultipart(t *testing.T, metadata string, fileContent []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if err := writer.WriteField("metadata", metadata); err != nil {
		t.Fatalf("failed writing metadata field: %s", err)
	}
	part, err := writer.CreateFormFile("file", "file.tar.gz")
	if err != nil {
		t.Fatalf("failed creating file part: %s", err)
	}
	if _, err := part.Write(fileContent); err != nil {
		t.Fatalf("failed writing file part: %s", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed closing multipart writer: %s", err)
	}
	return body, writer.FormDataContentType()
}

func TestAppCreateHandler(t *testing.T) {
	url := "/v2/projects/testProject/composite-apps/testCompositeApp/v1/apps"

	t.Run("Create App", func(t *testing.T) {
		body, contentType := buildAppMultipart(t,
			`{"metadata":{"name":"testApp","description":"a test app"}}`, makeTarGz(t))
		request := httptest.NewRequest("POST", url, body)
		request.Header.Set("Content-Type", contentType)
		client := &mockAppManager{Items: []moduleLib.App{
			{Metadata: moduleLib.AppMetaData{Name: "testApp"}},
		}}
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected %d; Got: %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("Bad metadata JSON", func(t *testing.T) {
		body, contentType := buildAppMultipart(t, `{"metadata": }`, makeTarGz(t))
		request := httptest.NewRequest("POST", url, body)
		request.Header.Set("Content-Type", contentType)
		client := &mockAppManager{}
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("Expected %d; Got: %d", http.StatusUnprocessableEntity, resp.StatusCode)
		}
	})

	t.Run("Invalid file format", func(t *testing.T) {
		body, contentType := buildAppMultipart(t,
			`{"metadata":{"name":"testApp"}}`, []byte("not a tar gz"))
		request := httptest.NewRequest("POST", url, body)
		request.Header.Set("Content-Type", contentType)
		client := &mockAppManager{}
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Fatalf("Expected %d; Got: %d", http.StatusUnprocessableEntity, resp.StatusCode)
		}
	})

	t.Run("Manager Error", func(t *testing.T) {
		body, contentType := buildAppMultipart(t,
			`{"metadata":{"name":"testApp"}}`, makeTarGz(t))
		request := httptest.NewRequest("POST", url, body)
		request.Header.Set("Content-Type", contentType)
		client := &mockAppManager{Err: pkgerrors.New("Internal Error")}
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestAppGetHandler(t *testing.T) {
	t.Run("Get App as JSON", func(t *testing.T) {
		client := &mockAppManager{
			Items:        []moduleLib.App{{Metadata: moduleLib.AppMetaData{Name: "testApp"}}},
			ItemsContent: moduleLib.AppContent{FileContent: ""},
		}
		request := httptest.NewRequest("GET", "/v2/projects/testProject/composite-apps/testCompositeApp/v1/apps/testApp", nil)
		request.Header.Set("Accept", "application/json")
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Get App Manager Error", func(t *testing.T) {
		client := &mockAppManager{Err: pkgerrors.New("Internal Error")}
		request := httptest.NewRequest("GET", "/v2/projects/testProject/composite-apps/testCompositeApp/v1/apps/testApp", nil)
		request.Header.Set("Accept", "application/json")
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestAppGetAllHandler(t *testing.T) {
	t.Run("Get All Apps", func(t *testing.T) {
		client := &mockAppManager{
			Items: []moduleLib.App{{Metadata: moduleLib.AppMetaData{Name: "app1"}}},
		}
		request := httptest.NewRequest("GET", "/v2/projects/testProject/composite-apps/testCompositeApp/v1/apps", nil)
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Get All Apps Manager Error", func(t *testing.T) {
		client := &mockAppManager{Err: pkgerrors.New("Internal Error")}
		request := httptest.NewRequest("GET", "/v2/projects/testProject/composite-apps/testCompositeApp/v1/apps", nil)
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}

func TestAppDeleteHandler(t *testing.T) {
	t.Run("Delete App", func(t *testing.T) {
		client := &mockAppManager{}
		request := httptest.NewRequest("DELETE", "/v2/projects/testProject/composite-apps/testCompositeApp/v1/apps/testApp", nil)
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected %d; Got: %d", http.StatusNoContent, resp.StatusCode)
		}
	})

	t.Run("Delete App Manager Error", func(t *testing.T) {
		client := &mockAppManager{Err: pkgerrors.New("Internal Error")}
		request := httptest.NewRequest("DELETE", "/v2/projects/testProject/composite-apps/testCompositeApp/v1/apps/testApp", nil)
		resp := executeRequest(request, NewRouter(nil, nil, client, nil, nil, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}
