#!/bin/bash

mkdir -p /opt/config/
echo "$protected_net_gw"           > /opt/config/protected_net_gw.txt
echo "$protected_private_net_cidr" > /opt/config/unprotected_net.txt

# NOTE: this script executes $ route add -net 192.168.10.0 netmask 255.255.255.0 gw 192.168.20.100
# which results in this error if doesn't have all nics required  -> SIOCADDRT: File exists
./v_sink_init.sh
sleep infinity
