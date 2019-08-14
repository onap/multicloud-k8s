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
source _common_test.sh
source _functions.sh

csar_id=a1c5b53e-d7ab-11e8-85b7-525400e8c29a

# Setup
populate_CSAR_ovn4nfv $csar_id

pushd ${CSAR_DIR}/${csar_id}
for net in ovn-priv-net ovn-port-net; do
    echo "Create OVN Network $net network"
    kubectl apply -f $net.yaml
done
kubectl apply -f onap-ovn4nfvk8s-network.yaml
setup $ovn4nfv_deployment_name

# Test
deployment_pod=$(kubectl get pods | grep  $ovn4nfv_deployment_name | awk '{print $1}')
echo "===== $deployment_pod details ====="
kubectl exec -it $deployment_pod -- ip a

ovn_nic=$(kubectl exec -it $deployment_pod -- ip a )
if [[ $ovn_nic != *"net1"* ]]; then
    echo "The $deployment_pod pod doesn't contain the net1 nic"
    exit 1
else
    echo "Test Completed!"
fi

# Teardown
teardown $ovn4nfv_deployment_name
for net in ovn-priv-net ovn-port-net; do
    echo "Delete OVN Network $net network"
    kubectl delete -f $net.yaml
done
popd
