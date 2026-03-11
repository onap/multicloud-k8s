/*
Copyright 2025  Deutsche Telekom.
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

// Package model provides local DTO types that mirror the JSON responses
// from the k8splugin REST API, avoiding a compile-time dependency on
// k8splugin's internal packages.
package model

// InstanceRequest contains the parameters needed for instantiation of profiles.
type InstanceRequest struct {
	RBName         string            `json:"rb-name"`
	RBVersion      string            `json:"rb-version"`
	ProfileName    string            `json:"profile-name"`
	ReleaseName    string            `json:"release-name"`
	CloudRegion    string            `json:"cloud-region"`
	Labels         map[string]string `json:"labels"`
	OverrideValues map[string]string `json:"override-values"`
}

// InstanceMiniResponse is the short form returned when listing instances.
type InstanceMiniResponse struct {
	ID          string          `json:"id"`
	Request     InstanceRequest `json:"request"`
	ReleaseName string          `json:"release-name"`
	Namespace   string          `json:"namespace"`
}

// ResourceStatus holds the runtime data for a single Kubernetes resource.
type ResourceStatus struct {
	Name   string                 `json:"name"`
	GVK    GroupVersionKind       `json:"GVK"`
	Status map[string]interface{} `json:"status"`
}

// GroupVersionKind identifies a Kubernetes resource type.
type GroupVersionKind struct {
	Group   string `json:"Group"`
	Version string `json:"Version"`
	Kind    string `json:"Kind"`
}

// InstanceStatus is returned when the status of an instance is queried.
type InstanceStatus struct {
	Request         InstanceRequest  `json:"request"`
	Ready           bool             `json:"ready"`
	ResourceCount   int32            `json:"resourceCount"`
	ResourcesStatus []ResourceStatus `json:"resourcesStatus"`
}

// Connection contains connectivity information for a cloud region.
type Connection struct {
	CloudRegion           string                 `json:"cloud-region"`
	CloudOwner            string                 `json:"cloud-owner"`
	Kubeconfig            string                 `json:"kubeconfig"`
	OtherConnectivityList ConnectivityRecordList `json:"other-connectivity-list"`
}

// ConnectivityRecordList covers lists of connectivity records.
type ConnectivityRecordList struct {
	ConnectivityRecords []map[string]string `json:"connectivity-records"`
}
