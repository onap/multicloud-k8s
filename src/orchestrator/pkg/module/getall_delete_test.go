/*
 * Copyright 2020 Intel Corporation, Inc
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

package module

import (
	"fmt"
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/state"

	pkgerrors "github.com/pkg/errors"
)

const (
	gdProject      = "testProject"
	gdCompositeApp = "testCompositeApp"
	gdVersion      = "v1"
	gdDig          = "testDig"
	gdIntent       = "testIntent"
)

// gdProjectItem returns a seeded project record.
func gdProjectItem(items map[string]map[string][]byte) {
	items[ProjectKey{ProjectName: gdProject}.String()] = map[string][]byte{
		"projectmetadata": []byte("{\"metadata\":{\"Name\":\"testProject\"}}"),
	}
}

// gdCompositeAppItem returns a seeded composite app record.
func gdCompositeAppItem(items map[string]map[string][]byte) {
	items[CompositeAppKey{CompositeAppName: gdCompositeApp, Version: gdVersion, Project: gdProject}.String()] = map[string][]byte{
		"compositeappmetadata": []byte("{\"metadata\":{\"Name\":\"testCompositeApp\"},\"spec\":{\"Version\":\"v1\"}}"),
	}
}

func TestGetApps(t *testing.T) {
	items := map[string]map[string][]byte{}
	items[fmt.Sprintf("%v", AppKey{App: "", Project: gdProject, CompositeApp: gdCompositeApp, CompositeAppVersion: gdVersion})] = map[string][]byte{
		"appmetadata": []byte("{\"metadata\":{\"name\":\"app1\"}}"),
	}

	t.Run("Get Apps", func(t *testing.T) {
		db.DBconn = &db.MockDB{Items: items}
		impl := NewAppClient()
		got, err := impl.GetApps(gdProject, gdCompositeApp, gdVersion)
		if err != nil {
			t.Fatalf("GetApps returned an unexpected error %s", err)
		}
		if len(got) != 1 || got[0].Metadata.Name != "app1" {
			t.Fatalf("GetApps returned unexpected body: %v", got)
		}
	})

	t.Run("Get Apps Error", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewAppClient()
		_, err := impl.GetApps(gdProject, gdCompositeApp, gdVersion)
		if err == nil || !strings.Contains(err.Error(), "DB Error") {
			t.Fatalf("GetApps expected DB Error, got %v", err)
		}
	})
}

func TestGetAllCompositeApps(t *testing.T) {
	t.Run("Get All Composite Apps", func(t *testing.T) {
		items := map[string]map[string][]byte{}
		gdProjectItem(items)
		items[CompositeAppKey{CompositeAppName: "", Version: "", Project: gdProject}.String()] = map[string][]byte{
			"compositeappmetadata": []byte("{\"metadata\":{\"name\":\"testCompositeApp\"},\"spec\":{\"version\":\"v1\"}}"),
		}
		db.DBconn = &db.MockDB{Items: items}
		impl := NewCompositeAppClient()
		got, err := impl.GetAllCompositeApps(gdProject)
		if err != nil {
			t.Fatalf("GetAllCompositeApps returned an unexpected error %s", err)
		}
		if len(got) != 1 || got[0].Metadata.Name != "testCompositeApp" {
			t.Fatalf("GetAllCompositeApps returned unexpected body: %v", got)
		}
	})

	t.Run("Get All Composite Apps Missing Project", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewCompositeAppClient()
		_, err := impl.GetAllCompositeApps(gdProject)
		if err == nil || !strings.Contains(err.Error(), "Unable to find the project") {
			t.Fatalf("GetAllCompositeApps expected project error, got %v", err)
		}
	})
}

func TestGetAllGenericPlacementIntents(t *testing.T) {
	t.Run("Get All Generic Placement Intents", func(t *testing.T) {
		items := map[string]map[string][]byte{}
		gdProjectItem(items)
		gdCompositeAppItem(items)
		items[GenericPlacementIntentKey{Name: "", Project: gdProject, CompositeApp: gdCompositeApp, Version: gdVersion, DigName: gdDig}.String()] = map[string][]byte{
			"genericplacementintentmetadata": []byte("{\"metadata\":{\"name\":\"gpi1\"}}"),
		}
		db.DBconn = &db.MockDB{Items: items}
		impl := NewGenericPlacementIntentClient()
		got, err := impl.GetAllGenericPlacementIntents(gdProject, gdCompositeApp, gdVersion, gdDig)
		if err != nil {
			t.Fatalf("GetAllGenericPlacementIntents returned an unexpected error %s", err)
		}
		if len(got) != 1 || got[0].MetaData.Name != "gpi1" {
			t.Fatalf("GetAllGenericPlacementIntents returned unexpected body: %v", got)
		}
	})

	t.Run("Get All Generic Placement Intents Missing Project", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewGenericPlacementIntentClient()
		_, err := impl.GetAllGenericPlacementIntents(gdProject, gdCompositeApp, gdVersion, gdDig)
		if err == nil || !strings.Contains(err.Error(), "Unable to find the project") {
			t.Fatalf("GetAllGenericPlacementIntents expected project error, got %v", err)
		}
	})
}

func TestDeleteGenericPlacementIntent(t *testing.T) {
	t.Run("Delete Generic Placement Intent", func(t *testing.T) {
		db.DBconn = &db.MockDB{}
		impl := NewGenericPlacementIntentClient()
		if err := impl.DeleteGenericPlacementIntent(gdIntent, gdProject, gdCompositeApp, gdVersion, gdDig); err != nil {
			t.Fatalf("DeleteGenericPlacementIntent returned an unexpected error %s", err)
		}
	})

	t.Run("Delete Generic Placement Intent Error", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewGenericPlacementIntentClient()
		err := impl.DeleteGenericPlacementIntent(gdIntent, gdProject, gdCompositeApp, gdVersion, gdDig)
		if err == nil || !strings.Contains(err.Error(), "DB Error") {
			t.Fatalf("DeleteGenericPlacementIntent expected DB Error, got %v", err)
		}
	})
}

func TestGetAllDeploymentIntentGroups(t *testing.T) {
	t.Run("Get All Deployment Intent Groups", func(t *testing.T) {
		items := map[string]map[string][]byte{}
		gdProjectItem(items)
		gdCompositeAppItem(items)
		items[DeploymentIntentGroupKey{Name: "", Project: gdProject, CompositeApp: gdCompositeApp, Version: gdVersion}.String()] = map[string][]byte{
			"deploymentintentgroupmetadata": []byte("{\"metadata\":{\"name\":\"dig1\"},\"spec\":{\"profile\":\"prof\",\"version\":\"v1\"}}"),
		}
		db.DBconn = &db.MockDB{Items: items}
		impl := NewDeploymentIntentGroupClient()
		got, err := impl.GetAllDeploymentIntentGroups(gdProject, gdCompositeApp, gdVersion)
		if err != nil {
			t.Fatalf("GetAllDeploymentIntentGroups returned an unexpected error %s", err)
		}
		if len(got) != 1 || got[0].MetaData.Name != "dig1" {
			t.Fatalf("GetAllDeploymentIntentGroups returned unexpected body: %v", got)
		}
	})

	t.Run("Get All Deployment Intent Groups Missing Project", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewDeploymentIntentGroupClient()
		_, err := impl.GetAllDeploymentIntentGroups(gdProject, gdCompositeApp, gdVersion)
		if err == nil || !strings.Contains(err.Error(), "Unable to find the project") {
			t.Fatalf("GetAllDeploymentIntentGroups expected project error, got %v", err)
		}
	})
}

func TestGetDeploymentIntentGroupState(t *testing.T) {
	t.Run("Get Deployment Intent Group State", func(t *testing.T) {
		items := map[string]map[string][]byte{}
		items[DeploymentIntentGroupKey{Name: gdDig, Project: gdProject, CompositeApp: gdCompositeApp, Version: gdVersion}.String()] = map[string][]byte{
			"stateInfo": []byte("{\"actions\":[{\"state\":\"Created\",\"instance\":\"\",\"time\":\"2020-01-01T00:00:00Z\"}]}"),
		}
		db.DBconn = &db.MockDB{Items: items}
		impl := NewDeploymentIntentGroupClient()
		got, err := impl.GetDeploymentIntentGroupState(gdDig, gdProject, gdCompositeApp, gdVersion)
		if err != nil {
			t.Fatalf("GetDeploymentIntentGroupState returned an unexpected error %s", err)
		}
		if len(got.Actions) != 1 || got.Actions[0].State != state.StateEnum.Created {
			t.Fatalf("GetDeploymentIntentGroupState returned unexpected body: %v", got)
		}
	})

	t.Run("Get Deployment Intent Group State Error", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewDeploymentIntentGroupClient()
		_, err := impl.GetDeploymentIntentGroupState(gdDig, gdProject, gdCompositeApp, gdVersion)
		if err == nil || !strings.Contains(err.Error(), "DB Error") {
			t.Fatalf("GetDeploymentIntentGroupState expected DB Error, got %v", err)
		}
	})
}

func TestDeleteDeploymentIntentGroup(t *testing.T) {
	t.Run("Delete Deployment Intent Group in Created state", func(t *testing.T) {
		items := map[string]map[string][]byte{}
		items[DeploymentIntentGroupKey{Name: gdDig, Project: gdProject, CompositeApp: gdCompositeApp, Version: gdVersion}.String()] = map[string][]byte{
			"stateInfo": []byte("{\"actions\":[{\"state\":\"Created\",\"instance\":\"\",\"time\":\"2020-01-01T00:00:00Z\"}]}"),
		}
		db.DBconn = &db.MockDB{Items: items}
		impl := NewDeploymentIntentGroupClient()
		if err := impl.DeleteDeploymentIntentGroup(gdDig, gdProject, gdCompositeApp, gdVersion); err != nil {
			t.Fatalf("DeleteDeploymentIntentGroup returned an unexpected error %s", err)
		}
	})

	t.Run("Delete Deployment Intent Group missing state", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewDeploymentIntentGroupClient()
		err := impl.DeleteDeploymentIntentGroup(gdDig, gdProject, gdCompositeApp, gdVersion)
		if err == nil || !strings.Contains(err.Error(), "Error getting stateInfo") {
			t.Fatalf("DeleteDeploymentIntentGroup expected stateInfo error, got %v", err)
		}
	})
}

func TestGetAllIntentsByApp(t *testing.T) {
	t.Run("Get All Intents By App", func(t *testing.T) {
		items := map[string]map[string][]byte{}
		items[fmt.Sprintf("%v", AppIntentFindByAppKey{Project: gdProject, CompositeApp: gdCompositeApp, CompositeAppVersion: gdVersion, Intent: gdIntent, DeploymentIntentGroupName: gdDig, AppName: "app1"})] = map[string][]byte{
			"appintentmetadata": []byte("{\"metadata\":{\"name\":\"ai1\"},\"spec\":{\"app-name\":\"app1\"}}"),
		}
		db.DBconn = &db.MockDB{Items: items}
		impl := NewAppIntentClient()
		got, err := impl.GetAllIntentsByApp("app1", gdProject, gdCompositeApp, gdVersion, gdIntent, gdDig)
		if err != nil {
			t.Fatalf("GetAllIntentsByApp returned an unexpected error %s", err)
		}
		if got.AppName != "app1" {
			t.Fatalf("GetAllIntentsByApp returned unexpected body: %v", got)
		}
	})

	t.Run("Get All Intents By App Error", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewAppIntentClient()
		_, err := impl.GetAllIntentsByApp("app1", gdProject, gdCompositeApp, gdVersion, gdIntent, gdDig)
		if err == nil || !strings.Contains(err.Error(), "DB Error") {
			t.Fatalf("GetAllIntentsByApp expected DB Error, got %v", err)
		}
	})
}

func TestGetAllAppIntents(t *testing.T) {
	t.Run("Get All App Intents", func(t *testing.T) {
		items := map[string]map[string][]byte{}
		items[AppIntentKey{Name: "", Project: gdProject, CompositeApp: gdCompositeApp, Version: gdVersion, Intent: gdIntent, DeploymentIntentGroupName: gdDig}.String()] = map[string][]byte{
			"appintentmetadata": []byte("{\"metadata\":{\"name\":\"ai1\"},\"spec\":{\"app-name\":\"app1\"}}"),
		}
		db.DBconn = &db.MockDB{Items: items}
		impl := NewAppIntentClient()
		got, err := impl.GetAllAppIntents(gdProject, gdCompositeApp, gdVersion, gdIntent, gdDig)
		if err != nil {
			t.Fatalf("GetAllAppIntents returned an unexpected error %s", err)
		}
		if len(got.ArrayOfAppClusterInfo) != 1 || got.ArrayOfAppClusterInfo[0].Name != "app1" {
			t.Fatalf("GetAllAppIntents returned unexpected body: %v", got)
		}
	})

	t.Run("Get All App Intents Error", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewAppIntentClient()
		_, err := impl.GetAllAppIntents(gdProject, gdCompositeApp, gdVersion, gdIntent, gdDig)
		if err == nil || !strings.Contains(err.Error(), "DB Error") {
			t.Fatalf("GetAllAppIntents expected DB Error, got %v", err)
		}
	})
}

func TestDeleteAppIntent(t *testing.T) {
	t.Run("Delete App Intent", func(t *testing.T) {
		db.DBconn = &db.MockDB{}
		impl := NewAppIntentClient()
		if err := impl.DeleteAppIntent("ai1", gdProject, gdCompositeApp, gdVersion, gdIntent, gdDig); err != nil {
			t.Fatalf("DeleteAppIntent returned an unexpected error %s", err)
		}
	})

	t.Run("Delete App Intent Error", func(t *testing.T) {
		db.DBconn = &db.MockDB{Err: pkgerrors.New("DB Error")}
		impl := NewAppIntentClient()
		err := impl.DeleteAppIntent("ai1", gdProject, gdCompositeApp, gdVersion, gdIntent, gdDig)
		if err == nil || !strings.Contains(err.Error(), "DB Error") {
			t.Fatalf("DeleteAppIntent expected DB Error, got %v", err)
		}
	})
}
