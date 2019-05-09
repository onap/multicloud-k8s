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
#set -o xtrace

source _common_test.sh
source _functions.sh

base_url="http://localhost:8081/v1/vnf_instances/"
cloud_region_id="kud"
namespace="default"
csar_id=cb009bfe-bbee-11e8-9766-525400435678

# Setup
install_deps
_checks_args ${csar_id}
cp -R ./edgex/* ${CSAR_DIR}/${csar_id}/

# Test
payload_raw="
{
    \"cloud_region_id\": \"$cloud_region_id\",
    \"namespace\": \"$namespace\",
    \"csar_id\": \"$csar_id\"
}
"
payload=$(echo $payload_raw | tr '\n' ' ')

echo "Creating EdgeX VNF Instance"

vnf_id=$(curl -s -d "$payload" "${base_url}" | jq -r '.vnf_id')

echo "=== Validating Kubernetes ==="
kubectl get --no-headers=true --namespace=${namespace} deployment ${cloud_region_id}-${namespace}-${vnf_id}-edgex-core-command
kubectl get --no-headers=true --namespace=${namespace} service ${cloud_region_id}-${namespace}-${vnf_id}-edgex-core-command
echo "VNF Instance created succesfully with id: $vnf_id"

# TODO: Add heath checks to verify EdgeX services

vnf_id_list=$(curl -s -X GET "${base_url}${cloud_region_id}/${namespace}" | jq -r '.vnf_id_list')
if [[ "$vnf_id_list" != *"${vnf_id}"* ]]; then
    echo $vnf_id_list
    echo "VNF Instance not stored"
    exit 1
fi

vnf_details=$(curl -s -X GET "${base_url}${cloud_region_id}/${namespace}/${vnf_id}")
if [[ -z "$vnf_details" ]]; then
    echo "Cannot retrieved VNF Instance details"
    exit 1
fi
echo "VNF details $vnf_details"

echo "Deleting $vnf_id VNF Instance"
curl -X DELETE "${base_url}${cloud_region_id}/${namespace}/${vnf_id}"
if [[ -n $(curl -s -X GET "${base_url}${cloud_region_id}/${namespace}/${vnf_id}") ]]; then
    echo "VNF Instance not deleted"
    exit 1
fi

# Teardown
