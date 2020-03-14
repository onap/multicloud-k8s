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
	"strings"

	"github.com/onap/multicloud-k8s/src/orchestrator/utils"
	pkgerrors "github.com/pkg/errors"
)

// It implements the interface for managing the ClusterProviders
type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserData1   string `json:"userData1"`
	UserData2   string `json:"userData2"`
}

type ClientDbInfo struct {
	storeName  string // name of the mongodb collection to use for client documents
	tagMeta    string // attribute key name for the json data of a client document
	tagContent string // attribute key name for the file data of a client document
}

type Ipv4Subnet struct {
	Subnet  string `json:"subnet"` // CIDR notation, e.g. 172.16.33.0/24
	Name    string `json:"name"`
	Gateway string `json:"gateway"`    // IPv4 addre, e.g. 172.16.33.1
	Exclude string `json:"excludeIps"` // space separated list of single IPs or ranges e.g. "172.16.33.2 172.16.33.5..172.16.33.10"
}

type Vlan struct {
	VlanId                string   `json:"vlanID"`
	ProviderInterfaceName string   `json:"providerInterfaceName"`
	LogicalInterfaceName  string   `json:"logicalInterfaceName"`
	VlanNodeSelector      string   `json:"vlanNodeSelector"`
	NodeLabelList         []string `json:"nodeLabelList"`
}

// Check for valid format of an Ipv4Subnet
func ValidateSubnet(sub Ipv4Subnet) error {
	// verify subnet is in valid cidr format
	err := utils.IsIpv4Cidr(sub.Subnet)
	if err != nil {
		return pkgerrors.Wrap(err, "invalid subnet")
	}

	// verify gateway is a valid ipv4 address
	if len(sub.Gateway) > 0 {
		err = utils.IsIpv4(sub.Gateway)
		if err != nil {
			return pkgerrors.Wrap(err, "invalid gateway")
		}
	}

	// verify excludeIps is composed of space separated ipv4 addresses and
	// ipv4 address ranges separated by '..'
	for _, value := range strings.Fields(sub.Exclude) {
		for _, ip := range strings.SplitN(value, "..", 2) {
			err = utils.IsIpv4(ip)
			if err != nil {
				return pkgerrors.Errorf("invalid ipv4 exclude list %v", sub.Exclude)
			}
		}
	}
	return nil
}
