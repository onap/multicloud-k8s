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

import (
	"encoding/json"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	pkgerrors "github.com/pkg/errors"
)

// GetAppContextFromStateInfo loads the appcontext present in the StateInfo input
func GetAppContextFromId(ctxid string) (appcontext.AppContext, error) {
	var cc appcontext.AppContext
	_, err := cc.LoadAppContext(ctxid)
	if err != nil {
		return appcontext.AppContext{}, err
	}
	return cc, nil
}

// GetCurrentStateFromStatInfo gets the last (current) state from StateInfo
func GetCurrentStateFromStateInfo(s StateInfo) (StateValue, error) {
	alen := len(s.Actions)
	if alen == 0 {
		return StateEnum.Undefined, pkgerrors.Errorf("No state information")
	}
	return s.Actions[alen-1].State, nil
}

// GetLastContextFromStatInfo gets the last (most recent) context id from StateInfo
func GetLastContextIdFromStateInfo(s StateInfo) string {
	alen := len(s.Actions)
	if alen > 0 {
		return s.Actions[alen-1].ContextId
	} else {
		return ""
	}
}

// GetContextIdsFromStatInfo return a list of the unique AppContext Ids in the StateInfo
func GetContextIdsFromStateInfo(s StateInfo) []string {
	m := make(map[string]string)

	for _, a := range s.Actions {
		if a.ContextId != "" {
			m[a.ContextId] = ""
		}
	}

	ids := make([]string, len(m))
	i := 0
	for k := range m {
		ids[i] = k
		i++
	}

	return ids
}

func GetAppContextStatus(ctxid string) (appcontext.AppContextStatus, error) {

	ac, err := GetAppContextFromId(ctxid)
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}

	h, err := ac.GetCompositeAppHandle()
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	sh, err := ac.GetLevelHandle(h, "status")
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	s, err := ac.GetValue(sh)
	if err != nil {
		return appcontext.AppContextStatus{}, err
	}
	acStatus := appcontext.AppContextStatus{}
	js, _ := json.Marshal(s)
	json.Unmarshal(js, &acStatus)

	return acStatus, nil

}
