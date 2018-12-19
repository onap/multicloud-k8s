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

package main

import (
	pkgerrors "github.com/pkg/errors"
	"k8splugin/plugins/network/v1"
)

// Err is the error message to be sent during functional testing
var Err string

// NetworkName is the output used for functional tests
var NetworkName string

// CreateNetwork resource
func CreateNetwork(network *v1.OnapNetwork) (string, error) {
	if Err != "" {
		return "", pkgerrors.New(Err)
	}
	return NetworkName, nil
}

// DeleteNetwork resource
func DeleteNetwork(name string) error {
	if Err != "" {
		return pkgerrors.New(Err)
	}
	return nil
}
