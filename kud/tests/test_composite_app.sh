
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

# Negative tests for "composite_app "
# Aditya Sharoff<aditya.sharoff@intel.com> 07/14/2020

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

base_url=${base_url:-"http://localhost:9015/v2"}

kubeconfig_path="$HOME/.kube/config"
csar_id=cb009bfe-bbee-11e8-9766-525400435678


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

# Clean up
delete_all

# Register project
create_project

# TEST-1 
# Register a null composite app
composite_app_name=""
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
call_api_negative -d "${payload}" "${base_url}/projects/${project_name}/composite-apps"
if [ $return_status == 405 ] ;then
print_msg "Test: composite application post expected value = 405 actual value = $return_status PASSED"
else
print_msg "Test: composite application post expected value = 405 actual value = $return_status FAILED"
fi

# TEST-2
# Delete a non existing composite app
composite_app_name=""
print_msg "Deleting ${composite_app_name}/${composite_app_version}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}"
if [ $return_status == 400 ] ;then
print_msg "Test: composite application delete expected value = 400 actual value = $return_status PASSED"
else
print_msg "Test: composite application delete expected value = 400 actual value = $return_status FAILED"
fi

composite_app_name=""
print_msg "Deleting ${composite_app_name}/${composite_app_version}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}"
if [ $return_status == 400 ] ;then
print_msg "Test: composite application get expected value = 400 actual value = $return_status PASSED"
else
print_msg "Test: composite application get expected value = 400 actual value = $return_status FAILED"
fi
# END: Register composite-app
