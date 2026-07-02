/*
Copyright 2026 Deutsche Telekom AG
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package connector

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	c := Init("context-123")
	if c == nil {
		t.Fatal("Init returned nil")
	}
	if c.cid != "context-123" {
		t.Fatalf("Expected cid 'context-123', got %q", c.cid)
	}
	if c.Clients == nil {
		t.Fatal("Expected an initialized (non-nil) Clients map")
	}
	if len(c.Clients) != 0 {
		t.Fatalf("Expected an empty Clients map, got %d entries", len(c.Clients))
	}
}

func TestInitStringifiesNonStringId(t *testing.T) {
	c := Init(42)
	if c.cid != "42" {
		t.Fatalf("Expected cid '42', got %q", c.cid)
	}
}

func TestGetKubeConfigInvalidClusterName(t *testing.T) {
	testCases := []struct {
		label       string
		clusterName string
	}{
		{"Name without a '+' separator", "notvalid"},
		{"Name with too many '+' separators", "a+b+c"},
	}

	for _, tCase := range testCases {
		t.Run(tCase.label, func(t *testing.T) {
			_, err := getKubeConfig(tCase.clusterName)
			if err == nil {
				t.Fatalf("Expected an error for cluster name %q, got nil", tCase.clusterName)
			}
		})
	}
}

func TestRemoveClient(t *testing.T) {
	c := Init("remove-me")

	// Create the on-disk directory RemoveClient is expected to clean up.
	dir := basePath + "/" + c.cid
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to set up test directory: %s", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("Test directory was not created: %s", err)
	}

	c.RemoveClient()

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("Expected directory %q to have been removed", dir)
	}
}
