/*
 * Copyright 2018 Intel Corporation, Inc
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

package helm

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Represents the template that is used to create a particular
// resource in Kubernetes
type KubernetesResourceTemplate struct {
	// Tracks the apiVersion and Kind of the resource
	GVK schema.GroupVersionKind
	// Path to the file that contains the resource info
	FilePath string
}

// KubernetesResource is the resource that is created in Kubernetes
// It is the type that will be used for tracking a resource.
// Any future information such as status, time can be added here
// for tracking.
type KubernetesResource struct {
	// Tracks the apiVersion and Kind of the resource
	GVK schema.GroupVersionKind
	// Name of resource in Kubernetes
	Name string
}
