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
	"testing"
)

func TestIsValidNetworkChain(t *testing.T) {
	t.Run("Valid Chains", func(t *testing.T) {
		validchains := []string{
			"app=abc,net1,app=xyz",
			"app=abc, net1, app=xyz",
			" app=abc , net1 , app=xyz ",
			"app.kubernets.io/name=abc,net1,app.kubernets.io/name=xyz",
			"app.kubernets.io/name=abc,net1,app.kubernets.io/name=xyz, net2, anotherlabel=wex",
		}
		for _, chain := range validchains {
			err := validateNetworkChain(chain)
			if err != nil {
				t.Errorf("Valid network chain failed to pass: %v %v", chain, err)
			}
		}
	})

	t.Run("Invalid Chains", func(t *testing.T) {
		invalidchains := []string{
			"",
			"app=abc,net1,app= xyz",
			"app=abc,net1,xyz",
			"app=abc,net1",
			"app.kubernets.io/name=abc,net1,=xyz",
			"abcdefg",
		}
		for _, chain := range invalidchains {
			err := validateNetworkChain(chain)
			if err == nil {
				t.Errorf("Invalid network chain passed: %v", chain)
			}
		}
	})
}
