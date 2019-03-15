#!/bin/bash
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

csar_id=a1c5b53e-d7ab-11e8-85b7-525400e8c29a

# Setup
install_ovn_deps
populate_CSAR_ovn4nfv $csar_id

pushd ${CSAR_DIR}/${csar_id}
for net in ovn-priv-net ovn-port-net; do
    cleanup_network $net.yaml
    echo "Create OVN Network $net network"
    init_network $net.yaml
done
kubectl apply -f onap-ovn4nfvk8s-network.yaml
setup $ovn4nfv_deployment_name

# Test
deployment_pod=$(kubectl get pods | grep  $ovn4nfv_deployment_name | awk '{print $1}')
echo "===== $deployment_pod details ====="
kubectl exec -it $deployment_pod -- ip a
multus_nic=$(kubectl exec -it $deployment_pod -- ifconfig | grep "net1")
if [ -z "$multus_nic" ]; then
    echo "The $deployment_pod pod doesn't contain the net1 nic"
    exit 1
fi

# Teardown
teardown $ovn4nfv_deployment_name
cleanup_network ovn-priv-net.yaml
cleanup_network ovn-port-net.yaml
popd
