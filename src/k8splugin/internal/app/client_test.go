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

package app

import (
	"reflect"
	"testing"
)

func TestGetKubeClient(t *testing.T) {
	t.Run("Successfully create Kube Client", func(t *testing.T) {

		clientset, err := GetKubeClient("../../mock_files/mock_configs/mock_config")
		if err != nil {
			t.Fatalf("TestGetKubeClient returned an error (%s)", err)
		}

		if reflect.TypeOf(clientset).Name() != "Clientset" {
			t.Fatalf("TestGetKubeClient returned :\n result=%v\n expected=%v", clientset, "Clientset")
		}

	})
}
