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

package installappserver

import (
	"context"
	"testing"

	installpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/installapp"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
)

// TestInstallApp_UnknownContextReportsFailure drives the real server handler
// with an app context id that does not exist in the in-memory context
// database. InstantiateComApp fails, and the handler must return a
// not-installed response together with the underlying error.
func TestInstallApp_UnknownContextReportsFailure(t *testing.T) {
	contextdb.Db = &contextdb.MockEtcd{}

	srv := NewInstallAppServer()
	resp, err := srv.InstallApp(context.Background(), &installpb.InstallAppRequest{
		AppContext: "nonexistent-appcontext",
	})
	if err == nil {
		t.Error("InstallApp() error = nil, want error for an unknown app context")
	}
	if resp == nil || resp.AppContextInstalled {
		t.Errorf("AppContextInstalled = %v, want false for an unknown app context", resp)
	}
}

// TestUninstallApp_UnknownContextReportsFailure exercises the terminate path
// through the same handler with a missing app context.
func TestUninstallApp_UnknownContextReportsFailure(t *testing.T) {
	contextdb.Db = &contextdb.MockEtcd{}

	srv := NewInstallAppServer()
	resp, err := srv.UninstallApp(context.Background(), &installpb.UninstallAppRequest{
		AppContext: "nonexistent-appcontext",
	})
	if err == nil {
		t.Error("UninstallApp() error = nil, want error for an unknown app context")
	}
	if resp == nil || resp.AppContextUninstalled {
		t.Errorf("AppContextUninstalled = %v, want false for an unknown app context", resp)
	}
}

func TestNewInstallAppServer(t *testing.T) {
	if NewInstallAppServer() == nil {
		t.Error("NewInstallAppServer() = nil, want non-nil server")
	}
}
