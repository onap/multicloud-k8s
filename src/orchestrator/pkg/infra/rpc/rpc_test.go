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

package rpc

import (
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
)

// resetConnections clears the package-level connection map so each test starts
// from a clean state (the map is process-global, mirroring the singleton the
// controllers use at runtime).
func resetConnections() {
	mutex.Lock()
	defer mutex.Unlock()
	rpcConnections = make(map[string]rpcInfo)
}

func TestGetRpcConn_UnknownReturnsNil(t *testing.T) {
	resetConnections()
	if conn := GetRpcConn("does-not-exist"); conn != nil {
		t.Errorf("GetRpcConn(unknown) = %v, want nil", conn)
	}
}

func TestUpdateRpcConn_AddsConnection(t *testing.T) {
	resetConnections()
	// grpc.Dial with WithInsecure is lazy - no server needs to exist for the
	// ClientConn to be created and cached.
	UpdateRpcConn("rsync", "localhost", 9031)

	conn := GetRpcConn("rsync")
	if conn == nil {
		t.Fatal("expected a cached connection for 'rsync' after UpdateRpcConn")
	}
	if len(rpcConnections) != 1 {
		t.Errorf("rpcConnections len = %d, want 1", len(rpcConnections))
	}
}

func TestUpdateRpcConn_SameHostPortIsNoOp(t *testing.T) {
	resetConnections()
	UpdateRpcConn("rsync", "localhost", 9031)
	first := GetRpcConn("rsync")

	// Re-registering with identical host/port must keep the same connection.
	UpdateRpcConn("rsync", "localhost", 9031)
	second := GetRpcConn("rsync")

	if first != second {
		t.Error("expected the same *grpc.ClientConn when host/port are unchanged")
	}
	if len(rpcConnections) != 1 {
		t.Errorf("rpcConnections len = %d, want 1", len(rpcConnections))
	}
}

func TestUpdateRpcConn_ChangedHostPortReplacesConnection(t *testing.T) {
	resetConnections()
	UpdateRpcConn("rsync", "localhost", 9031)
	first := GetRpcConn("rsync")

	// A mismatched host/port closes the old connection and dials a new one.
	UpdateRpcConn("rsync", "localhost", 9032)
	second := GetRpcConn("rsync")

	if second == nil {
		t.Fatal("expected a connection after host/port change")
	}
	if first == second {
		t.Error("expected a new *grpc.ClientConn after host/port change")
	}
	if len(rpcConnections) != 1 {
		t.Errorf("rpcConnections len = %d, want 1", len(rpcConnections))
	}
}

func TestRemoveRpcConn(t *testing.T) {
	resetConnections()
	UpdateRpcConn("rsync", "localhost", 9031)
	UpdateRpcConn("ovnaction", "localhost", 9032)

	RemoveRpcConn("rsync")

	if GetRpcConn("rsync") != nil {
		t.Error("expected 'rsync' connection to be removed")
	}
	if GetRpcConn("ovnaction") == nil {
		t.Error("expected 'ovnaction' connection to remain")
	}
	// Removing an unknown name must be a safe no-op.
	RemoveRpcConn("does-not-exist")
	if len(rpcConnections) != 1 {
		t.Errorf("rpcConnections len = %d, want 1", len(rpcConnections))
	}
}

func TestCloseAllRpcConn(t *testing.T) {
	resetConnections()
	UpdateRpcConn("rsync", "localhost", 9031)
	UpdateRpcConn("ovnaction", "localhost", 9032)

	CloseAllRpcConn()

	// CloseAllRpcConn closes the connections but (matching the current
	// implementation) does not clear the map; the connections should now be
	// closed. We assert both remain addressable so the call is a safe drain.
	if GetRpcConn("rsync") == nil || GetRpcConn("ovnaction") == nil {
		t.Error("expected connections to remain in the map after CloseAllRpcConn")
	}
}

func TestCreateClientConn_Insecure(t *testing.T) {
	// Default config has grpc-enable-tls = "disable", so the insecure dial
	// option path is taken and a lazy connection is returned without error.
	config.SetConfigValue("GrpcEnableTLS", "disable")

	conn, err := createClientConn("localhost", 9031)
	if err != nil {
		t.Fatalf("createClientConn err: %v", err)
	}
	if conn == nil {
		t.Fatal("expected a non-nil connection for the insecure path")
	}
	_ = conn.Close()
}

func TestCreateClientConn_TLSFallsBackToTestdataCA(t *testing.T) {
	// With TLS enabled and no CA file configured, createClientConn falls back
	// to the grpc testdata ca.pem, which exists in the module, so the dial
	// still succeeds (lazily).
	config.SetConfigValue("GrpcEnableTLS", "enable")
	config.SetConfigValue("GrpcCAFile", "")
	defer config.SetConfigValue("GrpcEnableTLS", "disable")

	conn, err := createClientConn("localhost", 9031)
	if err != nil {
		t.Fatalf("createClientConn (tls) err: %v", err)
	}
	if conn == nil {
		t.Fatal("expected a non-nil connection for the TLS path")
	}
	_ = conn.Close()
}
