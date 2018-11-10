/*
Copyright 2018 Intel Corporation.
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

package v1

import (
	"encoding/json"

	pkgerrors "github.com/pkg/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// OnapNetwork describes an ONAP network resouce
type OnapNetwork struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`
	Spec              OnapNetworkSpec `json:"spec"`
}

// OnapNetworkSpec is the spec for OnapNetwork resource
type OnapNetworkSpec struct {
	Config string `json:"config"`
}

// DeepCopyObject returns a generically typed copy of an object
func (in OnapNetwork) DeepCopyObject() runtime.Object {
	out := OnapNetwork{}
	out.TypeMeta = in.TypeMeta
	out.ObjectMeta = in.ObjectMeta
	out.Spec = in.Spec

	return &out
}

// GetObjectKind
func (in OnapNetwork) GetObjectKind() schema.ObjectKind {
	return &in.TypeMeta
}

// DecodeConfig content
func (in OnapNetwork) DecodeConfig() (map[string]interface{}, error) {
	var raw map[string]interface{}

	if in.Spec.Config == "" {
		return nil, pkgerrors.New("Invalid configuration value")
	}

	if err := json.Unmarshal([]byte(in.Spec.Config), &raw); err != nil {
		return nil, pkgerrors.Wrap(err, "JSON unmarshalling error")
	}

	return raw, nil
}
