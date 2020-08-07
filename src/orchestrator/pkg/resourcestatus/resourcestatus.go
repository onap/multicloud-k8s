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

package resourcestatus

// ResourceStatus struct is used to maintain the rsync status for resources in the appcontext
// that rsync is synchronizing to clusters
type ResourceStatus struct {
	Status RsyncStatus
}

type RsyncStatus = string

type statusValues struct {
	Pending  RsyncStatus
	Applied  RsyncStatus
	Failed   RsyncStatus
	Retrying RsyncStatus
	Deleted  RsyncStatus
}

var RsyncStatusEnum = &statusValues{
	Pending:  "Pending",
	Applied:  "Applied",
	Failed:   "Failed",
	Retrying: "Retrying",
	Deleted:  "Deleted",
}
