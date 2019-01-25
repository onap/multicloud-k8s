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

source _common.sh
source _functions.sh

base_url="http://localhost:8081"
cloud_region_id="krd"
namespace="default"
csar_id="94e414f6-9ca4-11e8-bb6a-52540067263b"
rbd_csar_id="7eb09e38-4363-9942-1234-3beb2e95fd85"
definition_id="9d117af8-30b8-11e9-af94-525400277b3d"
profile_id="ebe353d2-30b7-11e9-9515-525400277b3d"

# _build_generic_sim() - Creates a generic simulator image in case that doesn't exist
function _build_generic_sim {
    if [[ -n $(docker images -q generic_sim) ]]; then
        return
    fi
    BUILD_ARGS="--no-cache"
    if [ $HTTP_PROXY ]; then
        BUILD_ARGS+=" --build-arg HTTP_PROXY=${HTTP_PROXY}"
    fi
    if [ $HTTPS_PROXY ]; then
        BUILD_ARGS+=" --build-arg HTTPS_PROXY=${HTTPS_PROXY}"
    fi

    pushd generic_simulator
    echo "Building generic simulator image..."
    docker build ${BUILD_ARGS} -t generic_sim:latest .
    popd
}

# start_aai_service() - Starts a simulator for AAI service
function start_aai_service {
    _build_generic_sim
    if [[ $(docker ps -q --all --filter "name=aai") ]]; then
        docker rm aai -f
    fi
    echo "Start AAI simulator.."
    docker run --name aai -v $(mktemp):/tmp/generic_sim/ -v $(pwd)/generic_simulator/aai/:/etc/generic_sim/ -p 8443:8080 -d generic_sim
}

# Setup
destroy_deployment $plugin_deployment_name

#start_aai_service
populate_CSAR_plugin $csar_id

# Teardown
teardown $plugin_deployment_name

# Setup
populate_CSAR_rbdefinition $rbd_csar_id

# Test
print_msg "Create Resource Bundle Definition Metadata"
payload_raw="
{
    \"name\": \"test-rbdef\",
    \"chart-name\": \"vault-consul-dev\",
    \"description\": \"testing resource bundle definition api\",
    \"uuid\": \"$definition_id\",
    \"service-type\": \"firewall\"
}
"
payload=$(echo $payload_raw | tr '\n' ' ')
rbd_id=$(curl -s -d "$payload" -X POST "${base_url}/v1/rb/definition" | jq -r '.uuid')

print_msg "Upload Resource Bundle Definition Content"
curl -s --data-binary @${CSAR_DIR}/${rbd_csar_id}/${rbd_content_tarball}.gz -X POST "${base_url}/v1/rb/definition/$rbd_id/content"

print_msg "Listing Resource Bundle Definitions"
rbd_id_list=$(curl -s -X GET "${base_url}/v1/rb/definition")
if [[ "$rbd_id_list" != *"${rbd_id}"* ]]; then
    echo $rbd_id_list
    echo "Resource Bundle Definition not stored"
    exit 1
fi

print_msg "Create Resource Bundle Profile Metadata"
kubeversion=$(kubectl version | grep 'Server Version' | awk -F '"' '{print $6}')
payload_raw="
{
    \"name\": \"test-rbprofile\",
    \"namespace\": \"$namespace\",
    \"rbdid\": \"$definition_id\",
    \"uuid\": \"$profile_id\",
    \"kubernetesversion\": \"$kubeversion\"
}
"
payload=$(echo $payload_raw | tr '\n' ' ')
rbp_id=$(curl -s -d "$payload" -X POST "${base_url}/v1/rb/profile" | jq -r '.uuid')

print_msg "Upload Resource Bundle Profile Content"
curl -s --data-binary @${CSAR_DIR}/${rbd_csar_id}/${rbp_content_tarball}.gz -X POST "${base_url}/v1/rb/profile/$rbp_id/content"

print_msg "Listing Resource Bundle Profiles"
rbp_id_list=$(curl -s -X GET "${base_url}/v1/rb/profile")
if [[ "$rbp_id_list" != *"${rbp_id}"* ]]; then
    echo $rbd_id_list
    echo "Resource Bundle Profile not stored"
    exit 1
fi

print_msg "Instantiate Profile"
payload_raw="
{
    \"cloud_region_id\": \"$cloud_region_id\",
    \"rb_profile_id\": \"$profile_id\",
    \"csar_id\": \"$rbd_csar_id\"
}
"
payload=$(echo $payload_raw | tr '\n' ' ')
vnf_id=$(curl -s -d "$payload" "${base_url}/v1/vnf_instances/" | jq -r '.vnf_id')

print_msg "Deleting $rbd_id Resource Bundle Definition"
curl -X DELETE "${base_url}/v1/rb/definition/$rbd_id"
if [[ 500 -ne $(curl -o /dev/null -w %{http_code} -s -X GET "${base_url}/v1/rb/definition/$rbd_id") ]]; then
    echo "Resource Bundle Definition not deleted"
# TODO: Change the HTTP code for 404 when the resource is not found in the API
    exit 1
fi

print_msg "Deleting $vnf_id VNF Instance"
curl -X DELETE "${base_url}/v1/vnf_instances/${cloud_region_id}/${namespace}/${vnf_id}"
if [[ 404 -ne $(curl -o /dev/null -w %{http_code} -s -X GET "${base_url}${cloud_region_id}/${namespace}/${vnf_id}") ]]; then
    echo "VNF Instance not deleted"
    exit 1
fi
