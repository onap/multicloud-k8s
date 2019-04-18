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
	"testing"
)

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
