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

FUNCTIONS_DIR=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd)

source /etc/environment

function print_msg {
    local msg=$1
    local RED='\033[0;31m'
    local NC='\033[0m'

    echo -e "${RED} $msg ---------------------------------------${NC}"
}

function get_ovn_central_address {
    ansible_ifconfig=$(ansible ovn-central[0] -i ${FUNCTIONS_DIR}/../hosting_providers/vagrant/inventory/hosts.ini -m shell -a "ifconfig ${OVN_CENTRAL_INTERFACE} |grep \"inet addr\" |awk '{print \$2}' |awk -F: '{print \$2}'")
    if [[ $ansible_ifconfig != *CHANGED* ]]; then
        echo "Fail to get the OVN central IP address from ${OVN_CENTRAL_INTERFACE} nic"
        exit
    fi
    echo "$(echo ${ansible_ifconfig#*>>} | tr '\n' ':')6641"
}

# init_network() - This function creates the OVN resouces required by the test
function init_network {
    local fname=$1
    local router_name="ovn4nfv-master"

    name=$(cat $fname | yq '.spec.name' | xargs)
    subnet=$(cat $fname  | yq '.spec.subnet' | xargs)
    gateway=$(cat $fname  | yq '.spec.gateway' | xargs)
    ovn_central_address=$(get_ovn_central_address)

    router_mac=$(printf '00:00:00:%02X:%02X:%02X' $((RANDOM%256)) $((RANDOM%256)) $((RANDOM%256)))
    ovn-nbctl --may-exist --db tcp:$ovn_central_address ls-add $name -- set logical_switch $name other-config:subnet=$subnet external-ids:gateway_ip=$gateway
    ovn-nbctl --may-exist --db tcp:$ovn_central_address lrp-add $router_name rtos-$name $router_mac $gateway
    ovn-nbctl --may-exist --db tcp:$ovn_central_address lsp-add $name stor-$name -- set logical_switch_port stor-$name type=router options:router-port=rtos-$name addresses=\"$router_mac\"
}

# cleanup_network() - This function removes the OVN resources created for the test
function cleanup_network {
    local fname=$1

    name=$(cat $fname | yq '.spec.name' | xargs)
    ovn_central_address=$(get_ovn_central_address)

    for cmd in "ls-del $name" "lrp-del rtos-$name" "lsp-del stor-$name"; do
        ovn-nbctl --if-exist --db tcp:$ovn_central_address $cmd
    done
}

function _checks_args {
    if [[ -z $1 ]]; then
        echo "Missing CSAR ID argument"
        exit 1
    fi
    if [[ -z $CSAR_DIR ]]; then
        echo "CSAR_DIR global environment value is empty"
        exit 1
    fi
    mkdir -p ${CSAR_DIR}/${1}
}

# destroy_deployment() - This function ensures that a specific deployment is
# destroyed in Kubernetes
function destroy_deployment {
    local deployment_name=$1

    echo "$(date +%H:%M:%S) - $deployment_name : Destroying deployment"
    kubectl delete deployment $deployment_name --ignore-not-found=true --now
    while kubectl get deployment $deployment_name &>/dev/null; do
        echo "$(date +%H:%M:%S) - $deployment_name : Destroying deployment"
    done
}

# recreate_deployment() - This function destroys an existing deployment and
# creates an new one based on its yaml file
function recreate_deployment {
    local deployment_name=$1

    destroy_deployment $deployment_name
    kubectl create -f $deployment_name.yaml
}

# wait_deployment() - Wait process to Running status on the Deployment's pods
function wait_deployment {
    local deployment_name=$1

    status_phase=""
    while [[ $status_phase != "Running" ]]; do
        new_phase=$(kubectl get pods | grep  $deployment_name | awk '{print $3}')
        if [[ $new_phase != $status_phase ]]; then
            echo "$(date +%H:%M:%S) - $deployment_name : $new_phase"
            status_phase=$new_phase
        fi
        if [[ $new_phase == "Err"* ]]; then
            exit 1
        fi
    done
}

# setup() - Base testing setup shared among functional tests
function setup {
    if ! $(kubectl version &>/dev/null); then
        echo "This funtional test requires kubectl client"
        exit 1
    fi
    for deployment_name in $@; do
        recreate_deployment $deployment_name
    done
    sleep 5
    for deployment_name in $@; do
        wait_deployment $deployment_name
    done
}

# teardown() - Base testing teardown function
function teardown {
    for deployment_name in $@; do
        destroy_deployment $deployment_name
    done
}
test_folder=${FUNCTIONS_DIR}
