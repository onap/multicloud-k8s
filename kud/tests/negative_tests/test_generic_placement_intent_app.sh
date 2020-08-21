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

# Script name: ./test_generic_placement_intent_app.sh
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

# Clean up
delete_all

# Register project
create_project

# Register composite-app
create_composite_app

# Create entries for app1&app2 in the database
create_app "collectd.tar.gz" "collectd" "collectd_desc"
create_app "prometheus-operator.tar.gz" "prometheus" "prometheus_desc"

# Register the main composite-profile
create_main_composite_profile

# Adding profile to each of the two apps - app1(collectd) and app2(prometheus)
create_profile_app "test_composite_profile1" "collectd" "collectd_profile.tar.gz"
create_profile_app "test_composite_profile2" "prometheus" "prometheus-operator_profile.tar.gz"

# Register GenericPlacementIntents with the database
create_generic_placement_intent_app1

# Adding placement intent for each app in the composite app.
print_msg "Adding placement intent for app1(collectd)"
appIntentNameForApp1=""
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

# Test-1
# register with null va;ue
call_api_negative -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents"
if [ $return_status == 400 ] ;then
    print_msg "Test:generic_placement_intent_app-post1 with null project name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:generic_placement_intent_app post-1 with null project name. Expected = 400, Actual = $return_status FAILED"
fi

# Test-2
# delete a mull
print_msg "Deleting ${appIntentNameForApp1}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp1}"
if [ $return_status == 400 ] ;then
    print_msg "Test:generic_placement_intent_app-delete1 with null project name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:generic_placement_intent_app delete-1 with null project name. Expected = 400, Actual = $return_status FAILED"
fi

# Test-3
# get a null
print_msg "Deleting ${appIntentNameForApp1}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp1}"
if [ $return_status == 400 ] ;then
    print_msg "Test:generic_placement_intent_app-get1 with null project name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:generic_placement_intent_app get-1 with null project name. Expected = 400, Actual = $return_status FAILED"
fi

delete_all
#END
