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

package appcontext

import (
	"sort"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
)

// These tests drive a real AppContext through the real RunTimeContext against
// the in-memory contextdb double. They exercise the full
// appcontext -> rtcontext -> contextdb seam (the orchestrator<->rsync
// integration point) end to end, rather than mocking the rtcontext layer.

// newAppContext returns a fresh AppContext backed by an in-memory contextdb,
// already initialized with a composite app, and its root handle.
func newAppContext(t *testing.T) (AppContext, interface{}) {
	t.Helper()
	contextdb.Db = &contextdb.MockEtcd{}
	ac := AppContext{}
	if _, err := ac.InitAppContext(); err != nil {
		t.Fatalf("InitAppContext failed: %s", err)
	}
	h, err := ac.CreateCompositeApp()
	if err != nil {
		t.Fatalf("CreateCompositeApp failed: %s", err)
	}
	return ac, h
}

func TestSeamAppLifecycle(t *testing.T) {
	ac, root := newAppContext(t)

	appHdl, err := ac.AddApp(root, "app1")
	if err != nil {
		t.Fatalf("AddApp failed: %s", err)
	}

	got, err := ac.GetAppHandle("app1")
	if err != nil {
		t.Fatalf("GetAppHandle failed: %s", err)
	}
	if got != appHdl {
		t.Fatalf("Expected app handle %v, got %v", appHdl, got)
	}

	if _, err := ac.GetAppHandle("nonexistent"); err == nil {
		t.Fatal("Expected an error for a nonexistent app handle")
	}

	if err := ac.DeleteApp(appHdl); err != nil {
		t.Fatalf("DeleteApp failed: %s", err)
	}
	if _, err := ac.GetAppHandle("app1"); err == nil {
		t.Fatal("Expected an error fetching a deleted app handle")
	}
}

func TestSeamClusterLifecycle(t *testing.T) {
	ac, root := newAppContext(t)
	appHdl, _ := ac.AddApp(root, "app1")

	c1, err := ac.AddCluster(appHdl, "provider1+clusterA")
	if err != nil {
		t.Fatalf("AddCluster failed: %s", err)
	}
	if _, err := ac.AddCluster(appHdl, "provider1+clusterB"); err != nil {
		t.Fatalf("AddCluster failed: %s", err)
	}

	// GetClusterHandle should find a known cluster and reject an unknown one.
	got, err := ac.GetClusterHandle("app1", "provider1+clusterA")
	if err != nil {
		t.Fatalf("GetClusterHandle failed: %s", err)
	}
	if got != c1 {
		t.Fatalf("Expected cluster handle %v, got %v", c1, got)
	}
	if _, err := ac.GetClusterHandle("app1", "provider1+clusterZ"); err == nil {
		t.Fatal("Expected an error for an unknown cluster")
	}

	// GetClusterNames should list every cluster added under the app.
	names, err := ac.GetClusterNames("app1")
	if err != nil {
		t.Fatalf("GetClusterNames failed: %s", err)
	}
	sort.Strings(names)
	if len(names) != 2 || names[0] != "provider1+clusterA" || names[1] != "provider1+clusterB" {
		t.Fatalf("Expected [provider1+clusterA provider1+clusterB], got %v", names)
	}

	// Invalid arguments should be rejected.
	if _, err := ac.GetClusterHandle("", "c"); err == nil {
		t.Fatal("Expected an error for an empty app name")
	}
	if _, err := ac.GetClusterNames(""); err == nil {
		t.Fatal("Expected an error for an empty app name")
	}
}

func TestSeamClusterMetaGroup(t *testing.T) {
	ac, root := newAppContext(t)
	appHdl, _ := ac.AddApp(root, "app1")
	chA, _ := ac.AddCluster(appHdl, "provider1+clusterA")
	chB, _ := ac.AddCluster(appHdl, "provider1+clusterB")

	if err := ac.AddClusterMetaGrp(chA, "1"); err != nil {
		t.Fatalf("AddClusterMetaGrp failed: %s", err)
	}
	if err := ac.AddClusterMetaGrp(chB, "1"); err != nil {
		t.Fatalf("AddClusterMetaGrp failed: %s", err)
	}

	// GetClusterMetaHandle builds a deterministic handle string.
	mh, err := ac.GetClusterMetaHandle("app1", "provider1+clusterA")
	if err != nil {
		t.Fatalf("GetClusterMetaHandle failed: %s", err)
	}
	if mh == "" {
		t.Fatal("Expected a non-empty cluster meta handle")
	}

	// GetClusterGroupMap should group both clusters under group "1".
	gmap, err := ac.GetClusterGroupMap("app1")
	if err != nil {
		t.Fatalf("GetClusterGroupMap failed: %s", err)
	}
	if len(gmap["1"]) != 2 {
		t.Fatalf("Expected 2 clusters in group 1, got %v", gmap["1"])
	}

	if err := ac.DeleteClusterMetaGrpHandle(chA); err != nil {
		t.Fatalf("DeleteClusterMetaGrpHandle failed: %s", err)
	}
}

func TestSeamResourceLifecycle(t *testing.T) {
	ac, root := newAppContext(t)
	appHdl, _ := ac.AddApp(root, "app1")
	clusterHdl, _ := ac.AddCluster(appHdl, "provider1+clusterA")

	resHdl, err := ac.AddResource(clusterHdl, "res1", "manifest")
	if err != nil {
		t.Fatalf("AddResource failed: %s", err)
	}

	got, err := ac.GetResourceHandle("app1", "provider1+clusterA", "res1")
	if err != nil {
		t.Fatalf("GetResourceHandle failed: %s", err)
	}
	if got != resHdl {
		t.Fatalf("Expected resource handle %v, got %v", resHdl, got)
	}

	// Update and read back the resource value.
	if err := ac.UpdateResourceValue(resHdl, "manifest-v2"); err != nil {
		t.Fatalf("UpdateResourceValue failed: %s", err)
	}
	v, err := ac.GetValue(resHdl)
	if err != nil {
		t.Fatalf("GetValue failed: %s", err)
	}
	if v != "manifest-v2" {
		t.Fatalf("Expected 'manifest-v2', got %v", v)
	}

	// A status subhandle can be added and then located via GetResourceStatusHandle.
	if _, err := ac.AddLevelValue(resHdl, "status", "Pending"); err != nil {
		t.Fatalf("AddLevelValue(status) failed: %s", err)
	}
	if _, err := ac.GetResourceStatusHandle("app1", "provider1+clusterA", "res1"); err != nil {
		t.Fatalf("GetResourceStatusHandle failed: %s", err)
	}

	if _, err := ac.GetResourceHandle("app1", "provider1+clusterA", "missing"); err == nil {
		t.Fatal("Expected an error for a missing resource")
	}
}

func TestSeamInstructions(t *testing.T) {
	ac, root := newAppContext(t)
	appHdl, _ := ac.AddApp(root, "app1")
	clusterHdl, _ := ac.AddCluster(appHdl, "provider1+clusterA")

	t.Run("App instruction round-trips", func(t *testing.T) {
		if _, err := ac.AddInstruction(root, "app", "order", "appvalue"); err != nil {
			t.Fatalf("AddInstruction failed: %s", err)
		}
		v, err := ac.GetAppInstruction("order")
		if err != nil {
			t.Fatalf("GetAppInstruction failed: %s", err)
		}
		if v != "appvalue" {
			t.Fatalf("Expected 'appvalue', got %v", v)
		}
	})

	t.Run("Resource instruction round-trips", func(t *testing.T) {
		if _, err := ac.AddInstruction(clusterHdl, "resource", "order", "resvalue"); err != nil {
			t.Fatalf("AddInstruction failed: %s", err)
		}
		v, err := ac.GetResourceInstruction("app1", "provider1+clusterA", "order")
		if err != nil {
			t.Fatalf("GetResourceInstruction failed: %s", err)
		}
		if v != "resvalue" {
			t.Fatalf("Expected 'resvalue', got %v", v)
		}
	})

	t.Run("Invalid instruction type is rejected", func(t *testing.T) {
		if _, err := ac.AddInstruction(root, "app", "bogus", "v"); err == nil {
			t.Fatal("Expected an error for an invalid instruction type")
		}
		if _, err := ac.GetAppInstruction("bogus"); err == nil {
			t.Fatal("Expected an error for an invalid instruction type")
		}
	})

	t.Run("Invalid instruction level is rejected", func(t *testing.T) {
		if _, err := ac.AddInstruction(root, "bogus", "order", "v"); err == nil {
			t.Fatal("Expected an error for an invalid instruction level")
		}
	})
}

func TestSeamStatusValues(t *testing.T) {
	ac, root := newAppContext(t)

	sh, err := ac.AddLevelValue(root, "status", AppContextStatus{Status: AppContextStatusEnum.Instantiating})
	if err != nil {
		t.Fatalf("AddLevelValue(status) failed: %s", err)
	}

	// GetLevelHandle should locate the status level we just created.
	lh, err := ac.GetLevelHandle(root, "status")
	if err != nil {
		t.Fatalf("GetLevelHandle failed: %s", err)
	}
	if lh != sh {
		t.Fatalf("Expected level handle %v, got %v", sh, lh)
	}

	if err := ac.UpdateStatusValue(sh, AppContextStatus{Status: AppContextStatusEnum.Instantiated}); err != nil {
		t.Fatalf("UpdateStatusValue failed: %s", err)
	}

	if _, err := ac.GetLevelHandle(root, "nolevel"); err == nil {
		t.Fatal("Expected an error for a nonexistent level")
	}
}

func TestSeamCompositeAppMeta(t *testing.T) {
	contextdb.Db = &contextdb.MockEtcd{}
	ac := AppContext{}
	if _, err := ac.InitAppContext(); err != nil {
		t.Fatalf("InitAppContext failed: %s", err)
	}
	if _, err := ac.CreateCompositeApp(); err != nil {
		t.Fatalf("CreateCompositeApp failed: %s", err)
	}

	meta := CompositeAppMeta{
		Project:               "proj1",
		CompositeApp:          "ca1",
		Version:               "v1",
		Release:               "r1",
		DeploymentIntentGroup: "dig1",
	}
	if err := ac.AddCompositeAppMeta(meta); err != nil {
		t.Fatalf("AddCompositeAppMeta failed: %s", err)
	}

	got, err := ac.GetCompositeAppMeta()
	if err != nil {
		t.Fatalf("GetCompositeAppMeta failed: %s", err)
	}
	if got.Project != "proj1" || got.CompositeApp != "ca1" || got.Version != "v1" ||
		got.Release != "r1" || got.DeploymentIntentGroup != "dig1" {
		t.Fatalf("GetCompositeAppMeta returned unexpected value: %+v", got)
	}
}

func TestSeamLoadAppContext(t *testing.T) {
	contextdb.Db = &contextdb.MockEtcd{}
	ac := AppContext{}
	cid, err := ac.InitAppContext()
	if err != nil {
		t.Fatalf("InitAppContext failed: %s", err)
	}
	if _, err := ac.CreateCompositeApp(); err != nil {
		t.Fatalf("CreateCompositeApp failed: %s", err)
	}

	// A second AppContext should be able to load the same context by id.
	ac2 := AppContext{}
	if _, err := ac2.LoadAppContext(cid); err != nil {
		t.Fatalf("LoadAppContext failed: %s", err)
	}
	if _, err := ac2.GetCompositeAppHandle(); err != nil {
		t.Fatalf("GetCompositeAppHandle on loaded context failed: %s", err)
	}
}

func TestSeamGetAllHandles(t *testing.T) {
	ac, root := newAppContext(t)
	appHdl, _ := ac.AddApp(root, "app1")
	if _, err := ac.AddCluster(appHdl, "provider1+clusterA"); err != nil {
		t.Fatalf("AddCluster failed: %s", err)
	}

	handles, err := ac.GetAllHandles(root)
	if err != nil {
		t.Fatalf("GetAllHandles failed: %s", err)
	}
	// At minimum: the composite-app root, the app level, and the cluster level.
	if len(handles) < 3 {
		t.Fatalf("Expected at least 3 handles, got %d: %v", len(handles), handles)
	}
}
