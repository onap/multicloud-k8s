
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

# Script name: ./test_traffic_controller.sh
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
source _test_variables_setup.sh

if [ ${1:+1} ]; then
    if [ "$1" == "--external" ]; then
        master_ip=$(kubectl cluster-info | grep "Kubernetes master" | \
            awk -F ":" '{print $2}' | awk -F "//" '{print $2}')
        onap_svc_node_port=30498
        base_url="http://$master_ip:$onap_svc_node_port/v1"
    fi
fi

# Setup
install_deps
populate_CSAR_composite_app_helm "$csar_id"

# Cleanup
delete_all

# Register project
create_project

create_composite_app

# Create entries for app1&app2 in the database
create_app1
create_app2

# Register the main composite-profile
create_main_composite_profile


# Adding profile to each of the two apps - app1(collectd) and app2(prometheus)
create_profile_app1
create_profile_app2

# Register GenericPlacementIntents with the database
create_generic_placement_intent_app1

# Adding placement intent for each app in the composite app.
create_placement_intent_app1
create_placement_intent_app2

# Registering DeploymentIntentGroup in the database
create_deployment_intent_group

# Adding intents to an intent group
create_adding_all_intents_to_deployment_intent_group

# Adding controllers
create_cost_based_controller
create_HPA_controller

trafficIntent=""
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

# Test-1
# register null trafficIntent
call_api_negative -d "${payload}" "${base_url}/controllers"
if [ $return_status == 405 ] ;then
    print_msg "Test:traffic controller post with invalid project name. Expected = 405, Actual = $return_status PASSED"
else
    print_msg "Test:traffic controller post with invalid project name. Expected = 405, Actual = $return_status FAILED"
fi

# Test-2
# delete null trafficIntent
trafficIntent=""
print_msg "Deleting controller ${trafficIntent}"
delete_resource_negative "${base_url}/controllers/${trafficIntent}"
if [ $return_status == 400 ] ;then
    print_msg "Test:traffic controller delete with invalid project name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:traffic controller delete with invalid project name. Expected = 400, Actual = $return_status FAILED"
fi

# Test-3
# get null trafficIntent
trafficIntent=""
print_msg "Deleting controller ${trafficIntent}"
get_resource_negative "${base_url}/controllers/${trafficIntent}"
if [ $return_status == 400 ] ;then
    print_msg "Test:traffic controller get with invalid project name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:traffic controller get with invalid project name. Expected = 400, Actual = $return_status FAILED"
fi

#END
