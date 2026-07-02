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

package context

import (
	"encoding/json"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/resourcestatus"
)

// testAppContext bundles an AppContext with the identifiers needed to reach
// into the tree it was built from.
type testAppContext struct {
	ac      appcontext.AppContext
	cid     string
	rootHdl interface{}
}

// appOrder is the app instruction value shape consumed by initializeResourceStatus.
type appOrder struct {
	Apporder []string `json:"apporder"`
}

// resOrder is the resource instruction value shape.
type resOrder struct {
	Resorder []string `json:"resorder"`
}

// newMockContext installs a fresh in-memory contextdb and returns an empty,
// initialized AppContext plus its root handle and context id.
func newMockContext(t *testing.T) testAppContext {
	t.Helper()
	contextdb.Db = &contextdb.MockEtcd{}

	ac := appcontext.AppContext{}
	cid, err := ac.InitAppContext()
	if err != nil {
		t.Fatalf("InitAppContext failed: %s", err)
	}
	rootHdl, err := ac.CreateCompositeApp()
	if err != nil {
		t.Fatalf("CreateCompositeApp failed: %s", err)
	}
	return testAppContext{ac: ac, cid: cid.(string), rootHdl: rootHdl}
}

// addResource adds a resource under app/cluster with the given manifest value and
// an initial status, returning the resource handle.
func (tc *testAppContext) addResource(t *testing.T, appHdl interface{}, cluster, resName, value string, status resourcestatus.RsyncStatus) interface{} {
	t.Helper()
	clusterHdl, err := tc.ac.AddCluster(appHdl, cluster)
	if err != nil {
		t.Fatalf("AddCluster failed: %s", err)
	}
	resHdl, err := tc.ac.AddResource(clusterHdl, resName, value)
	if err != nil {
		t.Fatalf("AddResource failed: %s", err)
	}
	if _, err := tc.ac.AddLevelValue(resHdl, "status", resourcestatus.ResourceStatus{Status: status}); err != nil {
		t.Fatalf("AddLevelValue(status) failed: %s", err)
	}
	return resHdl
}

func TestGetRes(t *testing.T) {
	tc := newMockContext(t)
	appHdl, err := tc.ac.AddApp(tc.rootHdl, "app1")
	if err != nil {
		t.Fatalf("AddApp failed: %s", err)
	}
	tc.addResource(t, appHdl, "provider1+cluster1", "res1", "manifest-bytes", resourcestatus.RsyncStatusEnum.Pending)

	t.Run("Returns resource bytes for an existing resource", func(t *testing.T) {
		b, sh, err := getRes(tc.ac, "res1", "app1", "provider1+cluster1")
		if err != nil {
			t.Fatalf("getRes returned an error: %s", err)
		}
		if string(b) != "manifest-bytes" {
			t.Fatalf("Expected 'manifest-bytes', got %q", string(b))
		}
		if sh == nil {
			t.Fatal("Expected a non-nil status handle")
		}
	})

	t.Run("Returns an error for a missing resource", func(t *testing.T) {
		_, _, err := getRes(tc.ac, "nope", "app1", "provider1+cluster1")
		if err == nil {
			t.Fatal("Expected an error for a missing resource, got nil")
		}
	})
}

func TestUpdateResourceStatus(t *testing.T) {
	tc := newMockContext(t)
	appHdl, _ := tc.ac.AddApp(tc.rootHdl, "app1")
	cluster := "provider1+cluster1"
	// res1 is Pending (updatable); res2 is already Applied (a 'done' status).
	tc.addResource(t, appHdl, cluster, "res1", "m1", resourcestatus.RsyncStatusEnum.Pending)

	// Second resource shares the same cluster handle, so add it directly.
	clusterHdl, _ := tc.ac.GetClusterHandle("app1", cluster)
	res2Hdl, _ := tc.ac.AddResource(clusterHdl, "res2", "m2")
	tc.ac.AddLevelValue(res2Hdl, "status", resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied})

	aov := map[string][]string{"resorder": {"res1", "res2"}}
	newStatus := resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Applied}
	if err := updateResourceStatus(tc.ac, newStatus, "app1", cluster, aov); err != nil {
		t.Fatalf("updateResourceStatus returned an error: %s", err)
	}

	// res1 (Pending) should now be Applied.
	if got := readResourceStatus(t, tc.ac, "app1", cluster, "res1"); got != resourcestatus.RsyncStatusEnum.Applied {
		t.Fatalf("Expected res1 to be Applied, got %q", got)
	}
	// res2 was already Applied - it should be untouched (skip-done branch).
	if got := readResourceStatus(t, tc.ac, "app1", cluster, "res2"); got != resourcestatus.RsyncStatusEnum.Applied {
		t.Fatalf("Expected res2 to remain Applied, got %q", got)
	}
}

func TestAllResourcesDone(t *testing.T) {
	cluster := "provider1+cluster1"
	aov := map[string][]string{"resorder": {"res1", "res2"}}

	t.Run("True when every resource has reached a done status", func(t *testing.T) {
		tc := newMockContext(t)
		appHdl, _ := tc.ac.AddApp(tc.rootHdl, "app1")
		tc.addResource(t, appHdl, cluster, "res1", "m1", resourcestatus.RsyncStatusEnum.Applied)
		clusterHdl, _ := tc.ac.GetClusterHandle("app1", cluster)
		res2Hdl, _ := tc.ac.AddResource(clusterHdl, "res2", "m2")
		tc.ac.AddLevelValue(res2Hdl, "status", resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Deleted})

		if !allResourcesDone(tc.ac, "app1", cluster, aov) {
			t.Fatal("Expected allResourcesDone to be true")
		}
	})

	t.Run("False when a resource is still pending", func(t *testing.T) {
		tc := newMockContext(t)
		appHdl, _ := tc.ac.AddApp(tc.rootHdl, "app1")
		tc.addResource(t, appHdl, cluster, "res1", "m1", resourcestatus.RsyncStatusEnum.Applied)
		clusterHdl, _ := tc.ac.GetClusterHandle("app1", cluster)
		res2Hdl, _ := tc.ac.AddResource(clusterHdl, "res2", "m2")
		tc.ac.AddLevelValue(res2Hdl, "status", resourcestatus.ResourceStatus{Status: resourcestatus.RsyncStatusEnum.Pending})

		if allResourcesDone(tc.ac, "app1", cluster, aov) {
			t.Fatal("Expected allResourcesDone to be false")
		}
	})
}

func TestInitializeAndGetAppContextStatus(t *testing.T) {
	tc := newMockContext(t)
	acStatus := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Instantiating}

	if err := initializeAppContextStatus(tc.ac, acStatus); err != nil {
		t.Fatalf("initializeAppContextStatus failed: %s", err)
	}
	got, err := getAppContextStatus(tc.ac)
	if err != nil {
		t.Fatalf("getAppContextStatus failed: %s", err)
	}
	if got.Status != appcontext.AppContextStatusEnum.Instantiating {
		t.Fatalf("Expected Instantiating, got %q", got.Status)
	}

	// Calling initialize again should update (not duplicate) the status value.
	acStatus.Status = appcontext.AppContextStatusEnum.Instantiated
	if err := initializeAppContextStatus(tc.ac, acStatus); err != nil {
		t.Fatalf("initializeAppContextStatus (update) failed: %s", err)
	}
	got, _ = getAppContextStatus(tc.ac)
	if got.Status != appcontext.AppContextStatusEnum.Instantiated {
		t.Fatalf("Expected Instantiated after update, got %q", got.Status)
	}
}

func TestUpdateEndingAppContextStatus(t *testing.T) {
	testCases := []struct {
		label    string
		start    appcontext.StatusValue
		failure  bool
		expected appcontext.StatusValue
		wantErr  bool
	}{
		{"Instantiating success -> Instantiated", appcontext.AppContextStatusEnum.Instantiating, false, appcontext.AppContextStatusEnum.Instantiated, false},
		{"Instantiating failure -> InstantiateFailed", appcontext.AppContextStatusEnum.Instantiating, true, appcontext.AppContextStatusEnum.InstantiateFailed, false},
		{"Terminating success -> Terminated", appcontext.AppContextStatusEnum.Terminating, false, appcontext.AppContextStatusEnum.Terminated, false},
		{"Terminating failure -> TerminateFailed", appcontext.AppContextStatusEnum.Terminating, true, appcontext.AppContextStatusEnum.TerminateFailed, false},
		{"Terminated is an invalid starting state", appcontext.AppContextStatusEnum.Terminated, false, "", true},
	}

	for _, tCase := range testCases {
		t.Run(tCase.label, func(t *testing.T) {
			tc := newMockContext(t)
			if err := initializeAppContextStatus(tc.ac, appcontext.AppContextStatus{Status: tCase.start}); err != nil {
				t.Fatalf("initializeAppContextStatus failed: %s", err)
			}
			err := updateEndingAppContextStatus(tc.ac, tc.rootHdl, tCase.failure)
			if tCase.wantErr {
				if err == nil {
					t.Fatal("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("updateEndingAppContextStatus returned an error: %s", err)
			}
			got, _ := getAppContextStatus(tc.ac)
			if got.Status != tCase.expected {
				t.Fatalf("Expected %q, got %q", tCase.expected, got.Status)
			}
		})
	}
}

func TestAppContextFlag(t *testing.T) {
	tc := newMockContext(t)

	t.Run("Defaults to false when the flag is unset", func(t *testing.T) {
		flag, _ := getAppContextFlag(tc.ac)
		if flag {
			t.Fatal("Expected the stop flag to default to false")
		}
	})

	t.Run("Reflects a value set via updateAppContextFlag", func(t *testing.T) {
		if err := updateAppContextFlag(tc.cid, true); err != nil {
			t.Fatalf("updateAppContextFlag(true) failed: %s", err)
		}
		flag, err := getAppContextFlag(tc.ac)
		if err != nil {
			t.Fatalf("getAppContextFlag failed: %s", err)
		}
		if !flag {
			t.Fatal("Expected the stop flag to be true")
		}

		// Update the existing flag value back to false.
		if err := updateAppContextFlag(tc.cid, false); err != nil {
			t.Fatalf("updateAppContextFlag(false) failed: %s", err)
		}
		flag, _ = getAppContextFlag(tc.ac)
		if flag {
			t.Fatal("Expected the stop flag to be false after update")
		}
	})
}

func TestInitializeResourceStatus(t *testing.T) {
	tc := newMockContext(t)
	appHdl, _ := tc.ac.AddApp(tc.rootHdl, "app1")
	cluster := "provider1+cluster1"
	tc.addResource(t, appHdl, cluster, "res1", "m1", resourcestatus.RsyncStatusEnum.Failed)

	// Wire up the app-order and resource-order instructions the function reads.
	appOrderJSON, _ := json.Marshal(appOrder{Apporder: []string{"app1"}})
	if _, err := tc.ac.AddInstruction(tc.rootHdl, "app", "order", string(appOrderJSON)); err != nil {
		t.Fatalf("AddInstruction(app order) failed: %s", err)
	}
	clusterHdl, _ := tc.ac.GetClusterHandle("app1", cluster)
	resOrderJSON, _ := json.Marshal(resOrder{Resorder: []string{"res1"}})
	if _, err := tc.ac.AddInstruction(clusterHdl, "resource", "order", string(resOrderJSON)); err != nil {
		t.Fatalf("AddInstruction(resource order) failed: %s", err)
	}

	acStatus := appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Instantiating}
	if err := initializeResourceStatus(tc.ac, acStatus); err != nil {
		t.Fatalf("initializeResourceStatus failed: %s", err)
	}

	// On Instantiating, an existing status is reset to Pending.
	if got := readResourceStatus(t, tc.ac, "app1", cluster, "res1"); got != resourcestatus.RsyncStatusEnum.Pending {
		t.Fatalf("Expected res1 status Pending, got %q", got)
	}
}

func TestChanLifecycle(t *testing.T) {
	instca := &CompositeAppContext{}

	c1 := addChan(instca)
	c2 := addChan(instca)
	if len(instca.chans) != 2 {
		t.Fatalf("Expected 2 channels, got %d", len(instca.chans))
	}

	if err := deleteChan(instca, c1); err != nil {
		t.Fatalf("deleteChan failed: %s", err)
	}
	if len(instca.chans) != 1 {
		t.Fatalf("Expected 1 channel after delete, got %d", len(instca.chans))
	}

	// Deleting the same channel again must report that it was not found.
	if err := deleteChan(instca, c1); err == nil {
		t.Fatal("Expected an error deleting an already-removed channel")
	}

	if err := deleteChan(instca, c2); err != nil {
		t.Fatalf("deleteChan(c2) failed: %s", err)
	}
	if len(instca.chans) != 0 {
		t.Fatalf("Expected 0 channels, got %d", len(instca.chans))
	}
}

// readResourceStatus fetches and decodes the RsyncStatus stored at a resource's
// status level.
func readResourceStatus(t *testing.T, ac appcontext.AppContext, app, cluster, res string) resourcestatus.RsyncStatus {
	t.Helper()
	rh, err := ac.GetResourceHandle(app, cluster, res)
	if err != nil {
		t.Fatalf("GetResourceHandle failed: %s", err)
	}
	sh, err := ac.GetLevelHandle(rh, "status")
	if err != nil {
		t.Fatalf("GetLevelHandle(status) failed: %s", err)
	}
	v, err := ac.GetValue(sh)
	if err != nil {
		t.Fatalf("GetValue failed: %s", err)
	}
	rStatus := resourcestatus.ResourceStatus{}
	js, _ := json.Marshal(v)
	json.Unmarshal(js, &rStatus)
	return rStatus.Status
}
