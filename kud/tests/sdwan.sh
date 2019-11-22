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

sdwan_pod_name=sdwan-ovn-pod
ovn_pod_name=ovn-pod
wan_interface=net0

function login {
    login_url=http://$1/cgi-bin/luci/
    echo $(wget -S --spider --post-data "luci_username=root&luci_password=" $login_url 2>&1 | grep sysauth= | sed -r 's/.*sysauth=([^;]+);.*/\1/')
}

function disable_ping {
    command_url=http://$2/cgi-bin/luci/admin/config/command
    command="uci set firewall.@rule[1].target='REJECT';fw3 reload"
    echo $(wget -S --spider --header="Cookie:sysauth=$1" --post-data "command=$command" $command_url 2>&1)
}

function enable_ping {
    command_url=http://$2/cgi-bin/luci/admin/config/command
    command="uci set firewall.@rule[1].target='ACCEPT';fw3 reload"
    echo $(wget -S --spider --header="Cookie:sysauth=$1" --post-data "command=$command" $command_url 2>&1)
}

function wait_for_pod {
    status_phase=""
    while [[ "$status_phase" != "Running" ]]; do
        new_phase="$(kubectl get pods -o wide | grep ^$1 | awk '{print $3}')"
        if [[ "$new_phase" != "$status_phase" ]]; then
            status_phase="$new_phase"
        fi
        if [[ "$new_phase" == "Err"* ]]; then
            exit 1
        fi
        sleep 2
    done
}

function wait_for_pod_namespace {
    status_phase=""
    while [[ "$status_phase" != "Running" ]]; do
        new_phase="$(kubectl get pods -o wide -n $2 | grep ^$1 | awk '{print $3}')"
        if [[ "$new_phase" != "$status_phase" ]]; then
            status_phase="$new_phase"
        fi
        if [[ "$new_phase" == "Err"* ]]; then
            exit 1
        fi
        sleep 2
    done
}

echo "Waiting for pods to be ready ..."
wait_for_pod $ovn_pod_name
wait_for_pod $sdwan_pod_name
echo "* Create pods success"

sdwan_pod_ip=$(kubectl get pods -o wide | grep ^$sdwan_pod_name | awk '{print $6}')
ovn_pod_ip=$(kubectl get pods -o wide | grep ^$ovn_pod_name | awk '{print $6}')
echo "SDWAN pod ip:"$sdwan_pod_ip
echo "OVN pod ip:"$ovn_pod_ip

echo "Login to sdwan ..."
security_token=""
while [[ "$security_token" == "" ]]; do
    echo "Get Security Token ..."
    security_token=$(login $sdwan_pod_ip)
    sleep 2
done
echo "* Security Token: "$security_token

kubectl exec $sdwan_pod_name ifconfig

sdwan_pod_wan_ip=$(kubectl exec $sdwan_pod_name ifconfig $wan_interface  | awk '/inet/{print $2}' | cut -f2 -d ":" | awk 'NR==1 {print $1}')
echo "Verify ping is work through wan interface between $sdwan_pod_name and $ovn_pod_name"
ping_result=$(kubectl exec $ovn_pod_name -- ping -c 3 $sdwan_pod_wan_ip)
if [[ $ping_result == *", 0% packet loss"* ]]; then
    echo "* Ping is work through wan interface"
else
    echo "* Test failed!"
    exit 1
fi

echo "Disable ping rule of wan interface ..."
ret=$(disable_ping $security_token $sdwan_pod_ip)

echo "Verify ping is not work through wan interface after ping rule disabled"
ping_result=$(kubectl exec $ovn_pod_name -- ping -c 3 $sdwan_pod_wan_ip 2>&1 || true)
if [[ $ping_result == *", 100% packet loss"* ]]; then
    echo "* Ping is disabled"
else
    echo "* Test failed!"
    exit 1
fi

echo "Enable ping rule of wan interface ..."
ret=$(enable_ping $security_token $sdwan_pod_ip)

echo "Verify ping is work through wan interface after ping rule enabled"
ping_result=$(kubectl exec $ovn_pod_name -- ping -c 3 $sdwan_pod_wan_ip)
if [[ $ping_result == *", 0% packet loss"* ]]; then
    echo "* Ping is enabled"
else
    echo "* Test failed!"
    exit 1
fi


echo "Test Completed!"
