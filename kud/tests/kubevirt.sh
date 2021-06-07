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

source _common.sh
source _functions.sh

csar_id=07f9cfe1-25f6-41fe-b4da-e61a2c94c319

# Setup
populate_CSAR_kubevirt $csar_id

pushd ${CSAR_DIR}/${csar_id}

setup_type "vmi" $kubevirt_vmi_name

# Test
deployment_pod=$(kubectl get pods | grep $kubevirt_vmi_name | awk '{print $1}')
echo "Pod name: $deployment_pod"
echo "ssh testuser@$(kubectl get pods $deployment_pod -o jsonpath="{.status.podIP}")"
echo "kubectl virt console $kubevirt_vmi_name"

tries=18
interval=10
for ((try=1;try<=$tries;try++)); do
    echo "try $try/$tries: Wait for $interval seconds to check for ssh access"
    sleep $interval
    if sshpass -p testuser ssh -o "StrictHostKeyChecking=no" testuser@$(kubectl get pods $deployment_pod -o jsonpath="{.status.podIP}") -- uptime; then
        echo "ssh access check is success"
	break
    fi
done
if (($try > $tries)); then
    echo "ssh access check failed"
    exit 1
fi

popd

# Teardown
teardown_type "vmi" $kubevirt_vmi_name
