/*
Copyright 2026 Deutsche Telekom.

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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

func captureAccessLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	t.Cleanup(func() { logrus.SetOutput(os.Stderr) })
	return &buf
}

func lastLogEntry(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("access log line is not JSON: %q: %v", buf.String(), err)
	}
	return entry
}

func TestAccessLogRecordsRequest(t *testing.T) {
	buf := captureAccessLog(t)
	traceID, _ := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	spanID, _ := trace.SpanIDFromHex("00f067aa0ba902b7")
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	}))

	handler := accessLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("created"))
	}))
	req := httptest.NewRequest("POST", "/v1/instance?rb-name=test", nil).WithContext(ctx)
	handler.ServeHTTP(httptest.NewRecorder(), req)

	entry := lastLogEntry(t, buf)
	want := map[string]interface{}{
		"msg":      "HTTP request completed",
		"method":   "POST",
		"path":     "/v1/instance",
		"status":   float64(http.StatusCreated),
		"bytes":    float64(len("created")),
		"trace_id": traceID.String(),
		"span_id":  spanID.String(),
	}
	for key, value := range want {
		if entry[key] != value {
			t.Errorf("%s = %v, want %v", key, entry[key], value)
		}
	}
	if _, ok := entry["duration_ms"]; !ok {
		t.Errorf("duration_ms missing: %v", entry)
	}
}

func TestAccessLogDefaultsStatusToOK(t *testing.T) {
	buf := captureAccessLog(t)

	handler := accessLog(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/v1/healthcheck", nil))

	if status := lastLogEntry(t, buf)["status"]; status != float64(http.StatusOK) {
		t.Errorf("status = %v, want 200", status)
	}
}

func TestRouterLogsUnmatchedRequests(t *testing.T) {
	buf := captureAccessLog(t)

	router := NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/does-not-exist", nil))

	entry := lastLogEntry(t, buf)
	if entry["path"] != "/does-not-exist" || entry["status"] != float64(http.StatusNotFound) {
		t.Errorf("unexpected access log entry %v", entry)
	}
}
