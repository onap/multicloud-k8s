#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

map_list=""
function create_vlan {
    local dev=$1
    local id=$2
    local name=$3

    # Create VLAN for device
    sudo ip link add link $dev name $name type vlan id $id
    #sudo ip addr add $ip dev $name
    sudo ip link set $name up
}

# Create provider network for interface
function create_provider_network {
    local provider_name=$1
    local interface=$2

    bridge_name=br-$provider_name
    network_name=nw_$provider_name
    port_name=server-localnet_$provider_name

    # Create OVS bridge and move the interface to the bridge
    sudo ovs-vsctl --may-exist add-br $bridge_name
    sudo ovs-vsctl --may-exist add-port $bridge_name $interface

    #Create OVN Switch
    sudo ovn-nbctl --may-exist ls-add  $provider_name
    # Add port of type localnet to the Switch
    sudo ovn-nbctl --may-exist lsp-add  $provider_name $port_name
    sudo ovn-nbctl lsp-set-addresses  $port_name unknown
    sudo ovn-nbctl lsp-set-type  $port_name localnet
    #Set port with the network name to map to ovs bridge
    sudo ovn-nbctl lsp-set-options $port_name network_name=$network_name
    # Prepare bridge to network mapping for OVS
    map_list=${map_list}${network_name}:${bridge_name},
}

create_vlan eth1 100 eth1.100
create_vlan eth1 200 eth1.200

provider_net1=prod-net1
provider_net2=prod-net2

create_provider_network $provider_net1 eth1.100
create_provider_network $provider_net2 eth1.200

#Set OVS with the bridge to network mapping
map_list=${map_list%?}
sudo ovs-vsctl set open . external-ids:ovn-bridge-mappings=$map_list


