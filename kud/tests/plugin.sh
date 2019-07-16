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
source _common_test.sh
source _functions.sh

base_url="http://localhost:9015/v1"
kubeconfig_path="$HOME/.kube/config"
#Will resolve to file $KUBE_CONFIG_DIR/kud
cloud_region_id="kud"
cloud_region_owner="test_owner"
namespace="testns"
csar_id="94e414f6-9ca4-11e8-bb6a-52540067263b"
rb_name="test-rbdef"
rb_version="v1"
profile_name="profile1"
release_name="testrelease"
vnf_customization_uuid="ebe353d2-30b7-11e9-9515-525400277b3d"

# _build_generic_sim() - Creates a generic simulator image in case that doesn't exist
function _build_generic_sim {
    if [[ -n $(docker images -q generic_sim) ]]; then
        return
    fi
    BUILD_ARGS="--no-cache"
    if [ ${HTTP_PROXY:-} ]; then
        BUILD_ARGS+=" --build-arg HTTP_PROXY=${HTTP_PROXY}"
    fi
    if [ ${HTTPS_PROXY:-} ]; then
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
install_deps
destroy_deployment $plugin_deployment_name

#start_aai_service
populate_CSAR_plugin $csar_id
populate_CSAR_rbdefinition $csar_id

# Test
print_msg "Create Resource Bundle Definition Metadata"
payload="
{
    \"rb-name\": \"${rb_name}\",
    \"rb-version\": \"${rb_version}\",
    \"chart-name\": \"vault-consul-dev\",
    \"description\": \"testing resource bundle definition api\",
    \"labels\": {
        \"vnf_customization_uuid\": \"${vnf_customization_uuid}\"
    }
}
"
call_api -d "$payload" "${base_url}/rb/definition"

print_msg "Upload Resource Bundle Definition Content"
call_api --data-binary "@${CSAR_DIR}/${csar_id}/vault-consul-dev-0.0.0.tgz" "${base_url}/rb/definition/$rb_name/$rb_version/content"

print_msg "Listing Resource Bundle Definitions"
rb_list=$(call_api "${base_url}/rb/definition/$rb_name")
if [[ "$rb_list" != *"${rb_name}"* ]]; then
    echo $rb_list
    echo "Resource Bundle Definition not stored"
    exit 1
fi

print_msg "Create Resource Bundle Profile Metadata"
kubeversion=$(kubectl version | grep 'Server Version' | awk -F '"' '{print $6}')
payload="
{
    \"profile-name\": \"${profile_name}\",
    \"rb-name\": \"${rb_name}\",
    \"rb-version\": \"${rb_version}\",
    \"release-name\": \"${release_name}\",
    \"namespace\": \"$namespace\",
    \"kubernetesversion\": \"$kubeversion\",
    \"labels\": {
        \"vnf_customization_uuid\": \"${vnf_customization_uuid}\"
    }
}
"
call_api -d "$payload" "${base_url}/rb/definition/$rb_name/$rb_version/profile"

print_msg "Upload Resource Bundle Profile Content"
call_api --data-binary "@${CSAR_DIR}/${csar_id}/rb_profile.tar.gz" "${base_url}/rb/definition/$rb_name/$rb_version/profile/$profile_name/content"

print_msg "Getting Resource Bundle Profile"
rbp_ret=$(call_api "${base_url}/rb/definition/$rb_name/$rb_version/profile/$profile_name")
if [[ "$rbp_ret" != *"${profile_name}"* ]]; then
    echo $rbp_ret
    echo "Resource Bundle Profile not stored"
    exit 1
fi

print_msg "Setup cloud data"
payload="$(cat <<EOF
{
    "cloud-region": "$cloud_region_id",
    "cloud-owner": "$cloud_region_owner"
}
EOF
)"
call_api -F "metadata=$payload" \
    -F "file=@$kubeconfig_path" \
    "${base_url}/connectivity-info" >/dev/null

print_msg "Instantiate Profile"
payload="
{
    \"cloud-region\": \"$cloud_region_id\",
    \"rb-name\":\"$rb_name\",
    \"rb-version\":\"$rb_version\",
    \"profile-name\":\"$profile_name\"
}
"
inst_id=$(call_api -d "$payload" "${base_url}/instance")
echo "$inst_id"
inst_id=$(jq -r '.id' <<< "$inst_id")

print_msg "Validating Kubernetes"
kubectl get --no-headers=true --namespace=${namespace} deployment ${release_name}-vault-consul-dev
kubectl get --no-headers=true --namespace=${namespace} service override-vault-consul
echo "VNF Instance created succesfully with id: $inst_id"

print_msg "Getting $inst_id VNF Instance information"
call_api "${base_url}/instance/${inst_id}"

# Teardown
print_msg "Deleting $rb_name/$rb_version Resource Bundle Definition"
# TODO: Change the HTTP code for 404 when the resource is not found in the API
delete_resource "${base_url}/rb/definition/$rb_name/$rb_version"

print_msg "Deleting $profile_name Resource Bundle Profile"
# TODO: Change the HTTP code for 404 when the resource is not found in the API
delete_resource "${base_url}/rb/definition/$rb_name/$rb_version/profile/$profile_name"

print_msg "Deleting $inst_id VNF Instance"
delete_resource "${base_url}/instance/${inst_id}"

print_msg "Deleting ${cloud_region_id} cloud region connection"
delete_resource "${base_url}/connectivity-info/${cloud_region_id}"

teardown $plugin_deployment_name
