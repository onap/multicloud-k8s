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

csar_id=4f726e2a-b74a-11e8-ad7c-525400feed2

# Setup
populate_CSAR_containers_vFW $csar_id

pushd ${CSAR_DIR}/${csar_id}
for resource in $unprotected_private_net $protected_private_net $onap_private_net; do
    kubectl apply -f $resource.yaml
done
setup $packetgen_deployment_name $firewall_deployment_name $sink_deployment_name

# Test
popd

# Teardown
teardown $packetgen_deployment_name $firewall_deployment_name $sink_deployment_name
