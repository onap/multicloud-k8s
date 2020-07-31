#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# Additional functions to run negative tests
# Aditya Sharoff <aditya.sharoff@intel.com> 07/16/2020

set -o errexit
set -o nounset
set -o pipefail

FUNCTIONS_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

source /etc/environment
source $FUNCTIONS_DIR/_common_test.sh

base_url=${base_url:-"http://localhost:9015/v2"}
csar_id=cb009bfe-bbee-11e8-9766-525400435678

function call_api_negative {
    #Runs curl with passed flags and provides
    #additional error handling and debug information

    #Function outputs server response body
    #and performs validation of http_code

    local status
    local curl_response_file="$(mktemp -p /tmp)"
    local curl_common_flags=(-s -w "%{http_code}" -o "${curl_response_file}")
    local command=(curl "${curl_common_flags[@]}" "$@")

    echo "[INFO] Running '${command[@]}'" >&2
    if ! status="$("${command[@]}")"; then
        echo "[ERROR] Internal curl error! '$status'" >&2
        cat "${curl_response_file}"
        rm "${curl_response_file}"
        return 2
    else
        echo "[INFO] Server replied with status: ${status}" >&2
        cat "${curl_response_file}"
        rm "${curl_response_file}"
	return_status=$status
    fi
}


function delete_resource_negative {
    #Issues DELETE http call to provided endpoint
    #and further validates by following GET request

    call_api_negative -X DELETE "$1"
    #! call_api -X GET "$1" >/dev/null
}

function get_resource_negative {
    #! call_api_negative -X GET "$1" >/dev/null
    ! call_api_negative -X GET "$1" 
    echo $return_status
}


# Create a test rpoject 
# EOF must start as first chracter in a line for cat to identify the end
function create_project {
	project_name="test_project"
	project_description="test_project_description"
	userData1="user1"
	userData2="user2"
	print_msg "Registering project"
	payload="$(cat <<EOF
	{
	  "metadata": {
	    "name": "${project_name}",
	    "description": "${project_description}",
	    "userData1": "${userData1}",
	    "userData2": "${userData2}"
	   }
	}
EOF
	)"
call_api -d "${payload}" "${base_url}/projects"
}

function create_composite_app {

	project_name="test_project"
	composite_app_name="test_composite_app_collection"
	composite_app_description="test_project_description"
	composite_app_version="test_composite_app_version"
	userData1="user1"
	userData2="user2"

	print_msg "Registering composite-app"
	payload="$(cat <<EOF
	{
	  "metadata": {
	    "name": "${composite_app_name}",
	    "description": "${composite_app_description}",
	    "userData1": "${userData1}",
	    "userData2": "${userData2}"
	   },
	   "spec":{
	      "version":"${composite_app_version}"
	   }
	}
EOF
	)"
	call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps"
}

function delete_all {

	project_name="test_project"
	project_description="test_project_description"
	userData1="user1"
	userData2="user2"

	composite_app_name="test_composite_app_collection"
	composite_app_description="test_project_description"
	composite_app_version="test_composite_app_version"
	app1_helm_path="$CSAR_DIR/$csar_id/collectd.tar.gz"
	app2_helm_path="$CSAR_DIR/$csar_id/prometheus.tar.gz"
	app1_profile_path="$CSAR_DIR/$csar_id/collectd_profile.tar.gz"
	app2_profile_path="$CSAR_DIR/$csar_id/prometheus_profile.tar.gz"

	app1_name="collectd"
	app2_name="prometheus"
	app1_desc="collectd_desc"
	app2_desc="prometheus_desc"

	main_composite_profile_name="main_composite_profile"
	sub_composite_profile_name1="test_composite_profile1"
	sub_composite_profile_name2="test_composite_profile2"
	main_composite_profile_description="main_composite_profile_description"
	sub_composite_profile_description="sub_composite_profile_description"

	genericPlacementIntentName="test_gen_placement_intent1"
	genericPlacementIntentDesc="test_gen_placement_intent_desc"
	logicalCloud="logical_cloud_name"

	appIntentNameForApp1="appIntentForApp1"
	appIntentForApp1Desc="AppIntentForApp1Desc"
	appIntentNameForApp2="appIntentForApp2"
	appIntentForApp2Desc="AppIntentForApp2Desc"
	providerName1="cluster_provider1"
	providerName2="cluster_provider2"
	clusterName1="clusterName1"
	clusterName2="clusterName2"
	clusterLabelName1="clusterLabel1"
	clusterLabelName2="clusterLabel2"

	deploymentIntentGroupName="test_deployment_intent_group"
	deploymentIntentGroupNameDesc="test_deployment_intent_group_desc"
	releaseName="test"
	intentToBeAddedinDeploymentIntentGroup="name_of_intent_to_be_added_in_deployment_group"
	intentToBeAddedinDeploymentIntentGroupDesc="desc_of_intent_to_be_added_in_deployment_group"
	hpaIntentName="hpaIntentName"
	trafficIntentName="trafficIntentName"

	chart_name="edgex"
	profile_name="test_profile"
	release_name="test-release"
	namespace="plugin-tests-namespace"
	cloud_region_id="kud"
	cloud_region_owner="localhost"


	# Controllers
	genericPlacementIntent="genericPlacementIntent"
	OVNintent="OVNintent"
	OVNintentName="OVNintentName"
	OVNHostName="OVNHostName"
	OVNPort="9027"
	CostBasedIntent="costBasedIntent"
	CostBasedIntentName="CostBasedIntentName"
	CostBasedHostName="OVNHostName"
	CostBasedPort="9028"
	hpaIntent="hpaIntent"
	trafficIntent="trafficIntent"
	gpcHostName="gpcHostName"
	gpcPort="9029"
	hpaControllerIntentName="hpaControllerIntentName"
	hpaHostName="hpaHostName"
	hpaPort="9030"
	trafficControllerIntentName="trafficControllerIntentName"
	trafficHostName="trafficHostName"
	trafficPort="9031"

	#print_msg "Deleting controller ${genericPlacementIntent}"
	delete_resource "${base_url}/controllers/${genericPlacementIntent}"

	#print_msg "Deleting controller ${hpaIntent}"
	delete_resource "${base_url}/controllers/${hpaIntent}"

	#print_msg "Deleting controller ${trafficIntent}"
	delete_resource "${base_url}/controllers/${trafficIntent}"

	#print_msg "Deleting controller ${CostBasedIntent}"
	delete_resource "${base_url}/controllers/${CostBasedIntent}"

	#print_msg "Deleting controller ${OVNintent}"
	delete_resource "${base_url}/controllers/${OVNintent}"

	#print_msg "Deleting intentToBeAddedinDeploymentIntentGroup"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents/${intentToBeAddedinDeploymentIntentGroup}"

	#print_msg "Deleting ${deploymentIntentGroupName}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}"

	#print_msg "Deleting ${appIntentNameForApp2}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp2}"

	#print_msg "Deleting ${appIntentNameForApp1}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp1}"

	#print_msg "Deleting ${genericPlacementIntentName}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}"

	#print_msg "Deleting ${sub_composite_profile_name2}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name2}"

	#print_msg "Deleting ${sub_composite_profile_name1}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"

	#print_msg "Deleting ${main_composite_profile_name}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}"

	#print_msg "Deleting ${app2_name}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app2_name}"

	#print_msg "Deleting ${app1_name}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app1_name}"

	#print_msg "Deleting ${composite_app_name}/${composite_app_version}"
	delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}"

	#print_msg "Deleting ${project_name}"
	delete_resource "${base_url}/projects/${project_name}"
}
