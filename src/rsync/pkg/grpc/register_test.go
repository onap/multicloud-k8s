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

package grpc

import (
	"os"
	"testing"
)

func TestGetServerHostPort_Defaults(t *testing.T) {
	os.Unsetenv(ENV_RSYNC_NAME)
	os.Unsetenv("RSYNC_SERVICE_HOST")
	os.Unsetenv("RSYNC_SERVICE_PORT")

	host, port := GetServerHostPort()
	if host != default_host {
		t.Errorf("host = %q, want %q", host, default_host)
	}
	if port != default_port {
		t.Errorf("port = %d, want %d", port, default_port)
	}
}

func TestGetServerHostPort_FromEnv(t *testing.T) {
	os.Unsetenv(ENV_RSYNC_NAME) // use the default service name "rsync"
	os.Setenv("RSYNC_SERVICE_HOST", "10.0.0.5")
	os.Setenv("RSYNC_SERVICE_PORT", "12345")
	defer func() {
		os.Unsetenv("RSYNC_SERVICE_HOST")
		os.Unsetenv("RSYNC_SERVICE_PORT")
	}()

	host, port := GetServerHostPort()
	if host != "10.0.0.5" {
		t.Errorf("host = %q, want 10.0.0.5", host)
	}
	if port != 12345 {
		t.Errorf("port = %d, want 12345", port)
	}
}

func TestGetServerHostPort_CustomServiceName(t *testing.T) {
	// A custom RSYNC_NAME changes which <NAME>_SERVICE_* vars are consulted.
	os.Setenv(ENV_RSYNC_NAME, "myrsync")
	os.Setenv("MYRSYNC_SERVICE_HOST", "myhost")
	os.Setenv("MYRSYNC_SERVICE_PORT", "9999")
	defer func() {
		os.Unsetenv(ENV_RSYNC_NAME)
		os.Unsetenv("MYRSYNC_SERVICE_HOST")
		os.Unsetenv("MYRSYNC_SERVICE_PORT")
	}()

	host, port := GetServerHostPort()
	if host != "myhost" || port != 9999 {
		t.Errorf("GetServerHostPort() = %q,%d, want myhost,9999", host, port)
	}
}

func TestGetServerHostPort_InvalidPortFallsBackToDefault(t *testing.T) {
	os.Unsetenv(ENV_RSYNC_NAME)
	os.Unsetenv("RSYNC_SERVICE_HOST")
	os.Setenv("RSYNC_SERVICE_PORT", "not-a-number")
	defer os.Unsetenv("RSYNC_SERVICE_PORT")

	_, port := GetServerHostPort()
	if port != default_port {
		t.Errorf("port = %d, want default %d on invalid port value", port, default_port)
	}
}
