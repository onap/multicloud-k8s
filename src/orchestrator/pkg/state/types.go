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

import "time"

// StateInfo struct is used to maintain the values for state, contextid, (and other)
// information about resources which can be instantiated via rsync.
// The last Actions entry holds the current state of the container object.
type StateInfo struct {
	Actions []ActionEntry `json:"actions"`
}

// ActionEntry is used to keep track of the time an action (e.g. Created, Instantiate, Terminate) was invoked
// For actions where an AppContext is relevent, the ContextId field will be non-zero length
type ActionEntry struct {
	State     StateValue `json:"state"`
	ContextId string     `json:"instance"`
	TimeStamp time.Time  `json:"time"`
}

type StateValue = string

type states struct {
	Undefined    StateValue
	Created      StateValue
	Approved     StateValue
	Applied      StateValue
	Instantiated StateValue
	Terminated   StateValue
}

var StateEnum = &states{
	Undefined:    "Undefined",
	Created:      "Created",
	Approved:     "Approved",
	Applied:      "Applied",
	Instantiated: "Instantiated",
	Terminated:   "Terminated",
}
