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

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "http_request_duration_seconds",
	Help:    "Duration of HTTP requests by method, route template and status code.",
	Buckets: prometheus.DefBuckets,
}, []string{"method", "route", "code"})

func init() {
	prometheus.MustRegister(httpRequestDuration)
}

// instrumentRequests records httpRequestDuration labelled with the route
// template, so path parameters do not create new series.
func instrumentRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := "unmatched"
		if current := mux.CurrentRoute(r); current != nil {
			if template, err := current.GetPathTemplate(); err == nil {
				route = template
			}
		}
		observer := httpRequestDuration.MustCurryWith(prometheus.Labels{"route": route})
		promhttp.InstrumentHandlerDuration(observer, next).ServeHTTP(w, r)
	})
}

// WithMetricsEndpoint serves the Prometheus metrics on /metrics and passes
// every other request to h. It sits in front of the router so scrapes are
// neither traced nor access-logged.
func WithMetricsEndpoint(h http.Handler) http.Handler {
	metrics := promhttp.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" && r.Method == http.MethodGet {
			metrics.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}
