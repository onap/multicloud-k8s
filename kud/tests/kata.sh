#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2021
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

#source _common_test.sh
#source _common.sh
#source _functions.sh

kata_pods="kata-qemu kata-clh"

function wait_for_pod {
    status_phase=""
    while [[ "$status_phase" != "Running" ]]; do
        new_phase="$(kubectl get pods -o wide | grep ^$1 | awk '{print $3}')"
        if [[ "$new_phase" != "$status_phase" ]]; then
            status_phase="$new_phase"
        fi
        if [[ "$new_phase" == "Err"* ]]; then
            exit 1
        fi
        sleep 2
    done
}

for pod in ${kata_pods};do
    echo "Deploying ${pod} pod"
    kubectl apply -f ${pod}.yml
    wait_for_pod ${pod}
    echo "Pod ${pod} deployed successfully"
    kubectl delete -f ${pod}.yml
done
