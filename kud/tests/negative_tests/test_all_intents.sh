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

# Script name: ./test_all_intents.sh
# Purpose: To ascertain whether or not the POST/DELETE/GET API is able to register a null name
# Description, userdata1, and userdata2 have values that I have assigned.

set -o errexit
set -o nounset
set -o pipefail

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
install_deps
populate_CSAR_composite_app_helm "$csar_id"

# clean up
delete_all

# Create project
create_project

# Create composite app
create_composite_app

create_app "collectd.tar.gz" "collectd" "collectd_desc"
create_app "prometheus-operator.tar.gz" "prometheus" "prometheus_desc"

create_main_composite_profile

create_profile_app "test_composite_profile1" "collectd" "collectd_profile.tar.gz"
create_profile_app "test_composite_profile2" "prometheus" "prometheus-operator_profile.tar.gz"

create_generic_placement_intent_app1

create_placement_intent_app "appIntentForApp1" "AppIntentForApp1Desc" "collectd"
create_placement_intent_app "appIntentForApp2" "AppIntentForApp2Desc" "prometheus"

create_deployment_intent_group

# BEGIN: Adding intents to an intent group
print_msg "Adding all the intents to the deploymentIntent group"
intentToBeAddedinDeploymentIntentGroup=""
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

# Test-1
# register null intenttoBeAddedinDeploymentIntentGroup
call_api_negative -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents"

if [ $return_status == 405 ] ;then
    print_msg "Test:all_intents post with null all_intents name. Expected = 405, Actual = $return_status PASSED"
else
    print_msg "Test:all_intents-post with null all_intents name. Expected = 405, Actual = $return_status FAILED"
fi

# Test-2
# delete null intenttoBeAddedinDeploymentIntentGroup
intentToBeAddedinDeploymentIntentGroup=""
print_msg "Deleting intentToBeAddedinDeploymentIntentGroup"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents/${intentToBeAddedinDeploymentIntentGroup}"
if [ $return_status == 404 ] ;then
    print_msg "Test:all_intents-delete with null all_intents name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:all_intents-delete with null all_intents name. Expected = 404, Actual = $return_status FAILED"
fi

# Test-3
# get null intenttoBeAddedinDeploymentIntentGroup
intentToBeAddedinDeploymentIntentGroup=""
print_msg "Deleting intentToBeAddedinDeploymentIntentGroup"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents/${intentToBeAddedinDeploymentIntentGroup}"

if [ $return_status == 404 ] ;then
    print_msg "Test:all_intents-get with null all_intents name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:all_intents-get with null all_intents name. Expected = 404, Actual = $return_status FAILED"
fi

delete_all
# END
