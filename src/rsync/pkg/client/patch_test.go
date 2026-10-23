/*
Copyright 2026 Deutsche Telekom AG
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

package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"
	restfake "k8s.io/client-go/rest/fake"
	"k8s.io/kubectl/pkg/util"
)

var widgetGVK = schema.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Widget"}

func newWidget(replicas int64) *unstructured.Unstructured {
	w := &unstructured.Unstructured{}
	w.SetGroupVersionKind(widgetGVK)
	w.SetNamespace("default")
	w.SetName("w1")
	_ = unstructured.SetNestedField(w.Object, replicas, "spec", "replicas")
	return w
}

// appliedWidget returns the live object as a previous apply of desired
// would have left it, plus the modified configuration for desired.
func appliedWidget(t *testing.T, desired *unstructured.Unstructured) (runtime.Object, []byte) {
	t.Helper()
	modified, err := util.GetModifiedConfiguration(desired, true, unstructured.UnstructuredJSONScheme)
	if err != nil {
		t.Fatalf("GetModifiedConfiguration: %v", err)
	}
	live, _, err := unstructured.UnstructuredJSONScheme.Decode(modified, nil, nil)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return live, modified
}

type recordedRequest struct {
	method, path, contentType string
	body                      []byte
}

func widgetInfo(t *testing.T, desired *unstructured.Unstructured, reply runtime.Object, requests *[]recordedRequest) *resource.Info {
	t.Helper()
	replyBody, err := runtime.Encode(unstructured.UnstructuredJSONScheme, reply)
	if err != nil {
		t.Fatalf("encode reply: %v", err)
	}
	client := &restfake.RESTClient{
		NegotiatedSerializer: resource.UnstructuredPlusDefaultContentConfig().NegotiatedSerializer,
		GroupVersion:         widgetGVK.GroupVersion(),
		Client: restfake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)
			*requests = append(*requests, recordedRequest{req.Method, req.URL.Path, req.Header.Get("Content-Type"), body})
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
				Body:       io.NopCloser(bytes.NewReader(replyBody)),
			}, nil
		}),
	}
	return &resource.Info{
		Client: client,
		Mapping: &meta.RESTMapping{
			GroupVersionKind: widgetGVK,
			Resource:         widgetGVK.GroupVersion().WithResource("widgets"),
			Scope:            meta.RESTScopeNamespace,
		},
		Namespace: "default",
		Name:      "w1",
		Object:    desired,
	}
}

func TestPatchSimpleUnregisteredKindUnchangedIsNoop(t *testing.T) {
	desired := newWidget(1)
	live, modified := appliedWidget(t, desired)
	var requests []recordedRequest
	info := widgetInfo(t, desired, live, &requests)

	patch, obj, err := patchSimple(live, modified, info)
	if err != nil {
		t.Fatalf("patchSimple: %v", err)
	}
	if string(patch) != "{}" {
		t.Errorf("patch = %s, want {}", patch)
	}
	if obj != live {
		t.Errorf("want the live object returned unchanged")
	}
	if len(requests) != 0 {
		t.Errorf("want no API request, got %d", len(requests))
	}
}

func TestPatchSimpleUnregisteredKindSendsMergePatch(t *testing.T) {
	live, _ := appliedWidget(t, newWidget(1))
	desired := newWidget(3)
	_, modified := appliedWidget(t, desired)
	var requests []recordedRequest
	info := widgetInfo(t, desired, desired, &requests)

	patch, _, err := patchSimple(live, modified, info)
	if err != nil {
		t.Fatalf("patchSimple: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("want 1 API request, got %d", len(requests))
	}
	req := requests[0]
	if req.method != http.MethodPatch || req.path != "/namespaces/default/widgets/w1" {
		t.Errorf("request = %s %s, want PATCH /namespaces/default/widgets/w1", req.method, req.path)
	}
	if req.contentType != string(types.MergePatchType) {
		t.Errorf("Content-Type = %q, want %q", req.contentType, types.MergePatchType)
	}
	if !bytes.Equal(req.body, patch) {
		t.Errorf("sent body %s differs from returned patch %s", req.body, patch)
	}

	var got map[string]interface{}
	if err := json.Unmarshal(patch, &got); err != nil {
		t.Fatalf("patch is not JSON: %v", err)
	}
	if replicas, _, _ := unstructured.NestedFieldNoCopy(got, "spec", "replicas"); replicas != float64(3) {
		t.Errorf("patch spec.replicas = %v, want 3 (patch %s)", replicas, patch)
	}
	annotation, _, _ := unstructured.NestedString(got, "metadata", "annotations", "kubectl.kubernetes.io/last-applied-configuration")
	if annotation == "" {
		t.Errorf("patch does not update the last-applied annotation: %s", patch)
	}
	if len(got) != 2 {
		t.Errorf("patch touches more than metadata and spec: %s", patch)
	}
}

func TestPatchSimpleUnregisteredKindRejectsRename(t *testing.T) {
	live, _ := appliedWidget(t, newWidget(1))
	desired := newWidget(1)
	desired.SetName("w2")
	_, modified := appliedWidget(t, desired)
	var requests []recordedRequest
	info := widgetInfo(t, desired, desired, &requests)

	if _, _, err := patchSimple(live, modified, info); err == nil {
		t.Fatal("want precondition error when the name changes")
	}
	if len(requests) != 0 {
		t.Errorf("want no API request, got %d", len(requests))
	}
}
