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

package contextupdateserver

import (
	"context"
	"testing"

	contextpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/contextupdate"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
)

// TestUpdateAppContext_UnknownContextReportsFailure drives the real server
// handler with an app context id that does not exist in the (in-memory)
// context database. action.UpdateAppContext fails to load it, and the server
// must translate that error into a non-nil-message, not-updated response
// (without returning a transport-level error).
func TestUpdateAppContext_UnknownContextReportsFailure(t *testing.T) {
	contextdb.Db = &contextdb.MockEtcd{}

	srv := NewContextupdateServer()
	resp, err := srv.UpdateAppContext(context.Background(), &contextpb.ContextUpdateRequest{
		AppContext: "nonexistent-appcontext",
		IntentName: "intent-1",
	})
	if err != nil {
		t.Fatalf("UpdateAppContext returned transport error: %v (expected in-band failure)", err)
	}
	if resp.AppContextUpdated {
		t.Error("AppContextUpdated = true, want false for an unknown app context")
	}
	if resp.AppContextUpdateMessage == "" {
		t.Error("expected a non-empty failure message")
	}
}

func TestNewContextupdateServer(t *testing.T) {
	if NewContextupdateServer() == nil {
		t.Error("NewContextupdateServer() = nil, want non-nil server")
	}
}
