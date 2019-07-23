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
install_ovn_deps
populate_CSAR_ovn4nfv $csar_id

pushd ${CSAR_DIR}/${csar_id}
for net in ovn-priv-net ovn-port-net; do
    cleanup_network $net.yaml
    echo "Create OVN Network $net network"
    init_network $net.yaml
done
kubectl apply -f onap-ovn4nfvk8s-network.yaml
setup $ovn4nfv_deployment_name_1
setup $ovn4nfv_deployment_name_2

# Test
deployment_pod_1=$(kubectl get pods | grep  $ovn4nfv_deployment_name_1 | awk '{print $1}')
deployment_pod_2=$(kubectl get pods | grep  $ovn4nfv_deployment_name_2 | awk '{print $1}')
echo "===== $deployment_pod_1 details ====="
kubectl exec -it $deployment_pod_1 -- ip a

ovn_nic=$(kubectl exec -it $deployment_pod_1 -- ip addr show dev net1)
if [[ $ovn_nic != *"net1"* ]]; then
    echo "The $deployment_pod_1 pod doesn't contain the net1 nic"
    exit 1
else
    echo "OVN Interface exists, Ping test starting"
    ovn_nic0_ip_addr=$(kubectl exec -it $deployment_pod_1 -- ip addr show net0 | grep -Po 'inet \K[\d.]+')
    echo "Ping IP Address $ovn_nic0_ip_addr for net0 in pod $deployment_pod_1 from pod $deployment_pod_2"
    ping=$(kubectl exec -it $deployment_pod_2 -- ping -c 1 $ovn_nic0_ip_addr)
    if [[ $ping != *"0% packet loss"* ]]; then
        echo " Ping Test failed"
    else
        echo "Ping Test Passed!"
    fi
fi

# Teardown
teardown $ovn4nfv_deployment_name_1
teardown $ovn4nfv_deployment_name_2
cleanup_network ovn-priv-net.yaml
cleanup_network ovn-port-net.yaml
popd

