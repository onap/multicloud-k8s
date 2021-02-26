#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# Copyright © 2021 Samsung Electronics
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# Script respects environment variable SKIP_CNF_TEARDOWN
# If set to "yes", it will preserve CNF for manual handling

set -o errexit
set -o nounset
set -o pipefail
#set -o xtrace

source _common_test.sh
source _functions.sh
source _common.sh

if [ ${1:+1} ]; then
    if [ "$1" == "--external" ]; then
        master_ip=$(kubectl cluster-info | grep "Kubernetes master" | \
            awk -F ":" '{print $2}' | awk -F "//" '{print $2}')
        onap_svc_node_port=30498
        base_url="http://$master_ip:$onap_svc_node_port/v1"
    fi
fi

base_url=${base_url:-"http://localhost:9015/v1"}
kubeconfig_path="$HOME/.kube/config"
csar_id=cc009bfe-bbee-11e8-9766-525400435678
rb_name="vfw"
rb_version="plugin_test"
chart_name="firewall"
profile_name="test_profile"
release_name="test-release"
namespace="plugin-tests-namespace"
cloud_region_id="kud"
cloud_region_owner="localhost"

# Setup
install_deps
populate_CSAR_fw_rbdefinition "$csar_id"

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
    "release-name": "dummy",
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

print_msg "Creating vFW VNF Instance"
payload="$(cat <<EOF
{
    "rb-name": "${rb_name}",
    "rb-version": "${rb_version}",
    "profile-name": "${profile_name}",
    "release-name": "${release_name}",
    "cloud-region": "${cloud_region_id}",
    "labels": {"testCaseName": "plugin_fw.sh"},
    "override-values": {"global.onapPrivateNetworkName": "onap-private-net-test"}
}
EOF
)"
response="$(call_api -d "${payload}" "${base_url}/instance")"
echo "$response"
vnf_id="$(jq -r '.id' <<< "${response}")"

print_msg "[BEGIN] Basic checks for instantiated resource"
print_msg "Check if override value has been applied correctly"
kubectl get network -n "${namespace}" onap-private-net-test
print_msg "Wait for all pods to start"
wait_for_pod -n "${namespace}" -l app=sink
wait_for_pod -n "${namespace}" -l app=firewall
wait_for_pod -n "${namespace}" -l app=packetgen
# TODO: Provide some health check to verify vFW work
print_msg "Not waiting for vFW to fully install as no further checks are implemented in testcase"
#print_msg "Waiting 8minutes for vFW installation"
#sleep 8m
print_msg "[END] Basic checks for instantiated resource"

print_msg "Retrieving VNF status (this will result with long output)"
call_api "${base_url}/instance/${vnf_id}/status"

print_msg "Retrieving VNF details"
response="$(call_api "${base_url}/instance/${vnf_id}")"
echo "$response"
print_msg "Assert additional label has been assigned to rb instance"
test "$(jq -r '.request.labels.testCaseName' <<< "${response}")" == plugin_fw.sh
print_msg "Assert ReleaseName has been correctly overriden"
test "$(jq -r '.request."release-name"' <<< "${response}")" == "${release_name}"

#Teardown
if [ "${SKIP_CNF_TEARDOWN:-}" == "yes" ]; then
    print_msg "Leaving CNF running for further debugging"
    echo "Remember to later issue following DELETE calls to clean environment"
    cat <<EOF
    curl -X DELETE "${base_url}/instance/${vnf_id}"
    curl -X DELETE "${base_url}/rb/definition/${rb_name}/${rb_version}/profile/${profile_name}"
    curl -X DELETE "${base_url}/rb/definition/${rb_name}/${rb_version}"
    curl -X DELETE "${base_url}/connectivity-info/${cloud_region_id}"
EOF
else
    print_msg "Deleting VNF Instance"
    delete_resource "${base_url}/instance/${vnf_id}"

    print_msg "Deleting Profile"
    delete_resource "${base_url}/rb/definition/${rb_name}/${rb_version}/profile/${profile_name}"

    print_msg "Deleting Resource Bundle"
    delete_resource "${base_url}/rb/definition/${rb_name}/${rb_version}"

    print_msg "Deleting ${cloud_region_id} cloud region connection"
    delete_resource "${base_url}/connectivity-info/${cloud_region_id}"
fi
print_msg "Test finished successfully"
