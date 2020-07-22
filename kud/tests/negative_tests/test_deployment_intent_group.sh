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

# Script name: ./test_deployment_intent_group.sh
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

# Cleanup
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
create_placement_intent_app "appIntentForApp1" "AppIntentForApp1Desc" "collectd"
create_placement_intent_app "appIntentForApp2" "AppIntentForApp2Desc" "prometheus"

# BEGIN: Registering DeploymentIntentGroup in the database
print_msg "Registering DeploymentIntentGroup"

deploymentIntentGroupName=""
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

# Test-1
# register null deploymentIntentGroupName
call_api_negative -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups"
if [ $return_status == 405 ] ;then
    print_msg "Test:deployment_intent_group post-1 with null name. Expected = 405, Actual = $return_status PASSED"
else
    print_msg "Test:deployment_intent_group post1 with null name. Expected = 405, Actual = $return_status FAILED"
fi

# Test-2
# delete null deploymentIntentGroupName
deploymentIntentGroupName=""
print_msg "Deleting ${deploymentIntentGroupName}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}"
if [ $return_status == 404 ] ;then
    print_msg "Test:deployment_intent_group-delete-1 with null name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:deployment_intent_group-delete-1 with null name. Expected = 404, Actual = $return_status FAILED"
fi

# Test-3
# get null deploymentIntentGroupName
deploymentIntentGroupName=""
print_msg "Getting ${deploymentIntentGroupName}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}"
if [ $return_status == 404 ] ;then
    print_msg "Test:deployment_intent_group-get-1 with null name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:deployment_intent_group-get-1 with null name. Expected = 404, Actual = $return_status FAILED"
fi

delete_all
#END
