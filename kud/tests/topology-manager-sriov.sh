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

source _common.sh
source _functions.sh

sriov_capable_nodes=$(kubectl get nodes -o json | jq -r '.items[] | select((.status.capacity."intel.com/intel_sriov_nic"!=null) and ((.status.capacity."intel.com/intel_sriov_nic"|tonumber)>=2)) | .metadata.name')
if [ -z "$sriov_capable_nodes" ]; then
    echo "Ethernet adaptor version is not set. Topology manager test case cannot run on this machine"
    exit 0
else
    echo "NIC card specs match. Topology manager option avaiable for this version."
fi

pod_name=pod-topology-manager
csar_id=bd55cccc-bf34-11ea-b3de-0242ac130004

function create_pod_yaml {
    local csar_id=$1
    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << POD > $pod_name.yaml
kind: Pod
apiVersion: v1
metadata:
  name: $pod_name
  annotations:
    k8s.v1.cni.cncf.io/networks: sriov-intel
spec:
  containers:
  - name: $pod_name
    image: docker.io/centos/tools:latest
    command:
    - /sbin/init
    resources:
      limits:
        cpu: "1"
        memory: "500Mi"
        intel.com/intel_sriov_nic: '1'
      requests:
        cpu: "1"
        memory: "500Mi"
        intel.com/intel_sriov_nic: '1'
POD
    popd
}

create_pod_yaml ${csar_id}
kubectl delete pod $pod_name --ignore-not-found=true --now --wait
kubectl create -f ${CSAR_DIR}/${csar_id}/$pod_name.yaml --validate=false

status_phase=""
while [[ $status_phase != "Running" ]]; do
    new_phase=$(kubectl get pods $pod_name | awk 'NR==2{print $3}')
    if [[ $new_phase != $status_phase ]]; then
        echo "$(date +%H:%M:%S) - $pod_name : $new_phase"
        status_phase=$new_phase
    fi
    if [[ $new_phase == "Running" ]]; then
        echo "Pod is up and running.."
    fi
    if [[ $new_phase == "Err"* ]]; then
        exit 1
    fi
done

uid=$(kubectl get pod pod-topology-manager -o jsonpath='{.metadata.uid}')
node_name=$(kubectl get pod $pod_name -o jsonpath='{.spec.nodeName}')
node_ip=$(kubectl get node $node_name -o jsonpath='{.status.addresses[].address}')

apt-get install -y jq
cpu_core=$(ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $node_ip -- cat /var/lib/kubelet/cpu_manager_state | jq -r --arg UID "${uid}" --arg POD_NAME "${pod_name}" '.entries[$UID][$POD_NAME]')
numa_node_number=$(ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $node_ip -- lscpu | grep "NUMA node(s)" | awk -F ':' '{print $2}')
for (( node=0; node<$numa_node_number; node++ )); do
    ranges=$(ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $node_ip -- lscpu | grep "NUMA node"$node | awk -F ':' '{print $2}')
    ranges=(${ranges//,/ })
    for range in ${ranges[@]}; do
        min=$(echo $range | awk -F '-' '{print $1}')
        max=$(echo $range | awk -F '-' '{print $2}')
        if [ $cpu_core -ge $min ] && [ $cpu_core -le $max ]; then
            cpu_numa_node=$node
        fi
    done
done

vf_pci=$(kubectl exec -it $pod_name -- env | grep PCIDEVICE_INTEL_COM_INTEL_SRIOV_NIC | awk -F '=' '{print $2}' | sed 's/\r//g')
vf_numa_node=$(ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null $node_ip -- cat /sys/bus/pci/devices/$vf_pci/numa_node)

echo "The allocated cpu core is:" $cpu_core
echo "The numa node of the allocated cpu core is:" $cpu_numa_node
echo "The PCI address of the allocated vf is:" $vf_pci
echo "The numa node of the allocated vf is:" $vf_numa_node
if [ $cpu_numa_node == $vf_numa_node ]; then
    echo "The allocated cpu core and vf are on the same numa node"
else
    echo "The allocated cpu core and vf are on different numa nodes"
fi

kubectl delete pod $pod_name --now
echo "Test complete."
