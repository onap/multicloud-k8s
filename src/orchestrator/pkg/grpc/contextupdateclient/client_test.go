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

package contextupdateclient

import (
	"context"
	"net"
	"strconv"
	"testing"

	contextpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/contextupdate"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	"google.golang.org/grpc"
)

// stubContextupdateServer is a controllable implementation of the generated
// ContextupdateServer used to drive InvokeContextUpdate through a real gRPC
// transport on a loopback ephemeral port.
type stubContextupdateServer struct {
	contextpb.UnimplementedContextupdateServer
	updated bool
	message string
}

func (s *stubContextupdateServer) UpdateAppContext(ctx context.Context, req *contextpb.ContextUpdateRequest) (*contextpb.ContextUpdateResponse, error) {
	return &contextpb.ContextUpdateResponse{
		AppContextUpdated:       s.updated,
		AppContextUpdateMessage: s.message,
	}, nil
}

// startStubController registers the stub server under the given controller name
// in the rpc connection map and returns a cleanup func.
func startStubController(t *testing.T, name string, srv contextpb.ContextupdateServer) func() {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	gs := grpc.NewServer()
	contextpb.RegisterContextupdateServer(gs, srv)
	go func() { _ = gs.Serve(lis) }()

	host, portStr, _ := net.SplitHostPort(lis.Addr().String())
	port, _ := strconv.Atoi(portStr)
	rpc.UpdateRpcConn(name, host, port)

	return func() {
		rpc.RemoveRpcConn(name)
		gs.Stop()
	}
}

func TestInvokeContextUpdate_Success(t *testing.T) {
	cleanup := startStubController(t, "ovnaction", &stubContextupdateServer{updated: true, message: "applied"})
	defer cleanup()

	if err := InvokeContextUpdate("ovnaction", "intent-1", "appcontext-1"); err != nil {
		t.Errorf("InvokeContextUpdate() = %v, want nil", err)
	}
}

func TestInvokeContextUpdate_ServerReportsFailure(t *testing.T) {
	cleanup := startStubController(t, "ovnaction", &stubContextupdateServer{updated: false, message: "rejected"})
	defer cleanup()

	if err := InvokeContextUpdate("ovnaction", "intent-1", "appcontext-1"); err == nil {
		t.Fatal("InvokeContextUpdate() = nil, want error when server reports not-updated")
	}
}

func TestInvokeContextUpdate_NoConnection(t *testing.T) {
	// No connection registered for this controller name.
	rpc.RemoveRpcConn("missing-controller")

	if err := InvokeContextUpdate("missing-controller", "intent-1", "appcontext-1"); err == nil {
		t.Error("InvokeContextUpdate() = nil, want error when controller connection is missing")
	}
}
