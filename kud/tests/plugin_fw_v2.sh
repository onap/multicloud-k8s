#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

source _common_test.sh
source _functions.sh

# TODO KUBECONFIG may be a list of paths
KUBECONFIG_PATH="${KUBECONFIG:-$HOME/.kube/config}"
DEMO_FOLDER="${DEMO_FOLDER:-$test_folder/../demo}"

clusters="${KUD_PLUGIN_FW_CLUSTERS:-$(cat <<EOF
[
  {
    "metadata": {
      "name": "edge01",
      "description": "description of edge01",
      "userData1": "edge01 user data 1",
      "userData2": "edge01 user data 2"
    },
    "file": "$KUBECONFIG_PATH"
  }
]
EOF
)}"

function cluster_names {
    echo $clusters | jq -e -r '.[].metadata.name'
}

function cluster_metadata {
    cat<<EOF | jq .
{
  "metadata": $(echo $clusters | jq -e -r --arg name "$1" '.[]|select(.metadata.name==$name)|.metadata')
}
EOF
}

function cluster_file {
    echo $clusters | jq -e -r --arg name "$1" '.[]|select(.metadata.name==$name)|.file'
}

ARGS=()
while [[ $# -gt 0 ]]; do
    arg="$1"

    case $arg in
        "--external" )
            service_host=$(kubectl cluster-info | grep "Kubernetes master" | \
                awk -F ":" '{print $2}' | awk -F "//" '{print $2}')
            shift
            ;;
        * )
            ARGS+=("$1")
            shift
            ;;
    esac
done
set -- "${ARGS[@]}" # restore positional parameters

service_host=${service_host:-"localhost"}

CSAR_DIR="/opt/csar"
csar_id="4bf66240-a0be-4ce2-aebd-a01df7725f16"

function populate_CSAR_compositevfw_helm {
    _checks_args "$1"
    pushd "${CSAR_DIR}/$1"
    print_msg "Create Helm Chart Archives for compositevfw"
    rm -f *.tar.gz
    tar -czf packetgen.tar.gz -C $DEMO_FOLDER/composite-firewall packetgen
    tar -czf firewall.tar.gz -C $DEMO_FOLDER/composite-firewall firewall
    tar -czf sink.tar.gz -C $DEMO_FOLDER/composite-firewall sink
    tar -czf profile.tar.gz -C $DEMO_FOLDER/composite-firewall manifest.yaml override_values.yaml
    popd
}

project="testvfw"
composite_app="compositevfw"
version="v1"
deployment_intent_group="vfw_deployment_intent_group"

function setup {
    install_deps
    populate_CSAR_compositevfw_helm "$csar_id"
    cat <<EOF >plugin_fw_v2_config.yaml
orchestrator:
  host: ${service_host}
  port: 30415
clm:
  host: ${service_host}
  port: 30461
ncm:
  host: ${service_host}
  port: 30431
ovnaction:
  host: ${service_host}
  port: 30471
dcm:
  host: ${service_host}
  port: 30477
gac:
  host: ${service_host}
  port: 30491
dtc:
 host: ${service_host}
 port: 30481
EOF
    cat <<EOF >plugin_fw_v2_values.yaml
ClusterProvider: vfw-cluster-provider
ClusterLabel: LabelA
Clusters:
EOF
    echo $clusters | jq -r '.[] | "- Name: \(.metadata.name)\n  KubeConfig: \(.file)"' >>plugin_fw_v2_values.yaml
    cat <<EOF >>plugin_fw_v2_values.yaml
EmcoProviderNetwork: emco-private-net
UnprotectedProviderNetwork: unprotected-private-net
ProtectedNetwork: protected-private-net
Project: ${project}
LogicalCloud: lcadmin
CompositeApp: ${composite_app}
Version: ${version}
PackagesPath: ${CSAR_DIR}/${csar_id}
CompositeProfile: vfw_composite-profile
DeploymentIntentGroup: ${deployment_intent_group}
Release: fw0
DeploymentIntentsInGroup: vfw_deploy_intents
GenericPlacementIntent: generic-placement-intent
OvnActionIntent: vfw_ovnaction_intent
EOF
}

function call_emcoctl {
    rc=$1
    shift
    # retry due to known issue with emcoctl and instantiating/terminating multiple resources
    try=0
    until [[ $(emcoctl $@ | awk '/Response Code:/ {code=$3} END{print code}') =~ $rc ]]; do
        if [[ $try -lt 10 ]]; then
            sleep 1s
        else
            return 1
        fi
        try=$((try + 1))
    done
    return 0
}

function createData {
    call_emcoctl 2.. --config plugin_fw_v2_config.yaml apply -f plugin_fw_v2.yaml -v plugin_fw_v2_values.yaml
}

function getData {
    emcoctl --config plugin_fw_v2_config.yaml get -f plugin_fw_v2.yaml -v plugin_fw_v2_values.yaml
}

function deleteData {
    call_emcoctl 4.. --config plugin_fw_v2_config.yaml delete -f plugin_fw_v2.yaml -v plugin_fw_v2_values.yaml
}

function statusVfw {
    emcoctl --config plugin_fw_v2_config.yaml get projects/${project}/composite-apps/${composite_app}/${version}/deployment-intent-groups/${deployment_intent_group}/status
}

function waitForVfw {
    for try in {0..59}; do
        sleep 1
        new_phase="$(emcoctl --config plugin_fw_v2_config.yaml get projects/${project}/composite-apps/${composite_app}/${version}/deployment-intent-groups/${deployment_intent_group}/status | awk '/Response: / {print $2}' | jq -r .status)"
        echo "$(date +%H:%M:%S) - Filter=[$*] : $new_phase"
        if [[ "$new_phase" == "$1" ]]; then
            return 0
        fi
    done
}

function usage {
    echo "Usage: $0 setup|create|get|destroy|status"
    echo "    setup - creates the emcoctl files and packages needed for vfw"
    echo "    create - creates all ncm, ovnaction, clm resources needed for vfw"
    echo "    get - queries all resources in ncm, ovnaction, clm resources created for vfw"
    echo "    destroy - deletes all resources in ncm, ovnaction, clm resources created for vfw"
    echo "    status - get status of deployed resources"
    echo ""
    echo "    a reasonable test sequence:"
    echo "    1.  setup"
    echo "    2.  create"
    echo "    3.  destroy"

    exit
}

if [[ "$#" -gt 0 ]] ; then
    case "$1" in
        "setup" ) setup ;;
        "create" ) createData ;;
        "get" ) getData ;;
        "status" ) statusVfw ;;
        "wait" ) waitForVfw "Instantiated" ;;
        "delete" ) deleteData ;;
        *) usage ;;
    esac
else
    setup
    createData

    print_msg "[BEGIN] Basic checks for instantiated resource"
    print_msg "Wait for deployment to be instantiated"
    waitForVfw "Instantiated"
    for name in $(cluster_names); do
        print_msg "Check that networks were created on cluster $name"
        file=$(cluster_file "$name")
        KUBECONFIG=$file kubectl get network protected-private-net -o name
        KUBECONFIG=$file kubectl get providernetwork emco-private-net -o name
        KUBECONFIG=$file kubectl get providernetwork unprotected-private-net -o name
    done
    # Give some time for the Pods to show up on the clusters.  kubectl
    # wait may return with "error: no matching resources found" if the
    # Pods have not started yet.
    sleep 30s
    for name in $(cluster_names); do
        print_msg "Wait for all pods to start on cluster $name"
        file=$(cluster_file "$name")
        KUBECONFIG=$file kubectl wait pod -l release=fw0 --for=condition=Ready --timeout=5m
    done
    # TODO: Provide some health check to verify vFW work
    print_msg "Not waiting for vFW to fully install as no further checks are implemented in testcase"
    #print_msg "Waiting 8minutes for vFW installation"
    #sleep 8m
    print_msg "[END] Basic checks for instantiated resource"

    print_msg "Delete deployment"
    deleteData
fi
