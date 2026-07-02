/*
 * Copyright 2026 Deutsche Telekom AG
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

package status

import (
	"strings"
	"testing"

	yaml "github.com/ghodss/yaml"
	v1alpha1 "github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"
	appcontext "github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
)

func TestGetStatusCR(t *testing.T) {
	label := "deployment-id-1"
	y, err := GetStatusCR(label)
	if err != nil {
		t.Fatalf("GetStatusCR returned an error: %s", err)
	}

	// Decode the produced YAML back into a ResourceBundleState and assert on the
	// meaningful fields rather than on the raw byte layout.
	var cr v1alpha1.ResourceBundleState
	if err := yaml.Unmarshal(y, &cr); err != nil {
		t.Fatalf("Failed to unmarshal produced YAML: %s", err)
	}

	if cr.APIVersion != "k8splugin.io/v1alpha1" {
		t.Fatalf("Expected apiVersion 'k8splugin.io/v1alpha1', got %q", cr.APIVersion)
	}
	if cr.Kind != "ResourceBundleState" {
		t.Fatalf("Expected kind 'ResourceBundleState', got %q", cr.Kind)
	}
	if cr.GetName() != label {
		t.Fatalf("Expected name %q, got %q", label, cr.GetName())
	}
	if cr.GetLabels()["emco/deployment-id"] != label {
		t.Fatalf("Expected label emco/deployment-id=%q, got %q", label, cr.GetLabels()["emco/deployment-id"])
	}
	if cr.Spec.Selector == nil || cr.Spec.Selector.MatchLabels["emco/deployment-id"] != label {
		t.Fatalf("Expected selector matchLabels emco/deployment-id=%q", label)
	}

	// The produced document must be valid YAML with the correct casing for the
	// apiVersion key (the reason the code marshals via JSON first).
	if !strings.Contains(string(y), "apiVersion: k8splugin.io/v1alpha1") {
		t.Fatalf("Expected 'apiVersion:' key in YAML, got:\n%s", string(y))
	}
}

func TestHandleStatusUpdate(t *testing.T) {
	// Build a composite app with one app/cluster so HandleStatusUpdate has a
	// cluster handle to attach status to.
	contextdb.Db = &contextdb.MockEtcd{}
	ac := appcontext.AppContext{}
	cid, err := ac.InitAppContext()
	if err != nil {
		t.Fatalf("InitAppContext failed: %s", err)
	}
	ctxID := cid.(string)
	rootHdl, err := ac.CreateCompositeApp()
	if err != nil {
		t.Fatalf("CreateCompositeApp failed: %s", err)
	}
	appHdl, err := ac.AddApp(rootHdl, "app1")
	if err != nil {
		t.Fatalf("AddApp failed: %s", err)
	}
	clusterID := "provider1+cluster1"
	if _, err := ac.AddCluster(appHdl, clusterID); err != nil {
		t.Fatalf("AddCluster failed: %s", err)
	}

	rbState := &v1alpha1.ResourceBundleState{}
	rbState.Status.Ready = true
	rbState.Status.ResourceCount = 3

	t.Run("Writes status onto the cluster handle for a valid label", func(t *testing.T) {
		// Label format is "<contextId>-<app>".
		label := ctxID + "-app1"
		HandleStatusUpdate(clusterID, label, rbState)

		chandle, err := ac.GetClusterHandle("app1", clusterID)
		if err != nil {
			t.Fatalf("GetClusterHandle failed: %s", err)
		}
		sh, err := ac.GetLevelHandle(chandle, "status")
		if err != nil {
			t.Fatalf("Expected a status handle to have been created: %s", err)
		}
		v, err := ac.GetValue(sh)
		if err != nil {
			t.Fatalf("GetValue(status) failed: %s", err)
		}
		s, ok := v.(string)
		if !ok || !strings.Contains(s, "\"resourceCount\":3") {
			t.Fatalf("Expected stored status to contain the marshalled ResourceCount, got %v", v)
		}
	})

	t.Run("Ignores a label missing the appcontext identifier", func(t *testing.T) {
		// Leading separator -> empty context id; must return without panicking.
		HandleStatusUpdate(clusterID, "-app1", rbState)
	})

	t.Run("Ignores a label missing the app identifier", func(t *testing.T) {
		HandleStatusUpdate(clusterID, ctxID+"-", rbState)
	})

	t.Run("Ignores a label with an unknown context id", func(t *testing.T) {
		HandleStatusUpdate(clusterID, "0000000000-app1", rbState)
	})
}
