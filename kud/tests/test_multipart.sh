
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

# Negative tests for "Multipart"
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

app1_name="collectd"
app2_name="prometheus"
app1_desc="collectd_desc"
app2_desc="prometheus_desc"

# Cleanup
delete_all

# Register project
create_project

# Register composite-app
create_composite_app

# Create entries for app1&app2 in the database
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

app1_helm_path=""

call_api_negative -F "metadata=$payload" \
        -F "file=@$app1_helm_path" \
    "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps"
if [ $return_status == 400 ] ;then
    print_msg "Test:multipart-post-1 with null name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:multipart pos-1 with null name. Expected = 400, Actual = $return_status FAILED"
fi

app1_name=""
print_msg "Deleting ${app1_name}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app1_name}"
if [ $return_status == 400 ] ;then
    print_msg "Test:multipart delete-1 with null name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:multipart delete-1 with null name. Expected = 400, Actual = $return_status FAILED"
fi

app1_name=""
print_msg "getting ${app1_name}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app1_name}"
if [ $return_status == 400 ] ;then
    print_msg "Test:multipart get-1 with null name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:multipart get-1 with null name. Expected = 400, Actual = $return_status FAILED"
fi

#END
