#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o pipefail
set -o xtrace
set -o errexit
set -o nounset

echo 'start... vpp'
/usr/bin/vpp -c /etc/vpp/startup.conf
echo 'wait vpp be up ...'
until vppctl show ver; do
    sleep 1;
done

# Configure VPP for vFirewall
nic_protected=eth1
nic_unprotected=eth2
ip_protected_addr=$(ip addr show $nic_protected | grep inet | awk '{print $2}')
ip_unprotected_addr=$(ip addr show $nic_unprotected | grep inet | awk '{print $2}')

vppctl create host-interface name "$nic_protected"
vppctl create host-interface name "$nic_unprotected"

vppctl set int ip address "host-$nic_protected" "$ip_protected_addr"
vppctl set int ip address "host-$nic_unprotected" "$ip_unprotected_addr"

vppctl set int state "host-$nic_protected" up
vppctl set int state "host-$nic_unprotected" up

# Start HoneyComb
#/opt/honeycomb/honeycomb &>/dev/null &disown
/opt/honeycomb/honeycomb

# Start VES client
#/opt/VESreporting/vpp_measurement_reporter "$DCAE_COLLECTOR_IP" "$DCAE_COLLECTOR_PORT" eth1
