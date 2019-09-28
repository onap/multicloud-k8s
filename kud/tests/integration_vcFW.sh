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

#Default action
TEARDOWN_ACTION=preserve

for arg; do
    case "$arg" in
        --help|-h)
            {
                echo "Usage: $0 [--cleanup]"
                echo "Run integration testcase on KUD deployment"
                echo "deploying vFW demo on hybrid container-VM"
                echo "setup. Script by default preserves environment."
                echo "If you want it to be cleaned, launch script"
                echo "with --cleanup flag"
            } >&2
            exit 0
            ;;
        --cleanup)
            TEARDOWN_ACTION=cleanup
            break
            ;;
        *)
            #not implemented
            break
            ;;
    esac
done

source _common.sh
source _common_test.sh
source _functions.sh

csar_id=aa443e7e-c8ba-11e8-8877-525400b164ff

# Setup
if [[ ! -f $HOME/.ssh/id_rsa.pub ]]; then
    echo -e "\n\n\n" | ssh-keygen -t rsa -N ""
fi
populate_CSAR_vms_containers_vFW $csar_id
pushd ${CSAR_DIR}/${csar_id}

#Clean env
print_msg "Cleanup scenario leftovers"
teardown $packetgen_deployment_name $firewall_deployment_name $sink_deployment_name
for item in $unprotected_private_net $protected_private_net $onap_private_net sink-service sink_configmap; do
    kubectl delete -f $item.yaml --ignore-not-found
done

#Spin up
print_msg "Instantiate vcFW"
for net in $unprotected_private_net $protected_private_net $onap_private_net; do
    echo "Create OVN Network $net network"
    kubectl apply -f $net.yaml
done
for resource in onap-ovn4nfvk8s-network sink-service sink_configmap; do
    kubectl apply -f $resource.yaml
done
setup $packetgen_deployment_name $firewall_deployment_name $sink_deployment_name

# Test
print_msg "Verify integration functionality"
for deployment_name in $packetgen_deployment_name $firewall_deployment_name; do
    pod_name=$(kubectl get pods | grep  $deployment_name | awk '{print $1}')
    vm=$(kubectl virt virsh list | grep ".*$deployment_name"  | awk '{print $2}')
    echo "Pod name: $pod_name Virsh domain: $vm"
    echo "ssh -i ~/.ssh/id_rsa admin@$(kubectl get pods $pod_name -o jsonpath="{.status.podIP}")"
    echo "kubectl attach -it $pod_name"
    echo "=== Virtlet details ===="
    echo "$(kubectl virt virsh dumpxml $vm | grep VIRTLET_)\n"
done

# Teardown
if [ "${TEARDOWN_ACTION}" == "cleanup" ]; then
    print_msg "Teardown integration scenario"
    teardown $packetgen_deployment_name $firewall_deployment_name $sink_deployment_name
    for item in $unprotected_private_net $protected_private_net $onap_private_net sink-service sink_configmap; do
        kubectl delete -f $item.yaml
    done
else
    print_msg "Integration scenario access"
    echo "You can access darkstat service on your pc by (for example) port forwarding sink service"
    echo '`kubectl port-forward svc/sink-service 667`'
    echo "or by direct access to any k8s node under port $(kubectl get svc sink-service -o jsonpath='{.spec.ports[0].nodePort}')"
    print_msg ""
fi
popd
