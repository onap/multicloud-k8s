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
source _common_test.sh
source _functions.sh

# populate_CSAR_ovn4nfv() - Create content used for OVN4NFV functional test
function populate_CSAR_provider_network {
    local csar_id=$1

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << MULTUS_NET > onap-ovn4nfvk8s-network.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $ovn_multus_network_name
spec:
  config: '{
      "cniVersion": "0.3.1",
      "name": "ovn4nfv-k8s-plugin",
      "type": "ovn4nfvk8s-cni"
    }'
MULTUS_NET

    cat << NETWORK > ovn-virt-net1.yaml
apiVersion: v1
kind: onapNetwork
metadata:
  name: ovn-virt-net1
  cnitype : ovn4nfvk8s
spec:
  name: ovn-virt-net1
  subnet: 10.1.20.0/24
  gateway: 10.1.20.1/24
NETWORK

    cat << NETWORK > ovn-virt-net2.yaml
apiVersion: v1
kind: onapNetwork
metadata:
  name: ovn-virt-net2
  cnitype : ovn4nfvk8s
spec:
  name: ovn-virt-net2
  subnet: 10.1.21.0/24
  gateway: 10.1.21.1/24
NETWORK

    cat << DEPLOYMENT > firewall.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: firewall
  labels:
    app: ovn4nfv
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ovn4nfv
  template:
    metadata:
      labels:
        app: ovn4nfv
      annotations:
        k8s.v1.cni.cncf.io/networks: '[{ "name": "$ovn_multus_network_name"}]'
        ovnNetwork: '[ { "name": ""ovn-virt-net1"", "interface": "net0" , "defaultGateway": "false", "ipAddress":"10.1.20.2"},
                      { "name": "prod-net1", "interface": "net1", "defaultGateway": "false", "ipAddress":"10.1.5.1/24"}]'
        ovnNetworkRoutes: '[{ "dst": "0.0.0.0/0", "gw": "10.1.20.3", "dev": "net0" }]'

    spec:
      containers:
      - name: firewall
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT

    cat << DEPLOYMENT > webcache.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webcache
  labels:
    app: ovn4nfv
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ovn4nfv
  template:
    metadata:
      labels:
        app: ovn4nfv
      annotations:
        k8s.v1.cni.cncf.io/networks: '[{ "name": "$ovn_multus_network_name"}]'
        ovnNetwork: '[{ "name": "ovn-virt-net1", "interface": "net0" , "defaultGateway": "false", "ipAddress":"10.1.20.3"},
                      { "name": "ovn-virt-net2", "interface": "net1" , "defaultGateway": "false", "ipAddress":"10.1.21.2"}]'
        ovnNetworkRoutes: '[{ "dst": "10.1.5.0/24", "gw": "10.1.20.2", "dev": "net0" },
                            { "dst": "0.0.0.0/0", "gw": "10.1.21.3", "dev": "net1" }]'

    spec:
      containers:
      - name: webcache
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT

    cat << DEPLOYMENT > sdwan.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: sdwan
  labels:
    app: ovn4nfv
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ovn4nfv
  template:
    metadata:
      labels:
        app: ovn4nfv
      annotations:
        k8s.v1.cni.cncf.io/networks: '[{ "name": "$ovn_multus_network_name"}]'
        ovnNetwork: '[ { "name": ""ovn-virt-net2"", "interface": "net0" , "defaultGateway": "false", "ipAddress":"10.1.21.3"},
                      { "name": "prod-net2", "interface": "net1", "defaultGateway": "false", "ipAddress":"10.1.10.2/24"}]'
        ovnNetworkRoutes: '[{ "dst": "0.0.0.0/0", "gw": "10.1.10.1", "dev": "net1" },
                            { "dst": "10.1.5.0/24", "gw": "10.1.21.2", "dev": "net0" },
                            { "dst": "10.1.20.0/24", "gw": "10.1.21.2", "dev": "net0" }]'

    spec:
      containers:
      - name: sdwan
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT
    popd
}

csar_id=d5718572-3b9a-11e9-b210-d663bd873dda
# Setup
install_ovn_deps
populate_CSAR_provider_network $csar_id

pushd ${CSAR_DIR}/${csar_id}
for net in ovn-virt-net1 ovn-virt-net2; do
    cleanup_network $net.yaml
    echo "Create OVN Network $net network"
    init_network $net.yaml
done
kubectl apply -f onap-ovn4nfvk8s-network.yaml
setup firewall webcache sdwan

# Test
deployment_pod=$(kubectl get pods | grep  firewall | awk '{print $1}')
echo "===== $deployment_pod details ====="
kubectl exec -it $deployment_pod -- ip a
multus_nic=$(kubectl exec -it $deployment_pod -- ifconfig | grep "net1")
if [ -z "$multus_nic" ]; then
    echo "The $deployment_pod pod doesn't contain the net1 nic"
    exit 1
fi

# Teardown
teardown firewall webcache sdwan
cleanup_network ovn-virt-net1.yaml
cleanup_network ovn-virt-net2.yaml
popd



