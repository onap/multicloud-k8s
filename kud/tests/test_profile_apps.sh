
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

# BEGIN :: Delete statements are issued so that we clean up the 'orchestrator' collection
# and freshly populate the documents, also it serves as a direct test
# for all our DELETE APIs and an indirect test for all GET APIs
print_msg "Deleting ${sub_composite_profile_name2}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name2}"

print_msg "Deleting ${sub_composite_profile_name1}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"

print_msg "Deleting ${main_composite_profile_name}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}"

print_msg "Deleting ${app2_name}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app2_name}"

print_msg "Deleting ${app1_name}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app1_name}"

print_msg "Deleting ${composite_app_name}/${composite_app_version}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}"

print_msg "Deleting ${project_name}"
delete_resource "${base_url}/projects/${project_name}"




# END :: Delete statements were issued so that we clean up the db
# and freshly populate the documents, also it serves as a direct test
# for all our DELETE APIs and an indirect test for all GET APIs


# BEGIN: Register project
print_msg "Registering project"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${project_name}",
    "description": "${project_description}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/projects"
# END: Register project

# BEGIN: Register composite-app
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
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps"
# END: Register composite-app




# BEGIN: Create entries for app1&app2 in the database
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

call_api -F "metadata=$payload" \
         -F "file=@$app1_helm_path" \
         "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps"


# BEGIN: Create an entry for app2 in the database
print_msg "Making app entry in the database"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${app2_name}",
    "description": "${app2_desc}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"

call_api -F "metadata=$payload" \
         -F "file=@$app2_helm_path" \
         "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps"
# END: Create entries for app1&app2 in the database


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
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles"
# BEGIN: Register the main composite-profile


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
if [ $return_status == 405 ]
then
	print_msg "Test:profile app post-1 with null project name. Expected = 405, Actual = $return_status PASSED"
else
	print_msg "Test:profile app post-1 with null project name. Expected = 405, Actual = $return_status FAILED"
fi

sub_composite_profile_name1=""
print_msg "Deleting ${sub_composite_profile_name1}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"
if [ $return_status == 404 ]
then
	print_msg "Test:profile app delete-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
	print_msg "Test:profile app delete-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi

print_msg "Deleting ${sub_composite_profile_name1}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles/${sub_composite_profile_name1}"
if [ $return_status == 404 ]
then
	print_msg "Test:profile app get-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
	print_msg "Test:profile app get-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi
