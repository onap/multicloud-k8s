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

ethernet_adpator_version=$( lspci | grep "Ethernet Controller X710" | head -n 1 | cut -d " " -f 8 )

#checking for the right hardware version of NIC on the machine

if [ $ethernet_adpator_version == "X710" ]; then
    echo "NIC card specs match. SRIOV option avaiable for this version."
else
    echo "Failed. The version supplied does not match.\'\'n Test cannot be exec."
fi

kubectl describe node | grep  "intel.com/intel_sriov_700"

rm -f $HOME/*.yaml

pod_name=pod-case-01

kubectl delete pod $pod_name --ignore-not-found=true --now

    while kubectl get pod $pod_name &>/dev/null; do
        sleep 5
    done
resource=$(kubectl describe node | grep  "intel.com/intel_sriov_700" )

cat << POD > $HOME/$pod_name.yaml
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
kubectl create -f $HOME/$pod_name.yaml --validate=false
    for pod in $pod_name; do
        status_phase=""
        while [[ $status_phase != "Running" ]]; do
            new_phase=$(kubectl get pods $pod | awk 'NR==2{print $3}')
            if [[ $new_phase != $status_phase ]]; then
                echo "$(date +%H:%M:%S) - $pod : $new_phase"
                status_phase=$new_phase
            fi
            if [[ $new_phase == "Running" ]]; then
                echo " Test is complete.."
            fi
            if [[ $new_phase == "Err"* ]]; then
                exit 1
            fi
        done
    done
kubectl describe node | grep  "intel.com/intel_sriov_700" | tail -n1
