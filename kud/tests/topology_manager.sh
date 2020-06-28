#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o pipefail

ethernet_adpator_version=$( lspci | grep "Ethernet Controller XL710" | head -n 1 | cut -d " " -f 8 )
if [ -z "$ethernet_adpator_version" ]; then
    echo " Ethernet adapator version is not set. SRIOV test case cannot run on this machine"
    exit 0
fi
#checking for the right hardware version of NIC on the machine
if [ $ethernet_adpator_version == "XL710" ]; then
    echo "NIC card specs match. SRIOV option avaiable for this version."
else
    echo -e "Failed. The version supplied does not match.\nTest cannot be executed."
    exit 0
fi

pod_name=pod-case-01

function create_pod_yaml_with_single_VF {

cat << POD > $HOME/$pod_name-single.yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-case-01
  annotations:
    k8s.v1.cni.cncf.io/networks: sriov-eno2
spec:
  containers:
  - name: test-pod
    image: docker.io/centos/tools:latest
    command:
    - /sbin/init
    resources:
      requests:
        intel.com/intel_sriov_700: '1'
      limits:
        intel.com/intel_sriov_700: '1'
POD
}

function create_pod_yaml_with_multiple_VF {

cat << POD > $HOME/$pod_name-multiple.yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-case-01
  annotations:
    k8s.v1.cni.cncf.io/networks: sriov-eno2, sriov-eno2
spec:
  containers:
  - name: test-pod
    image: docker.io/centos/tools:latest
    command:
    - /sbin/init
    resources:
      requests:
        intel.com/intel_sriov_700: '2'
      limits:
        intel.com/intel_sriov_700: '2'
POD
}
create_pod_yaml_with_single_VF
create_pod_yaml_with_multiple_VF

for podType in ${POD_TYPE:-single multiple}; do

    kubectl delete pod $pod_name --ignore-not-found=true --now --wait
    allocated_node_resource=$(kubectl describe node | grep "intel.com/intel_sriov_700" | tail -n1 |awk '{print $(NF)}')

    echo "The allocated resource of the node is: " $allocated_node_resource

    kubectl create -f $HOME/$pod_name-$podType.yaml --validate=false

        for pod in $pod_name; do
            status_phase=""
            while [[ $status_phase != "Running" ]]; do
                new_phase=$(kubectl get pods $pod | awk 'NR==2{print $3}')
                if [[ $new_phase != $status_phase ]]; then
                    echo "$(date +%H:%M:%S) - $pod-$podType : $new_phase"
                    status_phase=$new_phase
                fi
                if [[ $new_phase == "Running" ]]; then
                    echo "Pod is up and running.."
                fi
                if [[ $new_phase == "Err"* ]]; then
                    exit 1
                fi
            done
        done
    allocated_node_resource=$(kubectl describe node | grep "intel.com/intel_sriov_700" | tail -n1 |awk '{print $(NF)}')

    echo " The current resource allocation after the pod creation is: " $allocated_node_resource
    kubectl delete pod $pod_name --now
    echo "Test complete."

done
