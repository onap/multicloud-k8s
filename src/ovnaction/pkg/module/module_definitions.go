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
package module

import (
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	pkgerrors "github.com/pkg/errors"
)

const CNI_TYPE_OVN4NFV string = "ovn4nfv"

var CNI_TYPES = [...]string{CNI_TYPE_OVN4NFV}

// It implements the interface for managing the ClusterProviders
const MAX_DESCRIPTION_LEN int = 1024
const MAX_USERDATA_LEN int = 4096

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
}

type ClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
	tagContext string // attribute key name for context object in App Context
}

// Check for valid format Metadata
func IsValidMetadata(metadata Metadata) error {
	errs := validation.IsValidName(metadata.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid Metadata name=[%v], errors: %v", metadata.Name, errs)
	}

	errs = validation.IsValidString(metadata.Description, 0, MAX_DESCRIPTION_LEN, validation.VALID_ANY_STR)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid Metadata description=[%v], errors: %v", metadata.Description, errs)
	}

	errs = validation.IsValidString(metadata.UserData1, 0, MAX_DESCRIPTION_LEN, validation.VALID_ANY_STR)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid Metadata description=[%v], errors: %v", metadata.UserData1, errs)
	}

	errs = validation.IsValidString(metadata.UserData2, 0, MAX_DESCRIPTION_LEN, validation.VALID_ANY_STR)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid Metadata description=[%v], errors: %v", metadata.UserData2, errs)
	}

	return nil
}
