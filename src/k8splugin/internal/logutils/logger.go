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

package logutils

import (
	"context"
	stdlog "log"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/klog/v2"
)

// Fields is type that will be used by the calling function
type Fields map[string]interface{}

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
	log.AddHook(traceHook{})

	// Libraries log through the standard library and klog; route both
	// through logrus so every line on stdout is JSON.
	stdlog.SetFlags(0)
	stdlog.SetOutput(stdlogWriter{})
	klog.SetLogger(klogAdapter{})
}

// Error uses the fields provided and logs
func Error(msg string, fields Fields) {
	log.WithFields(log.Fields(fields)).Error(msg)
}

// Warn uses the fields provided and logs
func Warn(msg string, fields Fields) {
	log.WithFields(log.Fields(fields)).Warn(msg)
}

// Info uses the fields provided and logs
func Info(msg string, fields Fields) {
	log.WithFields(log.Fields(fields)).Info(msg)
}

// Fatal uses the fields provided, logs and exits with status 1
func Fatal(msg string, fields Fields) {
	log.WithFields(log.Fields(fields)).Fatal(msg)
}

// ErrorContext is Error with the trace and span id of ctx added
func ErrorContext(ctx context.Context, msg string, fields Fields) {
	log.WithContext(ctx).WithFields(log.Fields(fields)).Error(msg)
}

// WarnContext is Warn with the trace and span id of ctx added
func WarnContext(ctx context.Context, msg string, fields Fields) {
	log.WithContext(ctx).WithFields(log.Fields(fields)).Warn(msg)
}

// InfoContext is Info with the trace and span id of ctx added
func InfoContext(ctx context.Context, msg string, fields Fields) {
	log.WithContext(ctx).WithFields(log.Fields(fields)).Info(msg)
}

// traceHook adds trace_id and span_id to entries logged with a context
// that carries a valid span.
type traceHook struct{}

func (traceHook) Levels() []log.Level {
	return log.AllLevels
}

func (traceHook) Fire(e *log.Entry) error {
	if e.Context == nil {
		return nil
	}
	sc := trace.SpanContextFromContext(e.Context)
	if !sc.IsValid() {
		return nil
	}
	e.Data["trace_id"] = sc.TraceID().String()
	e.Data["span_id"] = sc.SpanID().String()
	return nil
}

type stdlogWriter struct{}

func (stdlogWriter) Write(p []byte) (int, error) {
	log.Info(strings.TrimRight(string(p), "\n"))
	return len(p), nil
}
