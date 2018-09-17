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

source _common.sh
source _functions.sh

csar_id=6b54a728-b76a-11e8-a1ba-52540053ccc8

# Setup
popule_CSAR_virtlet $csar_id

pushd ${CSAR_DIR}/${csar_id}

setup $virtlet_deployment_name

# Test
kubectl plugin virt virsh list
deployment_pod=$(kubectl get pods | grep $virtlet_deployment_name | awk '{print $1}')
virsh_image=$(kubectl plugin virt virsh list | grep "virtlet-.*-$deployment_pod")
if [[ -z "$virsh_image" ]]; then
    echo "There is no Virtual Machine running by $deployment_pod pod"
    exit 1
fi
popd

# Teardown
teardown $virtlet_deployment_name
