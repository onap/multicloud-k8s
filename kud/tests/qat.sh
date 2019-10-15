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
    echo "True."
    QAT_ENABLED=True
fi

pod_name=pod-case-01
rm -f $HOME/$pod_name.yaml
kubectl delete pod $pod_name --ignore-not-found=true --now --wait

cat << POD > $HOME/$pod_name.yaml
kind: Pod
apiVersion: v1
metadata:
  name: qat_kernel
spec:
  containers:
  - name: qat_kernel
    image: crypto-perf:devel
    imagePullPolicy: IfNotPresent
    command: [ "/bin/bash", "-c", "--" ]
    args: [ "while true; do sleep 300000; done;" ]
    volumeMounts:
    - mountPath: /dev/hugepages
      name: hugepage
    resources:
      requests:
        cpu: "3"
        memory: "1Gi"
        qat.intel.com/cy2_dc2: '1'
        hugepages-2Mi: "1Gi"
      limits:
        cpu: "3"
        memory: "1Gi"
        qat.intel.com/cy2_dc2: '1'
        hugepages-2Mi: "1Gi"
    securityContext:
      capabilities:
        add:
          ["IPC_LOCK"]
  volumes:
  - name: hugepage
    emptyDir:
      medium: HugePages
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
kubectl exec -it $pod_name bash
./dpdk-test-crypto-perf -l 6-7 -w $QAT1 -- --ptest throughput --devtype crypto_qat --optype cipher-only --cipher-algo aes-cbc --cipher-op encrypt --cipher-key-sz 16 --total-ops 10000000 --burst-sz 32 --buffer-sz 64
kubectl delete pod $pod_name --now
echo "Test complete."
