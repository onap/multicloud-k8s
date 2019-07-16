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
source _common.sh

base_url="http://localhost:9015/v1"
kubeconfig_path="$HOME/.kube/config"
csar_id=cb009bfe-bbee-11e8-9766-525400435678
rb_name="edgex"
rb_version="plugin_test"
chart_name="edgex"
profile_name="test_profile"
release_name="test-release"
namespace="plugin-tests-namespace"
cloud_region_id="kud"
cloud_region_owner="localhost"

# Setup
install_deps
populate_CSAR_edgex_rbdefinition "$csar_id"

print_msg "Registering resource bundle"
payload="$(cat <<EOF
{
    "rb-name": "${rb_name}",
    "rb-version": "${rb_version}",
    "chart-name": "${chart_name}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/rb/definition"

print_msg "Uploading resource bundle content"
call_api --data-binary "@${CSAR_DIR}/${csar_id}/rb_definition.tar.gz" \
         "${base_url}/rb/definition/${rb_name}/${rb_version}/content"

print_msg "Registering rb's profile"
payload="$(cat <<EOF
{
    "rb-name": "${rb_name}",
    "rb-version": "${rb_version}",
    "profile-name": "${profile_name}",
    "release-name": "${release_name}",
    "namespace": "${namespace}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/rb/definition/${rb_name}/${rb_version}/profile"

print_msg "Uploading profile data"
call_api --data-binary "@${CSAR_DIR}/${csar_id}/rb_profile.tar.gz" \
         "${base_url}/rb/definition/${rb_name}/${rb_version}/profile/${profile_name}/content"

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
         "${base_url}/connectivity-info" >/dev/null #massive output

print_msg "Creating EdgeX VNF Instance"
payload="$(cat <<EOF
{
    "rb-name": "${rb_name}",
    "rb-version": "${rb_version}",
    "profile-name": "${profile_name}",
    "cloud-region": "${cloud_region_id}"
}
EOF
)"
response="$(call_api -d "${payload}" "${base_url}/instance")"
echo "$response"
vnf_id="$(jq -r '.id' <<< "${response}")"

print_msg "Validating Kubernetes"
kubectl get --no-headers=true --namespace=${namespace} deployment edgex-core-command
kubectl get --no-headers=true --namespace=${namespace} service edgex-core-command
# TODO: Add health checks to verify EdgeX services

print_msg "Retrieving VNF details"
call_api "${base_url}/instance/${vnf_id}"

#Teardown
print_msg "Deleting VNF Instance"
delete_resource "${base_url}/instance/${vnf_id}"

print_msg "Deleting Profile"
delete_resource "${base_url}/rb/definition/${rb_name}/${rb_version}/profile/${profile_name}"

print_msg "Deleting Resource Bundle"
delete_resource "${base_url}/rb/definition/${rb_name}/${rb_version}"

print_msg "Deleting ${cloud_region_id} cloud region connection"
delete_resource "${base_url}/connectivity-info/${cloud_region_id}"
