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
source _test_variables_setup.sh

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

    delete_resource "${base_url}/controllers/${genericPlacementIntent}"
    delete_resource "${base_url}/controllers/${hpaIntent}"
    delete_resource "${base_url}/controllers/${trafficIntent}"
    delete_resource "${base_url}/controllers/${CostBasedIntent}"
    delete_resource "${base_url}/controllers/${OVNintent}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents/${intentToBeAddedinDeploymentIntentGroup}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp2}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp1}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name2}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app2_name}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app1_name}"
    delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}"
    delete_resource "${base_url}/projects/${project_name}"
}

#Create entries for app1 in the database
function create_app1 {
    app1_helm_path="$CSAR_DIR/$csar_id/collectd.tar.gz"
    app1_name="collectd"
    app1_desc="collectd_desc"
    userData1="user1"
    userData2="user2"
    project_name="test_project"
    composite_app_name="test_composite_app_collection"
    composite_app_version="test_composite_app_version"

    print_msg "Making app entry in the database"
    payload="$(cat <<EOF
    {
        "metadata": {
        "name": "${app1_name}",
        "description": "${app1_desc}",
        "userData1": "${userData1}",
        "userData2": "${userData2}"
        }
    }
EOF
    )"

        call_api -F "metadata=$payload" \
        -F "file=@$app1_helm_path" \
            "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps"
}

# BEGIN: Create an entry for app2 in the database
function create_app2 {
    app2_helm_path="$CSAR_DIR/$csar_id/prometheus-operator.tar.gz"
    project_name="test_project"
    composite_app_name="test_composite_app_collection"
    composite_app_version="test_composite_app_version"
    app2_name="prometheus"
    app2_desc="prometheus_desc"
    userData1="user1"
    userData2="user2"

    print_msg "Making app entry in the database"
    payload="$(cat <<EOF
    {
        "metadata": {
        "name": "${app2_name}",
        "description": "${app2_desc}",
        "userData1": "${userData1}",
        "userData2": "${userData2}"
        }
    }
EOF
)"

call_api -F "metadata=$payload" \
        -F "file=@$app2_helm_path" \
        "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps"
}

# BEGIN: Register the main composite-profile
function create_main_composite_profile {
    print_msg "Registering the main composite-profile"
    payload="$(cat <<EOF
    {
        "metadata":{
        "name":"${main_composite_profile_name}",
        "description":"${main_composite_profile_description}",
        "userData1":"${userData1}",
        "userData2":"${userData2}"
        }
    }
EOF
    )"
    call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles"
}

# BEGIN : Adding profile to each of the two apps - app1(collectd) and app2(prometheus)
function create_profile_app1 {

    print_msg "Registering profile with app1(collectd)"
    payload="$(cat <<EOF
    {
        "metadata":{
        "name":"${sub_composite_profile_name1}",
        "description":"${sub_composite_profile_description}",
        "userData1":"${userData1}",
        "userData2":"${userData2}"
        },
        "spec":{
        "app-name":  "${app1_name}"
        }
    }
EOF
    )"

    call_api -F "metadata=$payload" \
        -F "file=@$app1_profile_path" \
        "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles"
}

function create_profile_app2 {
    print_msg "Registering profile with app2(prometheus)"
    payload="$(cat <<EOF
    {
        "metadata":{
        "name":"${sub_composite_profile_name2}",
        "description":"${sub_composite_profile_description}",
        "userData1":"${userData1}",
        "userData2":"${userData2}"
        },
        "spec":{
        "app-name":  "${app2_name}"
        }
    }
EOF
    )"

    call_api -F "metadata=$payload" \
    -F "file=@$app2_profile_path" \
    "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles"
}

# BEGIN: Register GenericPlacementIntents with the database
function create_generic_placement_intent_app1 {
    print_msg "Registering GenericPlacementIntent for app1"
    payload="$(cat <<EOF
    {
        "metadata":{
        "name":"${genericPlacementIntentName}",
        "description":"${genericPlacementIntentDesc}",
        "userData1":"${userData1}",
        "userData2":"${userData2}"
        },
        "spec":{
        "logical-cloud":"${logicalCloud}"
        }
    }
EOF
    )"
    call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents"
}

# BEGIN: Adding placement intent for each app in the composite app.
function create_placement_intent_app1 {
    print_msg "Adding placement intent for app1(collectd)"
    payload="$(cat <<EOF
    {
        "metadata":{
        "name":"${appIntentNameForApp1}",
        "description":"${appIntentForApp1Desc}",
        "userData1":"${userData1}",
        "userData2":"${userData2}"
        },
        "spec":{
        "app-name":"${app1_name}",
        "intent":{
         "allOf":[
            {  "provider-name":"${providerName1}",
               "cluster-name":"${clusterName1}"
            },
            {
               "provider-name":"${providerName2}",
               "cluster-name":"${clusterName2}"
            },
            {
               "anyOf":[
                  {
                     "provider-name":"${providerName1}",
                     "cluster-label-name":"${clusterLabelName1}"
                  },
                  {
                     "provider-name":"${providerName2}",
                     "cluster-label-name":"${clusterLabelName2}"
                  }
               ]
            }
         ]
      }
   }
}
EOF
    )"
    call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents"
}

function create_placement_intent_app2 {
print_msg "Adding placement intent for app2(prometheus)"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${appIntentNameForApp2}",
      "description":"${appIntentForApp2Desc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "app-name":"${app2_name}",
      "intent":{
         "allOf":[
            {
               "provider-name":"${providerName1}",
               "cluster-name":"${clusterName1}"
            },
            {
               "provider-name":"${providerName2}",
               "cluster-name":"${clusterName2}"
            },
            {
               "anyOf":[
                  {
                     "provider-name":"${providerName1}",
                     "cluster-label-name":"${clusterLabelName1}"
                  },
                  {
                     "provider-name":"${providerName2}",
                     "cluster-label-name":"${clusterLabelName2}"
                  }
               ]
            }
         ]
      }
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents"
}

# BEGIN: Registering DeploymentIntentGroup in the database
function create_deployment_intent_group {
print_msg "Registering DeploymentIntentGroup"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${deploymentIntentGroupName}",
      "description":"${deploymentIntentGroupNameDesc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "profile":"${main_composite_profile_name}",
      "version":"${releaseName}",
      "override-values":[
         {
            "app-name":"${app1_name}",
            "values":
               {
                  "collectd_prometheus.service.targetPort":"9104"
               }
         },
         {
            "app-name":"${app2_name}",
            "values":
               {
                  "prometheus.service.nameOfPort":"WebPort9090"
               }
         }
      ]
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups"
}

function create_adding_all_intents_to_deployment_intent_group {
# BEGIN: Adding intents to an intent group
print_msg "Adding all the intents to the deploymentIntent group"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${intentToBeAddedinDeploymentIntentGroup}",
      "description":"${intentToBeAddedinDeploymentIntentGroupDesc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "intent":{
         "${genericPlacementIntent}":"${genericPlacementIntentName}",
         "${hpaIntent}" : "${hpaControllerIntentName}", 
         "${trafficIntent}" : "${trafficControllerIntentName}",
         "${CostBasedIntent}" : "${CostBasedIntentName}",
         "${OVNintent}" : "${OVNintentName}"
      }
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents"
# END: Adding intents to an intent group
}

function create_cost_based_controller {
# BEGIN: Adding controllers
print_msg "Adding CostBased placement contoller"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${CostBasedIntent}",
      "description":"${CostBasedIntentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${CostBasedHostName}",
      "port": ${CostBasedPort},
      "type": "placement",
      "priority": 3
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/controllers"
}

function create_HPA_controller {
print_msg "Adding HPA contoller"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${hpaIntent}",
      "description":"${hpaControllerIntentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${hpaHostName}",
      "port": ${hpaPort},
      "type": "placement",
      "priority": 2
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/controllers"
}

function create_traffic_controller {
print_msg "Adding traffic contoller"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${trafficIntent}",
      "description":"${trafficControllerIntentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${trafficHostName}",
      "port": ${trafficPort},
      "type": "action",
      "priority": 3
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/controllers"
}
