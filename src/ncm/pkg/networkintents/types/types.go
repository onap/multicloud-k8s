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
package types

import (
	"strings"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/validation"
	pkgerrors "github.com/pkg/errors"
)

const VLAN_PROVIDER_NET_TYPE_VLAN string = "VLAN"
const VLAN_PROVIDER_NET_TYPE_DIRECT string = "DIRECT"

const SEPARATOR = "+"
const CONTEXT_CLUSTER_APP = "network-intents"

var PROVIDER_NET_TYPES = [...]string{VLAN_PROVIDER_NET_TYPE_VLAN, VLAN_PROVIDER_NET_TYPE_DIRECT}

const CNI_TYPE_OVN4NFV string = "ovn4nfv"

var CNI_TYPES = [...]string{CNI_TYPE_OVN4NFV}

type Ipv4Subnet struct {
	Subnet  string `json:"subnet" yaml:"subnet"` // CIDR notation, e.g. 172.16.33.0/24
	Name    string `json:"name" yaml:"name"`
	Gateway string `json:"gateway" yaml:"gateway"`       // IPv4 addre, e.g. 172.16.33.1/24
	Exclude string `json:"excludeIps" yaml:"excludeIps"` // space separated list of single IPs or ranges e.g. "172.16.33.2 172.16.33.5..172.16.33.10"
}

const VLAN_NODE_ANY = "any"
const VLAN_NODE_SPECIFIC = "specific"

var VLAN_NODE_SELECTORS = [...]string{VLAN_NODE_ANY, VLAN_NODE_SPECIFIC}

type Vlan struct {
	VlanId                string   `json:"vlanID" yaml:"vlanId"`
	ProviderInterfaceName string   `json:"providerInterfaceName" yaml:"providerInterfaceName"`
	LogicalInterfaceName  string   `json:"logicalInterfaceName" yaml:"logicalInterfaceName"`
	VlanNodeSelector      string   `json:"vlanNodeSelector" yaml:"vlanNodeSelector"`
	NodeLabelList         []string `json:"nodeLabelList" yaml:"nodeLabelList"`
}

// Check for valid format of an Ipv4Subnet
func ValidateSubnet(sub Ipv4Subnet) error {
	// verify subnet is in valid cidr format
	err := validation.IsIpv4Cidr(sub.Subnet)
	if err != nil {
		return pkgerrors.Wrap(err, "invalid subnet")
	}

	// just a size check on interface name - system dependent
	errs := validation.IsValidName(sub.Name)
	if len(errs) > 0 {
		return pkgerrors.Errorf("Invalid subnet name=[%v], errors: %v", sub.Name, errs)
	}

	// verify gateway is in valid cidr format
	if len(sub.Gateway) > 0 {
		err = validation.IsIpv4Cidr(sub.Gateway)
		if err != nil {
			return pkgerrors.Wrap(err, "invalid gateway")
		}
	}

	// verify excludeIps is composed of space separated ipv4 addresses and
	// ipv4 address ranges separated by '..'
	for _, value := range strings.Fields(sub.Exclude) {
		for _, ip := range strings.SplitN(value, "..", 2) {
			err = validation.IsIpv4(ip)
			if err != nil {
				return pkgerrors.Errorf("invalid ipv4 exclude list %v", sub.Exclude)
			}
		}
	}
	return nil
}
