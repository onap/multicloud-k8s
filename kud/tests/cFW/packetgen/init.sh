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

# Configure VPP for vPacketGenerator
nic=eth0
ip_addr=$(ip addr show $nic | grep inet | awk '{print $2}')

vppctl create host-interface name "$nic"
vppctl set int state "host-$nic" up
vppctl set int ip address "host-$nic" "$ip_addr"
vppctl ip route add "$PROTECTED_NET_CIDR" via "$FW_IPADDR"

vppctl loop create
vppctl set int ip address loop0 11.22.33.1/24
vppctl set int state loop0 up

# Install packet streams
for i in $(seq 1 10); do
    cat <<EOL > "/opt/pg_streams/stream_fw_udp"
packet-generator new {
  name fw_udp$i
  rate 10
  node ip4-input
  size 64-64
  no-recycle
  interface loop0
  data {
    UDP: ${ip_addr%/*} -> $SINK_IPADDR
    UDP: 15320 -> 8080
    length 128 checksum 0 incrementing 1
  }
}
EOL
    vppctl exec "/opt/pg_streams/stream_fw_udp"
done
vppctl packet-generator enable

# Start HoneyComb
/opt/honeycomb/honeycomb
