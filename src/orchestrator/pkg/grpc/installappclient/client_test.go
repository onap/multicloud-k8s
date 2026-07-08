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

package installappclient

import (
	"context"
	"net"
	"strconv"
	"testing"

	installpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/installapp"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	"google.golang.org/grpc"
)

// stubInstallappServer is a controllable in-process implementation of the
// generated InstallappServer used to drive the client through a real gRPC
// transport (an ephemeral loopback listener). Each field decides the reply.
type stubInstallappServer struct {
	installpb.UnimplementedInstallappServer
	installed   bool
	installMsg  string
	uninstalled bool
	uninstMsg   string
}

func (s *stubInstallappServer) InstallApp(ctx context.Context, req *installpb.InstallAppRequest) (*installpb.InstallAppResponse, error) {
	return &installpb.InstallAppResponse{
		AppContextInstalled:      s.installed,
		AppContextInstallMessage: s.installMsg,
	}, nil
}

func (s *stubInstallappServer) UninstallApp(ctx context.Context, req *installpb.UninstallAppRequest) (*installpb.UninstallAppResponse, error) {
	return &installpb.UninstallAppResponse{
		AppContextUninstalled:      s.uninstalled,
		AppContextUninstallMessage: s.uninstMsg,
	}, nil
}

// startStubRsync spins up the stub server on a loopback ephemeral port and
// registers it in the rpc connection map under "rsync" (the name the client
// looks up). It returns a cleanup func.
func startStubRsync(t *testing.T, srv installpb.InstallappServer) func() {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	gs := grpc.NewServer()
	installpb.RegisterInstallappServer(gs, srv)
	go func() { _ = gs.Serve(lis) }()

	host, portStr, _ := net.SplitHostPort(lis.Addr().String())
	port, _ := strconv.Atoi(portStr)
	rpc.UpdateRpcConn("rsync", host, port)

	return func() {
		rpc.RemoveRpcConn("rsync")
		gs.Stop()
	}
}

func TestNewRsyncInfo(t *testing.T) {
	info := NewRsyncInfo("rsync", "1.2.3.4", 9031)
	if info.RsyncName != "rsync" || info.hostName != "1.2.3.4" || info.portNumber != 9031 {
		t.Errorf("NewRsyncInfo = %+v, want {rsync 1.2.3.4 9031}", info)
	}
}

func TestInvokeInstallApp_Success(t *testing.T) {
	cleanup := startStubRsync(t, &stubInstallappServer{installed: true, installMsg: "ok"})
	defer cleanup()

	if err := InvokeInstallApp("appcontext-1"); err != nil {
		t.Errorf("InvokeInstallApp() = %v, want nil", err)
	}
}

func TestInvokeInstallApp_ServerReportsFailure(t *testing.T) {
	cleanup := startStubRsync(t, &stubInstallappServer{installed: false, installMsg: "boom"})
	defer cleanup()

	err := InvokeInstallApp("appcontext-1")
	if err == nil {
		t.Fatal("InvokeInstallApp() = nil, want error when server reports not-installed")
	}
}

func TestInvokeInstallApp_NoConnection(t *testing.T) {
	// No server, and RsyncInfo unset -> initRsyncClient cannot establish a
	// connection, so InvokeInstallApp must fail with the client-unavailable error.
	rpc.RemoveRpcConn("rsync")
	rsyncInfo = RsyncInfo{}

	if err := InvokeInstallApp("appcontext-1"); err == nil {
		t.Error("InvokeInstallApp() = nil, want error when no connection is available")
	}
}

func TestInvokeUninstallApp_Success(t *testing.T) {
	cleanup := startStubRsync(t, &stubInstallappServer{uninstalled: true, uninstMsg: "removed"})
	defer cleanup()

	if err := InvokeUninstallApp("appcontext-1"); err != nil {
		t.Errorf("InvokeUninstallApp() = %v, want nil", err)
	}
}

func TestInvokeUninstallApp_ServerReportsFailure(t *testing.T) {
	cleanup := startStubRsync(t, &stubInstallappServer{uninstalled: false, uninstMsg: "still there"})
	defer cleanup()

	if err := InvokeUninstallApp("appcontext-1"); err == nil {
		t.Fatal("InvokeUninstallApp() = nil, want error when server reports not-uninstalled")
	}
}

func TestInvokeUninstallApp_NoConnection(t *testing.T) {
	rpc.RemoveRpcConn("rsync")

	if err := InvokeUninstallApp("appcontext-1"); err == nil {
		t.Error("InvokeUninstallApp() = nil, want error when no connection is available")
	}
}
