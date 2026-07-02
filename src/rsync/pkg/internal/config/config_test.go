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

package config

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestReadConfigFile(t *testing.T) {
	t.Run("Non existent configuration file returns defaults and error", func(t *testing.T) {
		conf, err := readConfigFile("filedoesnotexist.json")
		if err == nil {
			t.Fatal("Expected an error for a missing file, got nil")
		}
		// On error the default configuration is still returned.
		if conf == nil {
			t.Fatal("Expected a default configuration, got nil")
		}
		if conf.DatabaseType != "mongo" {
			t.Fatalf("Expected default DatabaseType 'mongo', got %q", conf.DatabaseType)
		}
	})

	t.Run("Values in file override defaults", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "k8sconfig.json")
		content := `{"database-type": "etcd", "service-port": "9999"}`
		if err := ioutil.WriteFile(file, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write temp config: %s", err)
		}

		conf, err := readConfigFile(file)
		if err != nil {
			t.Fatalf("Unexpected error reading config: %s", err)
		}
		if conf.DatabaseType != "etcd" {
			t.Fatalf("Expected overridden DatabaseType 'etcd', got %q", conf.DatabaseType)
		}
		if conf.ServicePort != "9999" {
			t.Fatalf("Expected overridden ServicePort '9999', got %q", conf.ServicePort)
		}
		// A field not present in the file keeps its default value.
		if conf.EtcdIP != "127.0.0.1" {
			t.Fatalf("Expected default EtcdIP '127.0.0.1', got %q", conf.EtcdIP)
		}
	})

	t.Run("Malformed json returns an error", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "bad.json")
		if err := ioutil.WriteFile(file, []byte("{not valid json"), 0644); err != nil {
			t.Fatalf("Failed to write temp config: %s", err)
		}
		if _, err := readConfigFile(file); err == nil {
			t.Fatal("Expected an error for malformed json, got nil")
		}
	})
}

func TestDefaultConfiguration(t *testing.T) {
	conf := defaultConfiguration()
	if conf.DatabaseAddress != "127.0.0.1" {
		t.Fatalf("Expected default DatabaseAddress '127.0.0.1', got %q", conf.DatabaseAddress)
	}
	if conf.ServicePort != "9015" {
		t.Fatalf("Expected default ServicePort '9015', got %q", conf.ServicePort)
	}
	if conf.KubernetesLabelName != "k8splugin.io/rb-instance-id" {
		t.Fatalf("Expected default KubernetesLabelName, got %q", conf.KubernetesLabelName)
	}
}

func TestGetConfiguration(t *testing.T) {
	// GetConfiguration lazily loads (and caches) the global config. In the test
	// environment there is no k8sconfig.json, so it must fall back to defaults.
	gConfig = nil
	conf := GetConfiguration()
	if conf == nil {
		t.Fatal("Expected a configuration, got nil")
	}
	if conf.DatabaseType != "mongo" {
		t.Fatalf("Expected default DatabaseType 'mongo', got %q", conf.DatabaseType)
	}
	// The second call must return the cached instance rather than reloading.
	if GetConfiguration() != conf {
		t.Fatal("Expected GetConfiguration to return the cached instance")
	}
}

func TestSetConfigValue(t *testing.T) {
	// Reset the cached config so each sub-test starts from a known baseline.
	reset := func() { gConfig = nil }

	t.Run("Sets an existing string field", func(t *testing.T) {
		reset()
		c := SetConfigValue("DatabaseType", "etcd")
		if c.DatabaseType != "etcd" {
			t.Fatalf("Expected DatabaseType 'etcd', got %q", c.DatabaseType)
		}
	})

	t.Run("Empty key is a no-op", func(t *testing.T) {
		reset()
		before := GetConfiguration().DatabaseType
		c := SetConfigValue("", "etcd")
		if c.DatabaseType != before {
			t.Fatalf("Expected DatabaseType to remain %q, got %q", before, c.DatabaseType)
		}
	})

	t.Run("Empty value is a no-op", func(t *testing.T) {
		reset()
		before := GetConfiguration().DatabaseType
		c := SetConfigValue("DatabaseType", "")
		if c.DatabaseType != before {
			t.Fatalf("Expected DatabaseType to remain %q, got %q", before, c.DatabaseType)
		}
	})

	t.Run("Unknown field is ignored", func(t *testing.T) {
		reset()
		// Should not panic and should leave the config otherwise intact.
		c := SetConfigValue("NoSuchField", "value")
		if c == nil {
			t.Fatal("Expected a configuration, got nil")
		}
	})
}
