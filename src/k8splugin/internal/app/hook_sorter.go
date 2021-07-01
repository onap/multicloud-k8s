/*
Copyright © 2021 Nokia Bell Labs
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
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	sortLib "sort"
)

// sortByHookWeight does an in-place sort of hooks by their supplied weight.
func sortByHookWeight(hooks []*helm.Hook) []*helm.Hook {
	hs := newHookWeightSorter(hooks)
	sortLib.Sort(hs)
	return hs.hooks
}

type hookWeightSorter struct {
	hooks []*helm.Hook
}

func newHookWeightSorter(h []*helm.Hook) *hookWeightSorter {
	return &hookWeightSorter{
		hooks: h,
	}
}

func (hs *hookWeightSorter) Len() int { return len(hs.hooks) }

func (hs *hookWeightSorter) Swap(i, j int) {
	hs.hooks[i], hs.hooks[j] = hs.hooks[j], hs.hooks[i]
}

func (hs *hookWeightSorter) Less(i, j int) bool {
	if hs.hooks[i].Hook.Weight == hs.hooks[j].Hook.Weight {
		return hs.hooks[i].Hook.Name < hs.hooks[j].Hook.Name
	}
	return hs.hooks[i].Hook.Weight < hs.hooks[j].Hook.Weight
}

