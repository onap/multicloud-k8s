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
csar_id=8030a02a-7253-11ea-bc55-0242ac130003
rb_name="eaa"
rb_version="plugin_test"
chart_name="eaa"
profile_name="test_profile"
release_name="test-release"
namespace="openness"
cloud_region_id="kud"
cloud_region_owner="localhost"

# Setup
install_deps
populate_CSAR_eaa_rbdefinition "$csar_id"
populate_CSAR_eaa "$csar_id"

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

print_msg "Creating EAA"
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

#Create sample producer and sample consumer
kubectl apply -f "${CSAR_DIR}/${csar_id}/sample_policy.yml"
kubectl create -f "${CSAR_DIR}/${csar_id}/sample_producer.yml"
kubectl create -f "${CSAR_DIR}/${csar_id}/sample_consumer.yml"

print_msg "Waiting for EAA, sample producer and sample consumer"
sleep 60

print_msg "Validating EAA is running"
kubectl get --namespace=${namespace} pods | grep eaa

print_msg "Validating sample producer and sample consumer are running"
kubectl get pods | grep producer
kubectl get pods | grep consumer

print_msg "Validating logs of EAA"
EAA=`kubectl get --namespace=${namespace} pods | grep eaa | awk '{print $1}'`
kubectl logs --namespace=${namespace}  ${EAA}

print_msg "Validating logs of sample producer and sample consumer"
PRODUCER=`kubectl get pods | grep producer | awk '{print $1}'`
CONSUMER=`kubectl get pods | grep consumer | awk '{print $1}'`
kubectl logs ${PRODUCER}
kubectl logs ${CONSUMER}

print_msg "Retrieving VNF details"
call_api "${base_url}/instance/${vnf_id}"

#Teardown
print_msg "Deleting EAA"
delete_resource "${base_url}/instance/${vnf_id}"

print_msg "Deleting sample producer and sample consumer"
kubectl delete deployment producer consumer
kubectl delete networkpolicy eaa-prod-cons-policy 

print_msg "Deleting Profile"
delete_resource "${base_url}/rb/definition/${rb_name}/${rb_version}/profile/${profile_name}"

print_msg "Deleting Resource Bundle"
delete_resource "${base_url}/rb/definition/${rb_name}/${rb_version}"

print_msg "Deleting ${cloud_region_id} cloud region connection"
delete_resource "${base_url}/connectivity-info/${cloud_region_id}"
