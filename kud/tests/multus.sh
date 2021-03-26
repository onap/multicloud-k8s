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

source _common.sh
source _functions.sh

function generate_CRD_for_bridge_cni {
    local csar_id=$1
    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << NET > bridge-network.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: bridge-conf
spec:
  config: '{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "$multus_private_net_cidr"
    }
}'
NET
    popd
}

function generate_CRD_for_macvlan_cni {
    local csar_id=$1
    local master_name=$(ssh_cluster route | grep 'default' | awk '{print $8}' |head -n 1)
    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << NET > macvlan-network.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: macvlan-conf
spec:
  config: '{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "type": "macvlan",
    "master": "$master_name",
    "ipam": {
        "type": "host-local",
        "subnet": "$multus_private_net_cidr"
    }
}'
NET
    popd
}

function generate_CRD_for_ipvlan_cni {
    local csar_id=$1
    local master_name=$(ssh_cluster route | grep 'default' | awk '{print $8}' |head -n 1)
    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << NET > ipvlan-network.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: ipvlan-conf
spec:
  config: '{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "type": "ipvlan",
    "master": "$master_name",
    "ipam": {
        "type": "host-local",
        "subnet": "$multus_private_net_cidr"
    }
}'
NET
    popd
}

function generate_CRD_for_ptp_cni {
    local csar_id=$1
    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << NET > ptp-network.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: ptp-conf
spec:
  config: '{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "type": "ptp",
    "ipam": {
        "type": "host-local",
        "subnet": "$multus_private_net_cidr"
    }
}'
NET
    popd
}

csar_id=49408ca6-b75b-11e8-8076-525400feed26

# Setup
generate_CRD_for_bridge_cni $csar_id
generate_CRD_for_macvlan_cni $csar_id
generate_CRD_for_ipvlan_cni $csar_id
generate_CRD_for_ptp_cni $csar_id

pushd ${CSAR_DIR}/${csar_id}

kubectl apply -f bridge-network.yaml
kubectl apply -f macvlan-network.yaml
kubectl apply -f ipvlan-network.yaml
kubectl apply -f ptp-network.yaml

for cni in ${CNI_PLUGINS:-bridge macvlan ipvlan ptp}; do
    cat << DEPLOYMENT > $multus_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $multus_deployment_name
  labels:
    app: multus
spec:
  replicas: 1
  selector:
    matchLabels:
      app: multus
  template:
    metadata:
      labels:
        app: multus
      annotations:
        k8s.v1.cni.cncf.io/networks: ${cni}-conf
    spec:
      containers:
      - name: $multus_deployment_name
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT

setup $multus_deployment_name

# Test
deployment_pod=$(kubectl get pods | grep  $multus_deployment_name | awk '{print $1}')
echo "===== $deployment_pod details ====="
kubectl exec -it $deployment_pod -- ip a
multus_nic=$(kubectl exec -it $deployment_pod -- ip a)
net1_ip=$(kubectl exec -it $deployment_pod -- ifconfig net1 \
            | grep "inet addr" | awk '{ print $2}' |tr -d "addr:")

if [[ $multus_nic != *"net1"* ]]; then
    echo "The $deployment_pod pod doesn't contain the net1 nic"
    exit 1
else
    check_ip_range ${net1_ip} ${multus_private_net_cidr}
        if [[ $? -eq 1 ]]; then
            echo "unexpected ip range"
            exit 0
        fi
    echo "$cni Test Completed!"
fi

# Teardown
teardown $multus_deployment_name

done
popd
