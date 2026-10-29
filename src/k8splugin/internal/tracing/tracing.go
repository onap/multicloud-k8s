/*
Copyright 2025 Deutsche Telekom.

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
	"os"
	"strconv"

	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	"go.mongodb.org/mongo-driver/event"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/trace"
)

type Instrumentation struct {
	TracerProvider *sdktrace.TracerProvider
	Tracer         trace.Tracer
}

func initTracerProvider() *sdktrace.TracerProvider {
	exporter, err := otlptracegrpc.New(context.Background())

	if err != nil {
		panic(err)
	}

	resources, err := resource.New(context.Background(),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithProcess(),
		resource.WithOS(),
		resource.WithHost())

	if err != nil {
		panic(err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resources),
		sdktrace.WithSampler(samplerFromEnv()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}, b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))))

	return tp
}

// Enabled reports whether tracing is switched on with TRACING_ENABLED=true.
func Enabled() bool {
	return os.Getenv("TRACING_ENABLED") == "true"
}

// samplerFromEnv reads OTEL_TRACES_SAMPLER and OTEL_TRACES_SAMPLER_ARG as
// defined by the OpenTelemetry specification. The SDK version in use does
// not read them itself.
func samplerFromEnv() sdktrace.Sampler {
	name := os.Getenv("OTEL_TRACES_SAMPLER")
	arg := os.Getenv("OTEL_TRACES_SAMPLER_ARG")
	ratio := 1.0
	if arg != "" {
		r, err := strconv.ParseFloat(arg, 64)
		if err != nil || r < 0 || r > 1 {
			log.Warn("Invalid OTEL_TRACES_SAMPLER_ARG, using 1.0", log.Fields{"value": arg})
		} else {
			ratio = r
		}
	}

	switch name {
	case "always_on":
		return sdktrace.AlwaysSample()
	case "always_off":
		return sdktrace.NeverSample()
	case "traceidratio":
		return sdktrace.TraceIDRatioBased(ratio)
	case "parentbased_always_off":
		return sdktrace.ParentBased(sdktrace.NeverSample())
	case "parentbased_traceidratio":
		return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))
	case "", "parentbased_always_on":
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	default:
		log.Warn("Unknown OTEL_TRACES_SAMPLER, using parentbased_always_on", log.Fields{"value": name})
		return sdktrace.ParentBased(sdktrace.AlwaysSample())
	}
}

// ChildTransport traces requests that are made within an existing span and
// passes all others through untraced. Most Kubernetes API calls are made
// with context.TODO(), and tracing those would start a new trace per call.
func ChildTransport(rt http.RoundTripper) http.RoundTripper {
	return childTransport{traced: otelhttp.NewTransport(rt), plain: rt}
}

type childTransport struct {
	traced http.RoundTripper
	plain  http.RoundTripper
}

func (t childTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if trace.SpanContextFromContext(r.Context()).IsValid() {
		return t.traced.RoundTrip(r)
	}
	return t.plain.RoundTrip(r)
}

// ChildCommandMonitor is the Mongo counterpart of ChildTransport: it only
// passes commands issued within an existing span on to m, so background work
// such as the database health check does not start a trace per command.
func ChildCommandMonitor(m *event.CommandMonitor) *event.CommandMonitor {
	return &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			if trace.SpanContextFromContext(ctx).IsValid() {
				m.Started(ctx, evt)
			}
		},
		// Ending a command that m never started is a no-op, so these need no
		// guard of their own and cannot leave a started span open.
		Succeeded: m.Succeeded,
		Failed:    m.Failed,
	}
}

func NewInstrumentation() *Instrumentation {
	tp := initTracerProvider()

	return &Instrumentation{
		TracerProvider: tp,
		Tracer:         tp.Tracer("multicloud-k8s"),
	}
}
