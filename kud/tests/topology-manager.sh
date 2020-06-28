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

#checking for the right hardware version of NIC on the machine
ethernet_adpator_version=$( lspci | grep "Ethernet Controller XL710" | head -n 1 | cut -d " " -f 8 )
if [ -z "$ethernet_adpator_version" ]; then
    echo " Ethernet adapator version is not set. Topology manager test case cannot run on this machine"
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
    k8s.v1.cni.cncf.io/networks: sriov-eno2
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
        intel.com/intel_sriov_700: '1'
      requests:
        cpu: "1"
        memory: "500Mi"
        intel.com/intel_sriov_700: '1'
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

container_id=$(kubectl describe pod $pod_name | grep "Container ID" | awk '{print $3}' )
container_id=${container_id#docker://}
container_id=${container_id:0:12}

apt-get install -y jq
cpu_core=$(cat /var/lib/kubelet/cpu_manager_state | jq -r .| grep ${container_id} | awk -F ':' '{print $2}'| awk -F '"' '{print $2}')
numa_node_number=$(lscpu | grep "NUMA node(s)" | awk -F ':' '{print $2}')
for (( node=0; node<$numa_node_number; node++ )); do
    ranges=$(lscpu | grep "NUMA node"$node | awk -F ':' '{print $2}')
    ranges=(${ranges//,/ })
    for range in ${ranges[@]}; do
        min=$(echo $range | awk -F '-' '{print $1}')
        max=$(echo $range | awk -F '-' '{print $2}')
        if [ $cpu_core -ge $min ] && [ $cpu_core -le $max ]; then
            cpu_numa_node=$node
        fi
    done
done

vf_pci=$(kubectl exec -it $pod_name env | grep PCIDEVICE_INTEL_COM_INTEL_SRIOV_700 | awk -F '=' '{print $2}' | sed 's/\r//g')
vf_numa_node=$(cat /sys/bus/pci/devices/$vf_pci/numa_node)

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
