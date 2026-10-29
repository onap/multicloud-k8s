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

package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

func TestEnabled(t *testing.T) {
	for value, want := range map[string]bool{"true": true, "false": false, "": false, "TRUE": false} {
		t.Setenv("TRACING_ENABLED", value)
		if got := Enabled(); got != want {
			t.Errorf("TRACING_ENABLED=%q: Enabled() = %v, want %v", value, got, want)
		}
	}
}

func TestSamplerFromEnv(t *testing.T) {
	tests := []struct {
		sampler, arg string
		want         sdktrace.Sampler
	}{
		{"", "", sdktrace.ParentBased(sdktrace.AlwaysSample())},
		{"always_on", "", sdktrace.AlwaysSample()},
		{"always_off", "", sdktrace.NeverSample()},
		{"traceidratio", "0.25", sdktrace.TraceIDRatioBased(0.25)},
		{"traceidratio", "", sdktrace.TraceIDRatioBased(1)},
		{"traceidratio", "2", sdktrace.TraceIDRatioBased(1)},
		{"traceidratio", "not-a-number", sdktrace.TraceIDRatioBased(1)},
		{"parentbased_always_on", "", sdktrace.ParentBased(sdktrace.AlwaysSample())},
		{"parentbased_always_off", "", sdktrace.ParentBased(sdktrace.NeverSample())},
		{"parentbased_traceidratio", "0.1", sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))},
		{"jaeger_remote", "", sdktrace.ParentBased(sdktrace.AlwaysSample())},
	}
	for _, tt := range tests {
		t.Setenv("OTEL_TRACES_SAMPLER", tt.sampler)
		t.Setenv("OTEL_TRACES_SAMPLER_ARG", tt.arg)
		if got, want := samplerFromEnv().Description(), tt.want.Description(); got != want {
			t.Errorf("OTEL_TRACES_SAMPLER=%q OTEL_TRACES_SAMPLER_ARG=%q: got %s, want %s", tt.sampler, tt.arg, got, want)
		}
	}
}

func TestChildTransport(t *testing.T) {
	var traceparent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceparent = r.Header.Get("traceparent")
	}))
	defer server.Close()

	recorder := tracetest.NewSpanRecorder()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())
	})
	client := http.Client{Transport: ChildTransport(http.DefaultTransport)}

	do := func(ctx context.Context) {
		t.Helper()
		req, err := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
	}

	do(context.Background())
	if traceparent != "" || len(recorder.Ended()) != 0 {
		t.Errorf("request without a span was traced: traceparent %q, %d spans", traceparent, len(recorder.Ended()))
	}

	ctx, parent := provider.Tracer("test").Start(context.Background(), "parent")
	do(ctx)
	parent.End()
	if traceparent == "" {
		t.Error("request within a span sent no traceparent")
	}
	spans := recorder.Ended()
	if len(spans) != 2 {
		t.Fatalf("expected the client span and the parent, got %d spans", len(spans))
	}
	if spans[0].Parent().SpanID() != parent.SpanContext().SpanID() {
		t.Error("client span is not a child of the caller's span")
	}
}
