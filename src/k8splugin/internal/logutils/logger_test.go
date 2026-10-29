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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	stdlog "log"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/klog/v2"
)

// captureLog redirects logrus to a buffer for the duration of the test and
// returns a function that decodes every line written so far.
func captureLog(t *testing.T) func() []map[string]interface{} {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return func() []map[string]interface{} {
		var entries []map[string]interface{}
		for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
			if line == "" {
				continue
			}
			var entry map[string]interface{}
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				t.Fatalf("log line is not JSON: %q: %v", line, err)
			}
			entries = append(entries, entry)
		}
		return entries
	}
}

func spanContext(t *testing.T) (context.Context, trace.SpanContext) {
	t.Helper()
	traceID, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	if err != nil {
		t.Fatal(err)
	}
	spanID, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	if err != nil {
		t.Fatal(err)
	}
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	return trace.ContextWithSpanContext(context.Background(), sc), sc
}

func TestInfoWritesJSONWithFields(t *testing.T) {
	entries := captureLog(t)

	Info("instance created", Fields{"instance": "inst-1"})

	got := entries()
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0]["msg"] != "instance created" || got[0]["level"] != "info" || got[0]["instance"] != "inst-1" {
		t.Errorf("unexpected entry %v", got[0])
	}
	if _, ok := got[0]["trace_id"]; ok {
		t.Errorf("entry without context must not carry trace_id: %v", got[0])
	}
}

func TestContextVariantsAddTraceAndSpanID(t *testing.T) {
	ctx, sc := spanContext(t)
	for level, logFn := range map[string]func(context.Context, string, Fields){
		"info":    InfoContext,
		"warning": WarnContext,
		"error":   ErrorContext,
	} {
		t.Run(level, func(t *testing.T) {
			entries := captureLog(t)

			logFn(ctx, "message", Fields{"key": "value"})

			got := entries()
			if len(got) != 1 {
				t.Fatalf("expected 1 entry, got %d", len(got))
			}
			if got[0]["level"] != level {
				t.Errorf("level = %v, want %s", got[0]["level"], level)
			}
			if got[0]["trace_id"] != sc.TraceID().String() || got[0]["span_id"] != sc.SpanID().String() {
				t.Errorf("trace ids missing or wrong: %v", got[0])
			}
			if got[0]["key"] != "value" {
				t.Errorf("fields lost: %v", got[0])
			}
		})
	}
}

func TestContextWithoutSpanAddsNoTraceID(t *testing.T) {
	entries := captureLog(t)

	InfoContext(context.Background(), "message", nil)

	got := entries()
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if _, ok := got[0]["trace_id"]; ok {
		t.Errorf("unexpected trace_id: %v", got[0])
	}
}

func TestStandardLibraryLogIsJSON(t *testing.T) {
	entries := captureLog(t)

	stdlog.Printf("from %s", "a library")

	got := entries()
	if len(got) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(got))
	}
	if got[0]["msg"] != "from a library" || got[0]["level"] != "info" {
		t.Errorf("unexpected entry %v", got[0])
	}
}

func TestKlogIsJSON(t *testing.T) {
	entries := captureLog(t)

	klog.Info("client-go info")
	klog.Error("client-go error")
	klog.InfoS("structured", "pod", "kube-system/coredns")
	klog.ErrorS(errors.New("boom"), "structured error", "pod", "kube-system/coredns")

	got := entries()
	if len(got) != 4 {
		t.Fatalf("expected 4 entries, got %d: %v", len(got), got)
	}
	want := []struct{ level, msg string }{
		{"info", "client-go info"},
		{"error", "client-go error"},
		{"info", "structured"},
		{"error", "structured error"},
	}
	for i, w := range want {
		if got[i]["level"] != w.level || got[i]["msg"] != w.msg {
			t.Errorf("entry %d = %v, want level %s msg %q", i, got[i], w.level, w.msg)
		}
	}
	if got[2]["pod"] != "kube-system/coredns" {
		t.Errorf("structured key/values lost: %v", got[2])
	}
	if got[3]["error"] != "boom" {
		t.Errorf("error lost: %v", got[3])
	}
}
