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

package state

import "github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"

// GetAppContextFromStateInfo loads the appcontext present in the StateInfo input
func GetAppContextFromStateInfo(s StateInfo) (appcontext.AppContext, error) {
	var cc appcontext.AppContext
	_, err := cc.LoadAppContext(s.ContextId)
	if err != nil {
		return appcontext.AppContext{}, err
	}
	return cc, nil
}
