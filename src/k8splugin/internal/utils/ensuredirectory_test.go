/*
Copyright 2018 Intel Corporation.
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

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDirectory(t *testing.T) {
	t.Run("Creates parent directory for a file path", func(t *testing.T) {
		base := t.TempDir()
		filePath := filepath.Join(base, "nested", "deeper", "file.txt")

		if err := EnsureDirectory(filePath); err != nil {
			t.Fatalf("EnsureDirectory returned unexpected error: %v", err)
		}

		info, err := os.Stat(filepath.Dir(filePath))
		if err != nil {
			t.Fatalf("Expected parent directory to exist: %v", err)
		}
		if !info.IsDir() {
			t.Fatalf("Expected parent path to be a directory")
		}
	})

	t.Run("Succeeds when directory already exists", func(t *testing.T) {
		base := t.TempDir()
		filePath := filepath.Join(base, "existing.txt")

		if err := EnsureDirectory(filePath); err != nil {
			t.Fatalf("EnsureDirectory returned unexpected error on existing dir: %v", err)
		}
	})
}
