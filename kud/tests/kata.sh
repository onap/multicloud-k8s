#!/bin/bash
#  Copyright 2021 Intel Corporation, Inc
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

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
