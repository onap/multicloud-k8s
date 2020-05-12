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
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
)

// Client for using the services in the orchestrator
type Client struct {
	Project                *ProjectClient
	CompositeApp           *CompositeAppClient
	App                    *AppClient
	Controller             *controller.ControllerClient
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
	c.Controller = controller.NewControllerClient()
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
