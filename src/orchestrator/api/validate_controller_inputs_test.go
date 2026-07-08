/*
 * Copyright 2020 Intel Corporation, Inc
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

package api

import (
	"strings"
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/types"
)

func TestValidateControllerInputs(t *testing.T) {
	validController := controller.Controller{
		Metadata: types.Metadata{Name: "testController"},
		Spec: controller.ControllerSpec{
			Host:     "localhost",
			Port:     8080,
			Type:     controller.CONTROLLER_TYPE_PLACEMENT,
			Priority: 1,
		},
	}

	testCases := []struct {
		label       string
		inp         controller.Controller
		expectError bool
		errContains string
	}{
		{
			label:       "Valid controller",
			inp:         validController,
			expectError: false,
		},
		{
			label: "Invalid metadata name",
			inp: func() controller.Controller {
				c := validController
				c.Metadata.Name = ""
				return c
			}(),
			expectError: true,
			errContains: "Invalid controller metadata",
		},
		{
			label: "Invalid host name",
			inp: func() controller.Controller {
				c := validController
				c.Spec.Host = ""
				return c
			}(),
			expectError: true,
			errContains: "Invalid host name",
		},
		{
			label: "Invalid port",
			inp: func() controller.Controller {
				c := validController
				c.Spec.Port = 70000
				return c
			}(),
			expectError: true,
			errContains: "Invalid controller port",
		},
		{
			label: "Invalid type",
			inp: func() controller.Controller {
				c := validController
				c.Spec.Type = "notARealType"
				return c
			}(),
			expectError: true,
			errContains: "Invalid controller type",
		},
		{
			label: "Invalid priority",
			inp: func() controller.Controller {
				c := validController
				c.Spec.Priority = 0
				return c
			}(),
			expectError: true,
			errContains: "Invalid controller priority",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			err := validateControllerInputs(testCase.inp)
			if testCase.expectError {
				if err == nil {
					t.Fatalf("validateControllerInputs expected an error but got none")
				}
				if !strings.Contains(err.Error(), testCase.errContains) {
					t.Fatalf("validateControllerInputs returned unexpected error: %s (expected to contain %q)", err, testCase.errContains)
				}
			} else if err != nil {
				t.Fatalf("validateControllerInputs returned an unexpected error: %s", err)
			}
		})
	}
}
