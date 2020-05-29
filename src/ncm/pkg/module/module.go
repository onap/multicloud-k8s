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
	"github.com/onap/multicloud-k8s/src/ncm/pkg/cluster"
	"github.com/onap/multicloud-k8s/src/ncm/pkg/networkintents"
	"github.com/onap/multicloud-k8s/src/ncm/pkg/scheduler"
)

// Client for using the services in the ncm
type Client struct {
	Cluster     *cluster.ClusterClient
	Network     *networkintents.NetworkClient
	ProviderNet *networkintents.ProviderNetClient
	Scheduler   *scheduler.SchedulerClient
	// Add Clients for API's here
}

// NewClient creates a new client for using the services
func NewClient() *Client {
	c := &Client{}
	c.Cluster = cluster.NewClusterClient()
	c.Network = networkintents.NewNetworkClient()
	c.ProviderNet = networkintents.NewProviderNetClient()
	c.Scheduler = scheduler.NewSchedulerClient()
	// Add Client API handlers here
	return c
}
