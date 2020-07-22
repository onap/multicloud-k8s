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

# Script name: ./test_profile_apps.sh
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


# TEST-1 null main composite profile name
# BEGIN : Adding profile to each of the two apps - app1(collectd) and app2(prometheus)
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

main_composite_profile_name=""

call_api_negative -F "metadata=$payload" \
        -F "file=@$app1_profile_path" \
    "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles"
if [ $return_status == 405 ] ;then
    print_msg "Test:profile app post-1 with null project name. Expected = 405, Actual = $return_status PASSED"
else
    print_msg "Test:profile app post-1 with null project name. Expected = 405, Actual = $return_status FAILED"
fi

# TEST-2 
# delete null sub composite profile name
sub_composite_profile_name1=""
print_msg "Deleting ${sub_composite_profile_name1}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"
if [ $return_status == 404 ] ;then
    print_msg "Test:profile app delete-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:profile app delete-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi


# TEST-3
# null get sub composite profile name
print_msg "Deleting ${sub_composite_profile_name1}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"
if [ $return_status == 404 ] ;then
    print_msg "Test:profile app get-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:profile app get-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi

delete_all
#END
