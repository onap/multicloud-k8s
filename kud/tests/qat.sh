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

qat_device=$( for i in 0434 0435 37c8 6f54 19e2; \
                do lspci -d 8086:$i -m; done |\
                grep -i "Quick*" | head -n 1 | cut -d " " -f 5 )
#Checking if the QAT device is on the node
if [ -z "$qat_device" ]; then
    echo "False. This test case cannot run. Qat device unavailable."
    QAT_ENABLED=False
    exit 0
else
    echo "True. Can run QAT on this device."
    QAT_ENABLED=True
fi

pod_name=pod-case-01
rm -f $HOME/$pod_name.yaml
kubectl delete pod $pod_name --ignore-not-found=true --now --wait
allocated_node_resource=$(kubectl describe node | grep "qat.intel.com" | tail -n1 |awk '{print $(NF)}')
echo "The allocated resource of the node is: " $allocated_node_resource
cat << POD > $HOME/$pod_name.yaml
kind: Pod
apiVersion: v1
metadata:
  name: pod-case-01
spec:
  containers:
  - name: pod-case-01
    image: integratedcloudnative/openssl-qat-engine:devel
    imagePullPolicy: IfNotPresent
    volumeMounts:
            - mountPath: /dev
              name: dev-mount
            - mountPath: /etc/c6xxvf_dev0.conf
              name: dev0
    command: [ "/bin/bash", "-c", "--" ]
    args: [ "while true; do sleep 300000; done;" ]
    resources:
      requests:
        qat.intel.com/cy2_dc2: '1'
      limits:
        qat.intel.com/cy2_dc2: '1'
  volumes:
  - name: dev-mount
    hostPath:
        path: /dev
  - name: dev0
    hostPath:
        path: /etc/c6xxvf_dev0.conf
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
                echo "Pod is up and running.."
            fi
            if [[ $new_phase == "Err"* ]]; then
                exit 1
            fi
        done
    done

allocated_node_resource=$(kubectl describe node | grep "qat.intel.com" | tail -n1 |awk '{print $(NF)}')
echo "The allocated resource of the node is: " $allocated_node_resource
adf_ctl restart
systemctl restart qat_service
kubectl exec -it pod-case-01 -- openssl engine -c -t qat

kubectl delete pod $pod_name --now
echo "Test complete."
