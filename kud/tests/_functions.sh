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

FUNCTIONS_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

source /etc/environment

function print_msg {
    local msg=$1
    local RED='\033[0;31m'
    local NC='\033[0m'

    echo -e "${RED} $msg ---------------------------------------${NC}"
}

function get_ovn_central_address {
    #Reuse OVN_CENTRAL_ADDRESS if available (bypassable by --force flag)
    if [[ "${1:-}" != "--force" ]] && [[ -n "${OVN_CENTRAL_ADDRESS:-}" ]]; then
        echo "${OVN_CENTRAL_ADDRESS}"
        return 0
    fi

    local remote_command="ip address show dev $OVN_CENTRAL_INTERFACE primary"
    declare -a ansible_command=(ansible ovn-central[0] -i \
                "${FUNCTIONS_DIR}/../hosting_providers/vagrant/inventory/hosts.ini")
    declare -a filter=(awk -F '[ \t/]+' \
                'BEGIN {r=1} {for (i=1; i<=NF; i++) if ($i == "inet") {print $(i+1); r=0}} END {exit r}')
    local result

    #Determine OVN_CENTRAL_INTERFACE address
    if ! result="$("${ansible_command[@]}" -a "${remote_command}")"; then
        echo "Ansible error for remote host ovn-central[0]" >&2
        return 1
    else
        if [[ "${result}" != *CHANGED* ]]; then
            echo "Failed to execute command on remote host ovn-central[0]" >&2
            return 2
        else
            if ! result="$("${filter[@]}" <<< "${result}")"; then
                echo "Failed to retrieve interface address from command output" >&2
                return 3
            else
                echo "${result}:6641"
            fi
        fi
    fi
}

function call_api {
    #Runs curl with passed flags and provides
    #additional error handling and debug information

    #Function outputs server response body
    #and performs validation of http_code

    local status
    local curl_response_file="$(mktemp -p /tmp)"
    local curl_common_flags=(-s -w "%{http_code}" -o "${curl_response_file}")
    local command=(curl "${curl_common_flags[@]}" "$@")

    echo "[INFO] Running '${command[@]}'" >&2
    if ! status="$("${command[@]}")"; then
        echo "[ERROR] Internal curl error! '$status'" >&2
        cat "${curl_response_file}"
        rm "${curl_response_file}"
        return 2
    else
        echo "[INFO] Server replied with status: ${status}" >&2
        cat "${curl_response_file}"
        rm "${curl_response_file}"
        if [[ "${status:0:1}" =~ [45] ]]; then
            return 1
        else
            return 0
        fi
    fi
}

function delete_resource {
    #Issues DELETE http call to provided endpoint
    #and further validates by following GET request

    call_api -X DELETE "$1"
    ! call_api -X GET "$1" >/dev/null
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

# wait_for_pod() - Wait until first pod matched by kubectl filters is in running status
function wait_for_pod {
    #Example usage:
    # wait_for_pods example_pod
    # wait_for_pods --namespace test different_pod
    # wait_for_pods -n test -l app=plugin_test

    status_phase=""
    while [[ "$status_phase" != "Running" ]]; do
        new_phase="$(kubectl get pods -o 'go-template={{ index .items 0 "status" "phase" }}' "$@" )"
        if [[ "$new_phase" != "$status_phase" ]]; then
            echo "$(date +%H:%M:%S) - Filter=[$*] : $new_phase"
            status_phase="$new_phase"
        fi
        if [[ "$new_phase" == "Err"* ]]; then
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
