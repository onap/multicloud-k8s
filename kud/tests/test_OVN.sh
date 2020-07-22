
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

# Script name: ./test_OVN.sh
# Purpose: To ascertain whether or not the POST/DELETE/GET API is able to register a null name
# Description, userdata1, and userdata2 have values that I have assigned.

set -o errexit
set -o nounset
set -o pipefail
#set -o xtrace

source _common_test.sh
source _functions.sh
source _common.sh
source _test_functions.sh

if [ ${1:+1} ]; then
    if [ "$1" == "--external" ]; then
        master_ip=$(kubectl cluster-info | grep "Kubernetes master" | \
            awk -F ":" '{print $2}' | awk -F "//" '{print $2}')
        onap_svc_node_port=30498
        base_url="http://$master_ip:$onap_svc_node_port/v1"
    fi
fi

# Setup
source _test_variables_setup.sh
install_deps
populate_CSAR_composite_app_helm "$csar_id"

# Clean up
delete_all

# Register project
create_project

# Register composite-app
create_composite_app

# Create entries for app1&app2 in the database
create_app1
create_app2

# Create entries for app1&app2 in the database
create_main_composite_profile

create_profile_app1
create_profile_app2


create_generic_placement_intent_app1
create_placement_intent_app1

create_placement_intent_app2

create_deployment_intent_group


create_adding_all_intents_to_deployment_intent_group

create_cost_based_controller
create_HPA_controller
create_traffic_controller

print_msg "Adding OVN action contoller"
OVNintent=""
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${OVNintent}",
      "description":"${OVNintentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${OVNHostName}",
      "port": ${OVNPort},
      "type": "action",
      "priority": 2
   }
}
EOF
)"
call_api_negative -d "${payload}" "${base_url}/controllers"

# Test-1
# register null OVNintent
print_msg "Getting the sorted templates for each of the apps.."
call_api_negative -d "" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/instantiate"
if [ $return_status == 405 ] ;then
    print_msg "Test:OVN post with null project name. Expected = 405, Actual = $return_status PASSED"
else
    print_msg "Test:OVN post with null project name. Expected = 405, Actual = $return_status FAILED"
fi

# Test-2
# delete null OVNintent
OVNintent=""
print_msg "Deleting controller ${OVNintent}"
delete_resource_negative "${base_url}/controllers/${OVNintent}"
if [ $return_status == 405 ] ;then
    print_msg "Test:OVN delete with null project name. Expected = 405, Actual = $return_status PASSED"
else
    print_msg "Test:OVN delete with null project name. Expected = 405, Actual = $return_status FAILED"
fi

# Test-3
# get null OVNintent
OVNintent=""
print_msg "Deleting controller ${OVNintent}"
get_resource_negative "${base_url}/controllers/${OVNintent}"
if [ $return_status == 404 ] ;then
    print_msg "Test:OVN get with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:OVN get with null project name. Expected = 404, Actual = $return_status FAILED"
fi
# END: Instantiation
