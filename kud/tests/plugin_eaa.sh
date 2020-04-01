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
csar_eaa_id=8030a02a-7253-11ea-bc55-0242ac130003
csar_sample_app_id=150da0b3-aa8c-481e-b661-2620b810765e
rb_eaa_name="eaa"
rb_sample_app_name="sample_app"
rb_version="plugin_test"
chart_eaa_name="eaa"
chart_sample_app_name="sample-app"
profile_eaa_name="test_eaa_profile"
profile_sample_app_name="test_sample_app_profile"
release_name="test-release"
namespace_eaa="openness"
namespace_sample_app="default"
cloud_region_id="kud"
cloud_region_owner="localhost"

# Setup
install_deps
populate_CSAR_eaa_rbdefinition "$csar_eaa_id"

print_msg "Registering resource bundle for EAA"
payload="$(cat <<EOF
{
    "rb-name": "${rb_eaa_name}",
    "rb-version": "${rb_version}",
    "chart-name": "${chart_eaa_name}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/rb/definition"

print_msg "Uploading resource bundle content for EAA"
call_api --data-binary "@${CSAR_DIR}/${csar_eaa_id}/rb_definition.tar.gz" \
         "${base_url}/rb/definition/${rb_eaa_name}/${rb_version}/content"

print_msg "Registering rb's profile for EAA"
payload="$(cat <<EOF
{
    "rb-name": "${rb_eaa_name}",
    "rb-version": "${rb_version}",
    "profile-name": "${profile_eaa_name}",
    "release-name": "${release_name}",
    "namespace": "${namespace_eaa}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/rb/definition/${rb_eaa_name}/${rb_version}/profile"

print_msg "Uploading profile data for EAA"
call_api --data-binary "@${CSAR_DIR}/${csar_eaa_id}/rb_profile.tar.gz" \
         "${base_url}/rb/definition/${rb_eaa_name}/${rb_version}/profile/${profile_eaa_name}/content"

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

print_msg "Creating EAA"
payload="$(cat <<EOF
{
    "rb-name": "${rb_eaa_name}",
    "rb-version": "${rb_version}",
    "profile-name": "${profile_eaa_name}",
    "cloud-region": "${cloud_region_id}"
}
EOF
)"
response="$(call_api -d "${payload}" "${base_url}/instance")"
echo "$response"
vnf_eaa_id="$(jq -r '.id' <<< "${response}")"

wait_for_deployment eaa 1

#Create sample producer and sample consumer
populate_CSAR_eaa_sample_app_rbdefinition "$csar_sample_app_id"

print_msg "Registering resource bundle for Sample App"
payload="$(cat <<EOF
{
    "rb-name": "${rb_sample_app_name}",
    "rb-version": "${rb_version}",
    "chart-name": "${chart_sample_app_name}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/rb/definition"

print_msg "Uploading resource bundle content for Sample App"
call_api --data-binary "@${CSAR_DIR}/${csar_sample_app_id}/rb_definition.tar.gz" \
         "${base_url}/rb/definition/${rb_sample_app_name}/${rb_version}/content"

print_msg "Registering rb's profile for Sample App"
payload="$(cat <<EOF
{
    "rb-name": "${rb_sample_app_name}",
    "rb-version": "${rb_version}",
    "profile-name": "${profile_sample_app_name}",
    "release-name": "${release_name}",
    "namespace": "${namespace_sample_app}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/rb/definition/${rb_sample_app_name}/${rb_version}/profile"

print_msg "Uploading profile data for Sample App"
call_api --data-binary "@${CSAR_DIR}/${csar_sample_app_id}/rb_profile.tar.gz" \
         "${base_url}/rb/definition/${rb_sample_app_name}/${rb_version}/profile/${profile_sample_app_name}/content"

print_msg "Creating Sample Apps: producer and consumer"
payload="$(cat <<EOF
{
    "rb-name": "${rb_sample_app_name}",
    "rb-version": "${rb_version}",
    "profile-name": "${profile_sample_app_name}",
    "cloud-region": "${cloud_region_id}"
}
EOF
)"
response="$(call_api -d "${payload}" "${base_url}/instance")"
echo "$response"
vnf_sample_app_id="$(jq -r '.id' <<< "${response}")"

wait_for_deployment producer 1
wait_for_deployment consumer 1

print_msg "Validating EAA is running"
kubectl get --namespace=${namespace_eaa} pods | grep eaa

print_msg "Validating sample producer and sample consumer are running"
kubectl get --namespace=${namespace_sample_app}  pods | grep producer
kubectl get --namespace=${namespace_sample_app} pods | grep consumer

print_msg "Validating logs of EAA"
EAA=`kubectl get --namespace=${namespace_eaa} pods | grep eaa | awk '{print $1}'`
kubectl logs --namespace=${namespace_eaa}  ${EAA}

print_msg "Validating logs of sample producer and sample consumer"
# sleep 5 seconds to let producer and consumer generate some logs
sleep 5
PRODUCER=`kubectl get --namespace=${namespace_sample_app} pods | grep producer | awk '{print $1}'`
CONSUMER=`kubectl get --namespace=${namespace_sample_app} pods | grep consumer | awk '{print $1}'`
kubectl logs --namespace=${namespace_sample_app} ${PRODUCER}
kubectl logs --namespace=${namespace_sample_app} ${CONSUMER}

print_msg "Retrieving EAA details"
call_api "${base_url}/instance/${vnf_eaa_id}"

print_msg "Retrieving Sample App details"
call_api "${base_url}/instance/${vnf_sample_app_id}"

#Teardown
print_msg "Deleting sample apps: producer and consumer"
delete_resource "${base_url}/instance/${vnf_sample_app_id}"

print_msg "Deleting Profile for sample app"
delete_resource "${base_url}/rb/definition/${rb_sample_app_name}/${rb_version}/profile/${profile_sample_app_name}"

print_msg "Deleting Resource Bundle for sample app"
delete_resource "${base_url}/rb/definition/${rb_sample_app_name}/${rb_version}"

print_msg "Deleting EAA"
delete_resource "${base_url}/instance/${vnf_eaa_id}"

print_msg "Deleting Profile for EAA"
delete_resource "${base_url}/rb/definition/${rb_eaa_name}/${rb_version}/profile/${profile_eaa_name}"

print_msg "Deleting Resource Bundle for EAA"
delete_resource "${base_url}/rb/definition/${rb_eaa_name}/${rb_version}"

print_msg "Deleting ${cloud_region_id} cloud region connection"
delete_resource "${base_url}/connectivity-info/${cloud_region_id}"
