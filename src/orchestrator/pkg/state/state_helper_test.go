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

package state

import (
	"sort"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
)

func TestGetCurrentStateFromStateInfo(t *testing.T) {
	testCases := []struct {
		label    string
		stateInfo StateInfo
		expected StateValue
		wantErr  bool
	}{
		{
			label:    "Empty actions returns an error",
			stateInfo: StateInfo{},
			expected: StateEnum.Undefined,
			wantErr:  true,
		},
		{
			label: "Returns the last action's state",
			stateInfo: StateInfo{Actions: []ActionEntry{
				{State: StateEnum.Created},
				{State: StateEnum.Instantiated},
			}},
			expected: StateEnum.Instantiated,
			wantErr:  false,
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.label, func(t *testing.T) {
			got, err := GetCurrentStateFromStateInfo(tCase.stateInfo)
			if tCase.wantErr {
				if err == nil {
					t.Fatal("Expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %s", err)
			}
			if got != tCase.expected {
				t.Fatalf("Expected %q, got %q", tCase.expected, got)
			}
		})
	}
}

func TestGetLastContextIdFromStateInfo(t *testing.T) {
	t.Run("Returns empty string when there are no actions", func(t *testing.T) {
		if got := GetLastContextIdFromStateInfo(StateInfo{}); got != "" {
			t.Fatalf("Expected empty string, got %q", got)
		}
	})

	t.Run("Returns the most recent context id", func(t *testing.T) {
		s := StateInfo{Actions: []ActionEntry{
			{State: StateEnum.Created, ContextId: "ctx1"},
			{State: StateEnum.Instantiated, ContextId: "ctx2"},
		}}
		if got := GetLastContextIdFromStateInfo(s); got != "ctx2" {
			t.Fatalf("Expected 'ctx2', got %q", got)
		}
	})
}

func TestGetContextIdsFromStateInfo(t *testing.T) {
	t.Run("Returns unique non-empty context ids", func(t *testing.T) {
		s := StateInfo{Actions: []ActionEntry{
			{State: StateEnum.Created, ContextId: ""},
			{State: StateEnum.Instantiated, ContextId: "ctx1"},
			{State: StateEnum.Terminated, ContextId: "ctx1"},
			{State: StateEnum.Instantiated, ContextId: "ctx2"},
		}}
		got := GetContextIdsFromStateInfo(s)
		sort.Strings(got)
		if len(got) != 2 || got[0] != "ctx1" || got[1] != "ctx2" {
			t.Fatalf("Expected [ctx1 ctx2], got %v", got)
		}
	})

	t.Run("Returns an empty slice when there are no context ids", func(t *testing.T) {
		s := StateInfo{Actions: []ActionEntry{{State: StateEnum.Created, ContextId: ""}}}
		if got := GetContextIdsFromStateInfo(s); len(got) != 0 {
			t.Fatalf("Expected no context ids, got %v", got)
		}
	})
}

func TestGetAppContextFromId(t *testing.T) {
	t.Run("Loads an existing appcontext", func(t *testing.T) {
		contextdb.Db = &contextdb.MockEtcd{}
		ac := appcontext.AppContext{}
		cid, err := ac.InitAppContext()
		if err != nil {
			t.Fatalf("InitAppContext failed: %s", err)
		}
		if _, err := ac.CreateCompositeApp(); err != nil {
			t.Fatalf("CreateCompositeApp failed: %s", err)
		}

		loaded, err := GetAppContextFromId(cid.(string))
		if err != nil {
			t.Fatalf("GetAppContextFromId returned an error: %s", err)
		}
		if _, err := loaded.GetCompositeAppHandle(); err != nil {
			t.Fatalf("Loaded appcontext is not usable: %s", err)
		}
	})

	t.Run("Returns an error for an unknown id", func(t *testing.T) {
		contextdb.Db = &contextdb.MockEtcd{}
		if _, err := GetAppContextFromId("does-not-exist"); err == nil {
			t.Fatal("Expected an error for an unknown context id, got nil")
		}
	})
}

func TestGetAppContextStatus(t *testing.T) {
	contextdb.Db = &contextdb.MockEtcd{}
	ac := appcontext.AppContext{}
	cid, err := ac.InitAppContext()
	if err != nil {
		t.Fatalf("InitAppContext failed: %s", err)
	}
	h, err := ac.CreateCompositeApp()
	if err != nil {
		t.Fatalf("CreateCompositeApp failed: %s", err)
	}
	if _, err := ac.AddLevelValue(h, "status", appcontext.AppContextStatus{Status: appcontext.AppContextStatusEnum.Instantiated}); err != nil {
		t.Fatalf("AddLevelValue(status) failed: %s", err)
	}

	got, err := GetAppContextStatus(cid.(string))
	if err != nil {
		t.Fatalf("GetAppContextStatus returned an error: %s", err)
	}
	if got.Status != appcontext.AppContextStatusEnum.Instantiated {
		t.Fatalf("Expected Instantiated, got %q", got.Status)
	}
}
