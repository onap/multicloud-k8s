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

package api

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	moduleLib "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/state"

	pkgerrors "github.com/pkg/errors"
)

// ----- GenericPlacementIntent mock -----

type mockGenericPlacementIntentManager struct {
	Items []moduleLib.GenericPlacementIntent
	Err   error
}

func (m *mockGenericPlacementIntentManager) CreateGenericPlacementIntent(g moduleLib.GenericPlacementIntent, p, ca, v, digName string) (moduleLib.GenericPlacementIntent, error) {
	if m.Err != nil {
		return moduleLib.GenericPlacementIntent{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockGenericPlacementIntentManager) GetGenericPlacementIntent(intentName, projectName, compositeAppName, version, digName string) (moduleLib.GenericPlacementIntent, error) {
	if m.Err != nil {
		return moduleLib.GenericPlacementIntent{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockGenericPlacementIntentManager) DeleteGenericPlacementIntent(intentName, projectName, compositeAppName, version, digName string) error {
	return m.Err
}

func (m *mockGenericPlacementIntentManager) GetAllGenericPlacementIntents(p, ca, v, digName string) ([]moduleLib.GenericPlacementIntent, error) {
	if m.Err != nil {
		return []moduleLib.GenericPlacementIntent{}, m.Err
	}
	return m.Items, nil
}

// ----- AppIntent mock -----

type mockAppIntentManager struct {
	Items []moduleLib.AppIntent
	Err   error
}

func (m *mockAppIntentManager) CreateAppIntent(a moduleLib.AppIntent, p, ca, v, i, digName string) (moduleLib.AppIntent, error) {
	if m.Err != nil {
		return moduleLib.AppIntent{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockAppIntentManager) GetAppIntent(ai, p, ca, v, i, digName string) (moduleLib.AppIntent, error) {
	if m.Err != nil {
		return moduleLib.AppIntent{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockAppIntentManager) GetAllIntentsByApp(aN, p, ca, v, i, digName string) (moduleLib.SpecData, error) {
	if m.Err != nil {
		return moduleLib.SpecData{}, m.Err
	}
	return m.Items[0].Spec, nil
}

func (m *mockAppIntentManager) GetAllAppIntents(p, ca, v, i, digName string) (moduleLib.ApplicationsAndClusterInfo, error) {
	if m.Err != nil {
		return moduleLib.ApplicationsAndClusterInfo{}, m.Err
	}
	return moduleLib.ApplicationsAndClusterInfo{}, nil
}

func (m *mockAppIntentManager) DeleteAppIntent(ai, p, ca, v, i, digName string) error {
	return m.Err
}

// ----- Intent (AddIntent) mock -----

type mockIntentManager struct {
	Items []moduleLib.Intent
	Err   error
}

func (m *mockIntentManager) AddIntent(a moduleLib.Intent, p, ca, v, di string) (moduleLib.Intent, error) {
	if m.Err != nil {
		return moduleLib.Intent{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockIntentManager) GetIntent(i, p, ca, v, di string) (moduleLib.Intent, error) {
	if m.Err != nil {
		return moduleLib.Intent{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockIntentManager) GetAllIntents(p, ca, v, di string) (moduleLib.ListOfIntents, error) {
	if m.Err != nil {
		return moduleLib.ListOfIntents{}, m.Err
	}
	return moduleLib.ListOfIntents{}, nil
}

func (m *mockIntentManager) GetIntentByName(i, p, ca, v, di string) (moduleLib.IntentSpecData, error) {
	if m.Err != nil {
		return moduleLib.IntentSpecData{}, m.Err
	}
	return m.Items[0].Spec, nil
}

func (m *mockIntentManager) DeleteIntent(i, p, ca, v, di string) error {
	return m.Err
}

// ----- DeploymentIntentGroup mock -----

type mockDeploymentIntentGroupManager struct {
	Items []moduleLib.DeploymentIntentGroup
	Err   error
}

func (m *mockDeploymentIntentGroupManager) CreateDeploymentIntentGroup(d moduleLib.DeploymentIntentGroup, p, ca, v string) (moduleLib.DeploymentIntentGroup, error) {
	if m.Err != nil {
		return moduleLib.DeploymentIntentGroup{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockDeploymentIntentGroupManager) GetDeploymentIntentGroup(di, p, ca, v string) (moduleLib.DeploymentIntentGroup, error) {
	if m.Err != nil {
		return moduleLib.DeploymentIntentGroup{}, m.Err
	}
	return m.Items[0], nil
}

func (m *mockDeploymentIntentGroupManager) GetDeploymentIntentGroupState(di, p, ca, v string) (state.StateInfo, error) {
	return state.StateInfo{}, m.Err
}

func (m *mockDeploymentIntentGroupManager) DeleteDeploymentIntentGroup(di, p, ca, v string) error {
	return m.Err
}

func (m *mockDeploymentIntentGroupManager) GetAllDeploymentIntentGroups(p, ca, v string) ([]moduleLib.DeploymentIntentGroup, error) {
	if m.Err != nil {
		return []moduleLib.DeploymentIntentGroup{}, m.Err
	}
	return m.Items, nil
}

func init() {
	gpiJSONFile = "../json-schemas/generic-placement-intent.json"
	appIntentJSONFile = "../json-schemas/generic-placement-intent-app.json"
	addIntentJSONFile = "../json-schemas/deployment-intent.json"
	dpiJSONFile = "../json-schemas/deployment-group-intent.json"
}

// ---------------- GenericPlacementIntent handler tests ----------------

func TestGenericPlacementIntentHandlers(t *testing.T) {
	base := "/v2/projects/testProject/composite-apps/testCompositeApp/v1/deployment-intent-groups/testDig/generic-placement-intents"
	gpi := []moduleLib.GenericPlacementIntent{
		{MetaData: moduleLib.GenIntentMetaData{Name: "testGpi"}},
	}

	t.Run("Create success", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte(`{"metadata":{"name":"testGpi"}}`))
		request := httptest.NewRequest("POST", base, reader)
		client := &mockGenericPlacementIntentManager{Items: gpi}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected %d; Got: %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("Create bad JSON", func(t *testing.T) {
		var reader io.Reader
		request := httptest.NewRequest("POST", base, reader)
		client := &mockGenericPlacementIntentManager{}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected %d; Got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Create manager error", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte(`{"metadata":{"name":"testGpi"}}`))
		request := httptest.NewRequest("POST", base, reader)
		client := &mockGenericPlacementIntentManager{Err: pkgerrors.New("Internal Error")}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("Get success", func(t *testing.T) {
		request := httptest.NewRequest("GET", base+"/testGpi", nil)
		client := &mockGenericPlacementIntentManager{Items: gpi}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("GetAll success", func(t *testing.T) {
		request := httptest.NewRequest("GET", base, nil)
		client := &mockGenericPlacementIntentManager{Items: gpi}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Delete success", func(t *testing.T) {
		request := httptest.NewRequest("DELETE", base+"/testGpi", nil)
		client := &mockGenericPlacementIntentManager{}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, client, nil, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected %d; Got: %d", http.StatusNoContent, resp.StatusCode)
		}
	})
}

// ---------------- AppIntent handler tests ----------------

func TestAppIntentHandlers(t *testing.T) {
	base := "/v2/projects/testProject/composite-apps/testCompositeApp/v1/deployment-intent-groups/testDig/generic-placement-intents/testGpi/app-intents"
	ai := []moduleLib.AppIntent{
		{MetaData: moduleLib.MetaData{Name: "testAppIntent"}, Spec: moduleLib.SpecData{AppName: "app1"}},
	}

	t.Run("Create success", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte(`{"metadata":{"name":"testAppIntent"},"spec":{"app-name":"app1"}}`))
		request := httptest.NewRequest("POST", base, reader)
		client := &mockAppIntentManager{Items: ai}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected %d; Got: %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("Create bad JSON", func(t *testing.T) {
		var reader io.Reader
		request := httptest.NewRequest("POST", base, reader)
		client := &mockAppIntentManager{}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected %d; Got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Create manager error", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte(`{"metadata":{"name":"testAppIntent"},"spec":{"app-name":"app1"}}`))
		request := httptest.NewRequest("POST", base, reader)
		client := &mockAppIntentManager{Err: pkgerrors.New("Internal Error")}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("Get success", func(t *testing.T) {
		request := httptest.NewRequest("GET", base+"/testAppIntent", nil)
		client := &mockAppIntentManager{Items: ai}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("GetAll success", func(t *testing.T) {
		request := httptest.NewRequest("GET", base, nil)
		client := &mockAppIntentManager{Items: ai}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Delete success", func(t *testing.T) {
		request := httptest.NewRequest("DELETE", base+"/testAppIntent", nil)
		client := &mockAppIntentManager{}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, client, nil, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected %d; Got: %d", http.StatusNoContent, resp.StatusCode)
		}
	})
}

// ---------------- AddIntent handler tests ----------------

func TestAddIntentHandlers(t *testing.T) {
	base := "/v2/projects/testProject/composite-apps/testCompositeApp/v1/deployment-intent-groups/testDig/intents"
	intents := []moduleLib.Intent{
		{
			MetaData: moduleLib.IntentMetaData{Name: "testIntent"},
			Spec:     moduleLib.IntentSpecData{Intent: map[string]string{"gpi": "gpiName"}},
		},
	}

	t.Run("Add success", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte(`{"metadata":{"name":"testIntent"},"spec":{"intent":{"gpi":"gpiName"}}}`))
		request := httptest.NewRequest("POST", base, reader)
		client := &mockIntentManager{Items: intents}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, client, nil, nil, nil))
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected %d; Got: %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("Add bad JSON", func(t *testing.T) {
		var reader io.Reader
		request := httptest.NewRequest("POST", base, reader)
		client := &mockIntentManager{}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, client, nil, nil, nil))
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected %d; Got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Add manager error", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte(`{"metadata":{"name":"testIntent"},"spec":{"intent":{"gpi":"gpiName"}}}`))
		request := httptest.NewRequest("POST", base, reader)
		client := &mockIntentManager{Err: pkgerrors.New("Internal Error")}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, client, nil, nil, nil))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("Get success", func(t *testing.T) {
		request := httptest.NewRequest("GET", base+"/testIntent", nil)
		client := &mockIntentManager{Items: intents}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, client, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("GetAll success", func(t *testing.T) {
		request := httptest.NewRequest("GET", base, nil)
		client := &mockIntentManager{Items: intents}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, client, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Delete success", func(t *testing.T) {
		request := httptest.NewRequest("DELETE", base+"/testIntent", nil)
		client := &mockIntentManager{}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, nil, client, nil, nil, nil))
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected %d; Got: %d", http.StatusNoContent, resp.StatusCode)
		}
	})
}

// ---------------- DeploymentIntentGroup handler tests ----------------

func TestDeploymentIntentGroupHandlers(t *testing.T) {
	base := "/v2/projects/testProject/composite-apps/testCompositeApp/v1/deployment-intent-groups"
	digs := []moduleLib.DeploymentIntentGroup{
		{
			MetaData: moduleLib.DepMetaData{Name: "testDig"},
			Spec:     moduleLib.DepSpecData{Profile: "prof", Version: "v1", LogicalCloud: "lc"},
		},
	}

	t.Run("Create success", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte(`{"metadata":{"name":"testDig"},"spec":{"profile":"prof","version":"v1","logical-cloud":"lc","override-values":[{"app-name":"app1","values":{"k":"v"}}]}}`))
		request := httptest.NewRequest("POST", base, reader)
		client := &mockDeploymentIntentGroupManager{Items: digs}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("Expected %d; Got: %d", http.StatusCreated, resp.StatusCode)
		}
	})

	t.Run("Create bad JSON", func(t *testing.T) {
		var reader io.Reader
		request := httptest.NewRequest("POST", base, reader)
		client := &mockDeploymentIntentGroupManager{}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected %d; Got: %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Create manager error", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte(`{"metadata":{"name":"testDig"},"spec":{"profile":"prof","version":"v1","logical-cloud":"lc","override-values":[{"app-name":"app1","values":{"k":"v"}}]}}`))
		request := httptest.NewRequest("POST", base, reader)
		client := &mockDeploymentIntentGroupManager{Err: pkgerrors.New("Internal Error")}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("Expected %d; Got: %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})

	t.Run("Get success", func(t *testing.T) {
		request := httptest.NewRequest("GET", base+"/testDig", nil)
		client := &mockDeploymentIntentGroupManager{Items: digs}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("GetAll success", func(t *testing.T) {
		request := httptest.NewRequest("GET", base, nil)
		client := &mockDeploymentIntentGroupManager{Items: digs}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected %d; Got: %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("Delete success", func(t *testing.T) {
		request := httptest.NewRequest("DELETE", base+"/testDig", nil)
		client := &mockDeploymentIntentGroupManager{}
		resp := executeRequest(request, NewRouter(nil, nil, nil, nil, nil, nil, client, nil, nil, nil, nil))
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected %d; Got: %d", http.StatusNoContent, resp.StatusCode)
		}
	})
}
