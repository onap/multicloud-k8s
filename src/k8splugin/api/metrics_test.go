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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

// requestCount returns how many requests httpRequestDuration recorded for
// the given labels.
func requestCount(t *testing.T, labels map[string]string) uint64 {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		if family.GetName() != "http_request_duration_seconds" {
			continue
		}
	metrics:
		for _, metric := range family.GetMetric() {
			for _, pair := range metric.GetLabel() {
				if labels[pair.GetName()] != pair.GetValue() {
					continue metrics
				}
			}
			return metric.GetHistogram().GetSampleCount()
		}
	}
	return 0
}

func TestInstrumentRequestsLabelsByRouteTemplate(t *testing.T) {
	router := mux.NewRouter()
	router.Use(instrumentRequests)
	router.HandleFunc("/test/instance/{instID}", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}).Methods("DELETE")
	labels := map[string]string{"method": "delete", "route": "/test/instance/{instID}", "code": "202"}
	before := requestCount(t, labels)

	for _, id := range []string{"a", "b"} {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("DELETE", "/test/instance/"+id, nil))
	}

	if got := requestCount(t, labels) - before; got != 2 {
		t.Errorf("recorded %d requests for %v, want 2", got, labels)
	}
}

func TestRouterRecordsUnmatchedRequests(t *testing.T) {
	labels := map[string]string{"method": "get", "route": "unmatched", "code": "404"}
	before := requestCount(t, labels)

	router := NewRouter(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/does-not-exist", nil))

	if got := requestCount(t, labels) - before; got != 1 {
		t.Errorf("recorded %d requests for %v, want 1", got, labels)
	}
}

func TestWithMetricsEndpoint(t *testing.T) {
	var passedThrough []string
	handler := WithMetricsEndpoint(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		passedThrough = append(passedThrough, r.Method+" "+r.URL.Path)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest("GET", "/metrics", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /metrics returned %d", rec.Code)
	}
	for _, name := range []string{"go_goroutines", "process_cpu_seconds_total"} {
		if !strings.Contains(rec.Body.String(), name) {
			t.Errorf("GET /metrics does not expose %s", name)
		}
	}

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", "/metrics", nil))
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/v1/instance", nil))
	want := []string{"POST /metrics", "GET /v1/instance"}
	if strings.Join(passedThrough, ",") != strings.Join(want, ",") {
		t.Errorf("passed through %v, want %v", passedThrough, want)
	}
}
