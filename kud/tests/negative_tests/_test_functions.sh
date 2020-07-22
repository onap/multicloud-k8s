# /*
#  * Copyright 2020 Intel Corporation, Inc
#  *
#  * Licensed under the Apache License, Version 2.0 (the "License");
#  * you may not use this file except in compliance with the License.
#  * You may obtain a copy of the License at
#  *
#  *     http://www.apache.org/licenses/LICENSE-2.0
#  *
#  * Unless required by applicable law or agreed to in writing, software
#  * distributed under the License is distributed on an "AS IS" BASIS,
#  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  * See the License for the specific language governing permissions and
#  * limitations under the License.
#  */

# Additional functions to run negative tests

set -o errexit
set -o nounset
set -o pipefail

FUNCTIONS_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
my_directory=$(dirname $PWD)

source /etc/environment
source ${my_directory}/negative_tests/_test_variables_setup.sh
source ${my_directory}/_common_test.sh
source ${my_directory}/_functions.sh
source ${my_directory}/_common.sh

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

    delete_resource_negative "${base_url}/controllers/${genericPlacementIntent}"
    delete_resource_negative "${base_url}/controllers/${hpaIntent}"
    delete_resource_negative "${base_url}/controllers/${trafficIntent}"
    delete_resource_negative "${base_url}/controllers/${CostBasedIntent}"
    delete_resource_negative "${base_url}/controllers/${OVNintent}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents/${intentToBeAddedinDeploymentIntentGroup}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp2}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp1}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name2}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app2_name}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app1_name}"
    delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}"
    delete_resource_negative "${base_url}/projects/${project_name}"
}

function create_app {
    app_helm_path="$CSAR_DIR/$csar_id/$1"
    app_name=$2
    app_desc=$3
    userData1="user1"
    userData2="user2"
    project_name="test_project"
    composite_app_name="test_composite_app_collection"
    composite_app_version="test_composite_app_version"

    print_msg "Making app entry in the database"
    payload="$(cat <<EOF
    {
        "metadata": {
        "name": "${app_name}",
        "description": "${app_desc}",
        "userData1": "${userData1}",
        "userData2": "${userData2}"
        }
    }
EOF
    )"

        call_api -F "metadata=$payload" \
        -F "file=@$app_helm_path" \
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

function create_profile_app {

    sub_composite_profile_name=$1
    app_name=$2
    app_profile_path="$CSAR_DIR/$csar_id/$3"

    print_msg "Registering profile with app1(collectd)"
    payload="$(cat <<EOF
    {
        "metadata":{
        "name":"${sub_composite_profile_name}",
        "description":"${sub_composite_profile_description}",
        "userData1":"${userData1}",
        "userData2":"${userData2}"
        },
        "spec":{
        "app-name":  "${app_name}"
        }
    }
EOF
    )"

    call_api -F "metadata=$payload" \
        -F "file=@$app_profile_path" \
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
function create_placement_intent_app {
    appIntentNameForApp=$1
    appIntentForAppDesc=$2
    app_name=$3

    print_msg "Adding placement intent for app"
    payload="$(cat <<EOF
    {
        "metadata":{
        "name":"${appIntentNameForApp}",
        "description":"${appIntentForAppDesc}",
        "userData1":"${userData1}",
        "userData2":"${userData2}"
        },
        "spec":{
        "app-name":"${app_name}",
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

