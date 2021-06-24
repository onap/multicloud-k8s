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

sriov_capable_nodes=$(kubectl get nodes -o json | jq -r '.items[] | select((.status.capacity."intel.com/intel_sriov_700"!=null) and ((.status.capacity."intel.com/intel_sriov_700"|tonumber)>=2)) | .metadata.name')
if [ -z "$sriov_capable_nodes" ]; then
    echo "SRIOV test case cannot run on the cluster."
    exit 0
else
    echo "SRIOV option avaiable in the cluster."
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
