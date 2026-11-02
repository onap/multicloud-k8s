/*
 * Copyright 2019 Intel Corporation, Inc
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

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAuthConfiguration(t *testing.T) {
	for _, test := range []struct {
		name, enabled, contents                string
		unset, missing, wantError, wantEnabled bool
	}{
		{name: "unset defaults disabled", unset: true, missing: true},
		{name: "empty flag", enabled: "", missing: true, wantError: true},
		{name: "whitespace flag", enabled: " \t", missing: true, wantError: true},
		{name: "numeric false", enabled: "0", missing: true, wantError: true},
		{name: "numeric true", enabled: "1", contents: `[{"username":"first","password":"one"},{"username":"second","password":"two:with:colons"}]`, wantError: true},
		{name: "abbreviated false", enabled: "f", missing: true, wantError: true},
		{name: "abbreviated true", enabled: "t", contents: `[{"username":"first","password":"one"},{"username":"second","password":"two:with:colons"}]`, wantError: true},
		{name: "explicit disabled ignores file", enabled: "false", missing: true},
		{name: "normalized disabled ignores file", enabled: " \tFaLsE\n", missing: true},
		{name: "disabled ignores invalid JSON", enabled: "false", contents: "invalid"},
		{name: "invalid flag", enabled: "secret-sentinel", wantError: true},
		{name: "enabled requires path", enabled: "true", missing: true, wantError: true},
		{name: "empty file", enabled: "true", wantError: true},
		{name: "invalid JSON", enabled: "true", contents: `[{"username":"secret-sentinel"`, wantError: true},
		{name: "object not array", enabled: "true", contents: `{}`, wantError: true},
		{name: "null", enabled: "true", contents: `null`, wantError: true},
		{name: "empty users", enabled: "true", contents: `[]`, wantError: true},
		{name: "null user", enabled: "true", contents: `[null]`, wantError: true},
		{name: "missing password", enabled: "true", contents: `[{"username":"secret-sentinel"}]`, wantError: true},
		{name: "blank username", enabled: "true", contents: `[{"username":" \t","password":"secret-sentinel"}]`, wantError: true},
		{name: "blank password", enabled: "true", contents: `[{"username":"secret-sentinel","password":" \n"}]`, wantError: true},
		{name: "colon username", enabled: "true", contents: `[{"username":"secret:sentinel","password":"secret-sentinel"}]`, wantError: true},
		{name: "duplicate username", enabled: "true", contents: `[{"username":"secret-sentinel","password":"one"},{"username":"secret-sentinel","password":"two"}]`, wantError: true},
		{name: "wrong field type", enabled: "true", contents: `[{"username":"secret-sentinel","password":123}]`, wantError: true},
		{name: "trailing JSON", enabled: "true", contents: `[{"username":"secret-sentinel","password":"one"}] []`, wantError: true},
		{name: "two users", enabled: "true", wantEnabled: true, contents: `[{"username":"first","password":"one"},{"username":"second","password":"two:with:colons"}]`},
		{name: "normalized enabled", enabled: " \tTrUe\n", wantEnabled: true, contents: `[{"username":"first","password":"one"},{"username":"second","password":"two:with:colons"}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("MULTICLOUD_AUTH_ENABLED", test.enabled)
			if test.unset {
				if err := os.Unsetenv("MULTICLOUD_AUTH_ENABLED"); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("MULTICLOUD_AUTH_USERS_FILE", "")
			if !test.missing {
				path := filepath.Join(t.TempDir(), "users.json")
				if err := os.WriteFile(path, []byte(test.contents), 0600); err != nil {
					t.Fatal(err)
				}
				t.Setenv("MULTICLOUD_AUTH_USERS_FILE", path)
			}
			configuration, err := LoadAuthConfiguration()
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, wantError = %v", err, test.wantError)
			}
			if err != nil {
				if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "sentinel") {
					t.Fatal("configuration error exposed credentials")
				}
				return
			}
			if configuration.Enabled != test.wantEnabled {
				t.Fatal("unexpected enable state")
			}
			if configuration.Enabled && (len(configuration.Users) != 2 || configuration.Users[1].Password != "two:with:colons") {
				t.Fatal("configured users were not preserved")
			}
		})
	}
	t.Run("unreadable file error is redacted", func(t *testing.T) {
		t.Setenv("MULTICLOUD_AUTH_ENABLED", "true")
		t.Setenv("MULTICLOUD_AUTH_USERS_FILE", filepath.Join(t.TempDir(), "secret-sentinel"))
		_, err := LoadAuthConfiguration()
		if err == nil || strings.Contains(err.Error(), "secret-sentinel") {
			t.Fatal("missing file must fail without exposing its path")
		}
	})
}

func TestReadConfigurationFile(t *testing.T) {
	t.Run("Non Existent Configuration File", func(t *testing.T) {
		_, err := readConfigFile("filedoesnotexist.json")
		if err == nil {
			t.Fatal("ReadConfiguationFile: Expected Error, got nil")
		}
	})

	t.Run("Read Configuration File", func(t *testing.T) {
		conf, err := readConfigFile("../../mock_files/mock_configs/mock_config.json")
		if err != nil {
			t.Fatal("ReadConfigurationFile: Error reading file")
		}
		if conf.DatabaseType != "mock_db_test" {
			t.Fatal("ReadConfigurationFile: Incorrect entry read from file")
		}
	})
}
