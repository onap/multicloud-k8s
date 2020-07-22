
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

main_composite_profile_name="main_composite_profile"
sub_composite_profile_name1="test_composite_profile1"
sub_composite_profile_name2="test_composite_profile2"
main_composite_profile_description="main_composite_profile_description"
sub_composite_profile_description="sub_composite_profile_description"

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
create_app1
create_app2


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

# TEST-2 delete null sub composite profile name
sub_composite_profile_name1=""
print_msg "Deleting ${sub_composite_profile_name1}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"
if [ $return_status == 404 ] ;then
    print_msg "Test:profile app delete-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:profile app delete-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi


# TEST-3null get sub composite profile name
print_msg "Deleting ${sub_composite_profile_name1}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"
if [ $return_status == 404 ] ;then
    print_msg "Test:profile app get-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:profile app get-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi
