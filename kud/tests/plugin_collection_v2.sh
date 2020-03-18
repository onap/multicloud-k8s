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

composite_app_name="test_composite_app"
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
composite_profile_description="test_composite_profile_description"

genericPlacementIntentName1="test_gen_placement_intent1"
genericPlacementIntentName2="test_gen_placement_intent2"
genericPlacementIntentDesc="test_gen_placement_intent_desc"
logicalCloud="logical_cloud_name"
clusterName1="edge1"
clusterName2="edge2"
clusterLabelName1="east-us1"
clusterLabelName2="east-us2"

deploymentIntentGroupName="test_deployment_intent_group"
deploymentIntentGroupNameDesc="test_deployment_intent_group_desc"

chart_name="edgex"
profile_name="test_profile"
release_name="test-release"
namespace="plugin-tests-namespace"
cloud_region_id="kud"
cloud_region_owner="localhost"

# Setup
install_deps
populate_CSAR_composite_app_helm "$csar_id"

# BEGIN: Register project API
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
# END: Register project API

# BEGIN: Register composite-app API
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
# END: Register composite-app API




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
      "description":"${composite_profile_description}",
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
      "description":"${composite_profile_description}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "app-name":  "${app1_name}"
   }
}
EOF
)"

call_api -F "metadata=$payload" \
         -F "file=@$app1_profile_path" \
         "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles"

print_msg "Registering profile with app2(prometheus)"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${sub_composite_profile_name2}",
      "description":"${composite_profile_description}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "app-name":  "${app2_name}"
   }
}
EOF
)"

call_api -F "metadata=$payload" \
         -F "file=@$app2_profile_path" \
         "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles"
# END : Adding profile to each of the two apps - app1(collectd) and app2(prometheus)

# BEGIN: Register GenericPlacementIntents with the database
print_msg "Registering GenericPlacementIntent for app1"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${genericPlacementIntentName1}",
      "description":"${genericPlacementIntentDesc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "logical-cloud":"${logicalCloud}"
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents"


print_msg "Registering GenericPlacementIntent for app2"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${genericPlacementIntentName2}",
      "description":"${genericPlacementIntentDesc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "logical-cloud":"${logicalCloud}"
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents"
# END: Register GenericPlacementIntents with the database

# BEGIN: Adding placement intent for each app in the composite app.
print_msg "Adding placement intent for app1(collectd)"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${genericPlacementIntentName1}",
      "description":"${genericPlacementIntentDesc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "app-name":"${app1_name}",
      "intent":{
         "allOf":[
            {
               "cluster-name":"${clusterName1}"
            },
            {
               "cluster-name":"${clusterName2}"
            },
            {
               "anyOf":[
                  {
                     "cluster-label-name":"${clusterLabelName1}"
                  },
                  {
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
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName1}/app-intents"

print_msg "Adding placement intent for app2(prometheus)"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${genericPlacementIntentName2}",
      "description":"${genericPlacementIntentDesc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "app-name":"${app2_name}",
      "intent":{
         "allOf":[
            {
               "cluster-name":"${clusterName1}"
            },
            {
               "cluster-name":"${clusterName2}"
            },
            {
               "anyOf":[
                  {
                     "cluster-label-name":"${clusterLabelName1}"
                  },
                  {
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
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName2}/app-intents"
# END: Adding placement intent for each app in the composite app.

# BEGIN: Registering DeploymentIntentGroup in the database
print_msg "Registering DeploymentIntentGroup"
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
      "version":"${composite_app_version}",
      "override-values":[
         {
            "app-name":"${app1_name}",
            "values":
               {
                  "imageRepository":"registry.hub.docker.com"
               }
         },
         {
            "app-name":"${app2_name}",
            "values":
               {
                  "imageRepository":"registry.hub.docker.com"
               }
         }
      ]
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups"
# END: Registering DeploymentIntentGroup in the database

# BEGIN: Adding intents to an intent group
print_msg "Adding two intents to the intent group"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${deploymentIntentGroupName}",
      "description":"${deploymentIntentGroupNameDesc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "intent":{
         "generic-placement-intent":"${genericPlacementIntentName1}",
         "generic-placement-intent":"${genericPlacementIntentName2}"
      }
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents"
# END: Adding intents to an intent group

