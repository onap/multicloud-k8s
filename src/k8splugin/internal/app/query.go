/*
 * Copyright 2018 Intel Corporation, Inc
 * Copyright © 2021 Samsung Electronics
 * Copyright © 2021 Orange
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

package app

import (
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/helm"
	"k8s.io/apimachinery/pkg/runtime/schema"

	pkgerrors "github.com/pkg/errors"
)

// QueryStatus is what is returned when status is queried for an instance
type QueryStatus struct {
	ResourceCount   int32            `json:"resourceCount"`
	ResourcesStatus []ResourceStatus `json:"resourcesStatus"`
}

// QueryManager is an interface exposes the instantiation functionality
type QueryManager interface {
	Query(namespace, cloudRegion, apiVersion, kind, name, labels, id string) (QueryStatus, error)
}

// QueryClient implements the InstanceManager interface
// It will also be used to maintain some localized state
type QueryClient struct {
	storeName string
	tagInst   string
}

// NewQueryClient returns an instance of the QueryClient
// which implements the InstanceManager
func NewQueryClient() *QueryClient {
	return &QueryClient{
		storeName: "rbdef",
		tagInst:   "instance",
	}
}

// Query returns state of instance's filtered resources
func (v *QueryClient) Query(namespace, cloudRegion, apiVersion, kind, name, labels, id string) (QueryStatus, error) {

	//Read the status from the DD

	k8sClient := KubernetesClient{}
	err := k8sClient.Init(cloudRegion, id)
	if err != nil {
		return QueryStatus{}, pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}

	var resourcesStatus []ResourceStatus
	if labels != "" {
		resList, err := k8sClient.queryResources(apiVersion, kind, labels, namespace)
		if err != nil {
			return QueryStatus{}, pkgerrors.Wrap(err, "Querying Resources")
		}
		// If user specifies both label and name, we want to pick up only single resource from these matching label
		if name != "" {
			//Assigning 0-length, because we may actually not find matching name
			resourcesStatus = make([]ResourceStatus, 0)
			for _, res := range resList {
				if res.Name == name {
					resourcesStatus = append(resourcesStatus, res)
					break
				}
			}
		} else {
			resourcesStatus = resList
		}
	} else if name != "" {
		resIdentifier := helm.KubernetesResource{
			Name: name,
			GVK:  schema.FromAPIVersionAndKind(apiVersion, kind),
		}
		res, err := k8sClient.GetResourceStatus(resIdentifier, namespace)
		if err != nil {
			return QueryStatus{}, pkgerrors.Wrap(err, "Querying Resource")
		}
		resourcesStatus = []ResourceStatus{res}
	} else {
		resList, err := k8sClient.queryResources(apiVersion, kind, labels, namespace)
		if err != nil {
			return QueryStatus{}, pkgerrors.Wrap(err, "Querying Resources")
		}
		resourcesStatus = resList
	}

	resp := QueryStatus{
		ResourceCount:   int32(len(resourcesStatus)),
		ResourcesStatus: resourcesStatus,
	}
	return resp, nil
}
