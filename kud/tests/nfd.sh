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

source _common_test.sh

rm -f $HOME/*.yaml
pod_name=nfd-pod

install_deps

function create_pod_yaml_with_affinity {

cat << POD > $HOME/$pod_name-affinity.yaml
apiVersion: v1
kind: Pod
metadata:
  name: $pod_name
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: "feature.node.kubernetes.io/kernel-version.major"
            operator: Gt
            values:
            - '3'
        - matchExpressions:
          - key: "feature.node.kubernetes.io/kernel-version.major"
            operator: Lt
            values:
            - '20'
        - matchExpressions:
          - key: "feature.node.kubernetes.io/kernel-version.major"
            operator: In
            values:
            - '3'
            - '4'
            - '5'
        - matchExpressions:
          - key: "feature.node.kubernetes.io/kernel-version.major"
            operator: NotIn
            values:
            - '1'
        - matchExpressions:
          - key: "feature.node.kubernetes.io/kernel-version.major"
            operator: Exists
        - matchExpressions:
          - key: "feature.node.kubernetes.io/label_does_not_exist"
            operator: DoesNotExist
  containers:
  - name: with-node-affinity
    image: gcr.io/google_containers/pause:2.0
POD
}

function create_pod_yaml_with_nodeSelector {

cat << POD > $HOME/$pod_name-nodeSelector.yaml
apiVersion: v1
kind: Pod
metadata:
  name: $pod_name
spec:
  nodeSelector:
    feature.node.kubernetes.io/kernel-version.major: '4'
  containers:
  - name: with-node-affinity
    image: gcr.io/google_containers/pause:2.0
POD

}

if $(kubectl version &>/dev/null); then
    labels=$(kubectl get nodes -o json | jq .items[].metadata.labels)

    echo $labels
    if [[ $labels != *"kubernetes.io"* ]]; then
        exit 1
    fi

    create_pod_yaml_with_affinity
    create_pod_yaml_with_nodeSelector

    for podType in ${POD_TYPE:-nodeSelector affinity}; do

        kubectl delete pod $pod_name --ignore-not-found=true --now
        while kubectl get pod $pod_name &>/dev/null; do
            sleep 5
        done

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
                    echo " Test is complete.."
                fi
                if [[ $new_phase == "Err"* ]]; then
                    exit 1
                fi
            done
        done
        kubectl delete pod $pod_name
        while kubectl get pod $pod_name &>/dev/null; do
            sleep 5
        done

    done
fi
