
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

# Register composite-app
create_composite_app

# Create entries for app1&app2 in the database
create_app1
create_app2

# TEST-1 null composite app name
# BEGIN: Register the main composite-profile
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

composite_app_name=""
call_api_negative -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles"
if [ $return_status == 405 ] ;then
    print_msg "Test:profile post-1 with null project name. Expected = 405, Actual = $return_status PASSED"
else
    print_msg "Test:profile post-1 with null project name. Expected = 405, Actual = $return_status FAILED"
fi

# TEST-2 null composite profile name
main_composite_profile_name=""
print_msg "Deleting ${main_composite_profile_name}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}"
if [ $return_status == 404 ] ;then
    print_msg "Test:profile delete-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:profile delete-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi

# TEST-3 null main composite profile name
main_composite_profile_name=""
print_msg "Deleting ${main_composite_profile_name}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}"
if [ $return_status == 404 ] ;then
    print_msg "Test:profile get-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:profile get-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi
# BEGIN: Register the main composite-profile

