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

csar_id=aa443e7e-c8ba-11e8-8877-525400b164ff

# Setup
install_ovn_deps
if [[ ! -f $HOME/.ssh/id_rsa.pub ]]; then
    echo -e "\n\n\n" | ssh-keygen -t rsa -N ""
fi
populate_CSAR_vms_containers_vFW $csar_id

pushd ${CSAR_DIR}/${csar_id}
for net in $unprotected_private_net $protected_private_net $onap_private_net; do
    cleanup_network $net.yaml
    echo "Create OVN Network $net network"
    init_network $net.yaml
done
for resource in onap-ovn4nfvk8s-network sink-service; do
    kubectl apply -f $resource.yaml
done
setup $packetgen_deployment_name $firewall_deployment_name $sink_deployment_name
#kubectl port-forward deployment/$sink_deployment_name 667:667

# Test
for deployment_name in $packetgen_deployment_name $firewall_deployment_name; do
    pod_name=$(kubectl get pods | grep  $deployment_name | awk '{print $1}')
    vm=$(kubectl plugin virt virsh list | grep ".*$deployment_name"  | awk '{print $2}')
    echo "Pod name: $pod_name Virsh domain: $vm"
    echo "ssh -i ~/.ssh/id_rsa.pub admin@$(kubectl get pods $pod_name -o jsonpath="{.status.podIP}")"
    echo "kubectl attach -it $pod_name"
    echo "=== Virtlet details ===="
    echo "$(kubectl plugin virt virsh dumpxml $vm | grep VIRTLET_)\n"
done

# Teardown
#teardown $packetgen_deployment_name $firewall_deployment_name $sink_deployment_name
#for net in $unprotected_private_net $protected_private_net $onap_private_net; do
#    cleanup_network $net.yaml
#done
popd
