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

package status

import (
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/state"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// StatusQueryParam defines the type of the query parameter
type StatusQueryParam = string
type queryparams struct {
	Instance StatusQueryParam // identify which AppContext to use - default is latest
	Summary  StatusQueryParam // only show high level summary
	All      StatusQueryParam // include basic resource information
	Detail   StatusQueryParam // show resource details
	Rsync    StatusQueryParam // select rsync (appcontext) data as source for query
	App      StatusQueryParam // filter results by specified app(s)
	Cluster  StatusQueryParam // filter results by specified cluster(s)
	Resource StatusQueryParam // filter results by specified resource(s)
}

// StatusQueryEnum defines the set of valid query parameter strings
var StatusQueryEnum = &queryparams{
	Instance: "instance",
	Summary:  "summary",
	All:      "all",
	Detail:   "detail",
	Rsync:    "rsync",
	App:      "app",
	Cluster:  "cluster",
	Resource: "resource",
}

type StatusResult struct {
	Name          string                 `json:"name,omitempty,inline"`
	State         state.StateInfo        `json:"states,omitempty,inline"`
	Status        appcontext.StatusValue `json:"status,omitempty,inline"`
	RsyncStatus   map[string]int         `json:"rsync-status,omitempty,inline"`
	ClusterStatus map[string]int         `json:"cluster-status,omitempty,inline"`
	Apps          []AppStatus            `json:"apps,omitempty,inline"`
}

type AppStatus struct {
	Name     string          `json:"name,omitempty"`
	Clusters []ClusterStatus `json:"clusters,omitempty"`
}

type ClusterStatus struct {
	ClusterProvider string           `json:"cluster-provider,omitempty"`
	Cluster         string           `json:"cluster,omitempty"`
	Resources       []ResourceStatus `json:"resources,omitempty"`
}

type ResourceStatus struct {
	Gvk           schema.GroupVersionKind `json:"GVK,omitempty"`
	Name          string                  `json:"name,omitempty"`
	Detail        interface{}             `json:"detail,omitempty"`
	RsyncStatus   string                  `json:"rsync-status,omitempty"`
	ClusterStatus string                  `json:"cluster-status,omitempty"`
}
