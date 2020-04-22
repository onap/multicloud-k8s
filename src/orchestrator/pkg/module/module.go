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

// Client for using the services in the orchestrator
type Client struct {
	Project                *ProjectClient
	CompositeApp           *CompositeAppClient
	App                    *AppClient
	Controller             *ControllerClient
	GenericPlacementIntent *GenericPlacementIntentClient
	AppIntent              *AppIntentClient
	DeploymentIntentGroup  *DeploymentIntentGroupClient
	Intent                 *IntentClient
	CompositeProfile       *CompositeProfileClient
	AppProfile             *AppProfileClient
	// Add Clients for API's here
	Instantiation *InstantiationClient
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.Project = NewProjectClient()
	c.CompositeApp = NewCompositeAppClient()
	c.App = NewAppClient()
	c.Controller = NewControllerClient()
	c.GenericPlacementIntent = NewGenericPlacementIntentClient()
	c.AppIntent = NewAppIntentClient()
	c.DeploymentIntentGroup = NewDeploymentIntentGroupClient()
	c.Intent = NewIntentClient()
	c.CompositeProfile = NewCompositeProfileClient()
	c.AppProfile = NewAppProfileClient()
	// Add Client API handlers here
	c.Instantiation = NewInstantiationClient()
	return c
}

// It implements the interface for managing the ClusterProviders
const MAX_DESCRIPTION_LEN int = 1024
const MAX_USERDATA_LEN int = 4096

type Metadata struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"-"`
	UserData1   string `json:"userData1" yaml:"-"`
	UserData2   string `json:"userData2" yaml:"-"`
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
