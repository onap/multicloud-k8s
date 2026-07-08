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
	os.Unsetenv(ENV_OVNACTION_NAME)
	os.Unsetenv("OVNACTION_SERVICE_HOST")
	os.Unsetenv("OVNACTION_SERVICE_PORT")

	host, port := GetServerHostPort()
	if host != default_host {
		t.Errorf("host = %q, want %q", host, default_host)
	}
	if port != default_port {
		t.Errorf("port = %d, want %d", port, default_port)
	}
}

func TestGetServerHostPort_FromEnv(t *testing.T) {
	os.Unsetenv(ENV_OVNACTION_NAME) // use the default service name "ovnaction"
	os.Setenv("OVNACTION_SERVICE_HOST", "10.0.0.9")
	os.Setenv("OVNACTION_SERVICE_PORT", "23456")
	defer func() {
		os.Unsetenv("OVNACTION_SERVICE_HOST")
		os.Unsetenv("OVNACTION_SERVICE_PORT")
	}()

	host, port := GetServerHostPort()
	if host != "10.0.0.9" {
		t.Errorf("host = %q, want 10.0.0.9", host)
	}
	if port != 23456 {
		t.Errorf("port = %d, want 23456", port)
	}
}

func TestGetServerHostPort_CustomServiceName(t *testing.T) {
	os.Setenv(ENV_OVNACTION_NAME, "myovn")
	os.Setenv("MYOVN_SERVICE_HOST", "ovnhost")
	os.Setenv("MYOVN_SERVICE_PORT", "8888")
	defer func() {
		os.Unsetenv(ENV_OVNACTION_NAME)
		os.Unsetenv("MYOVN_SERVICE_HOST")
		os.Unsetenv("MYOVN_SERVICE_PORT")
	}()

	host, port := GetServerHostPort()
	if host != "ovnhost" || port != 8888 {
		t.Errorf("GetServerHostPort() = %q,%d, want ovnhost,8888", host, port)
	}
}

func TestGetServerHostPort_InvalidPortFallsBackToDefault(t *testing.T) {
	os.Unsetenv(ENV_OVNACTION_NAME)
	os.Unsetenv("OVNACTION_SERVICE_HOST")
	os.Setenv("OVNACTION_SERVICE_PORT", "bogus")
	defer os.Unsetenv("OVNACTION_SERVICE_PORT")

	_, port := GetServerHostPort()
	if port != default_port {
		t.Errorf("port = %d, want default %d on invalid port value", port, default_port)
	}
}
