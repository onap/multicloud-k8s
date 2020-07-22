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
app2_helm_path="$CSAR_DIR/$csar_id/prometheus-operator.tar.gz"
app1_profile_path="$CSAR_DIR/$csar_id/collectd_profile.tar.gz"
app2_profile_path="$CSAR_DIR/$csar_id/prometheus-operator_profile.tar.gz"

app1_name="collectd"
app2_name="prometheus-operator"
app1_desc="collectd_desc"
app2_desc="prometheus_desc"

main_composite_profile_name="main_composite_profile"
sub_composite_profile_name1="test_composite_profile1"
sub_composite_profile_name2="test_composite_profile2"
main_composite_profile_description="main_composite_profile_description"
sub_composite_profile_description="sub_composite_profile_description"

genericPlacementIntentName="test_gen_placement_intent1"
genericPlacementIntentDesc="test_gen_placement_intent_desc"
logicalCloud="logical_cloud_name"

appIntentNameForApp1="appIntentForApp1"
appIntentForApp1Desc="AppIntentForApp1Desc"
appIntentNameForApp2="appIntentForApp2"
appIntentForApp2Desc="AppIntentForApp2Desc"
providerName1="cluster_provider1"
providerName2="cluster_provider2"
clusterName1="clusterName1"
clusterName2="clusterName2"
clusterLabelName1="clusterLabel1"
clusterLabelName2="clusterLabel2"

deploymentIntentGroupName="test_deployment_intent_group"
deploymentIntentGroupNameDesc="test_deployment_intent_group_desc"
releaseName="test"
intentToBeAddedinDeploymentIntentGroup="name_of_intent_to_be_added_in_deployment_group"
intentToBeAddedinDeploymentIntentGroupDesc="desc_of_intent_to_be_added_in_deployment_group"
hpaIntentName="hpaIntentName"
trafficIntentName="trafficIntentName"

chart_name="edgex"
profile_name="test_profile"
release_name="test-release"
namespace="plugin-tests-namespace"
cloud_region_id="kud"
cloud_region_owner="localhost"


# Controllers
genericPlacementIntent="genericPlacementIntent"
OVNintent="OVNintent"
OVNintentName="OVNintentName"
OVNHostName="OVNHostName"
OVNPort="9027"
CostBasedIntent="costBasedIntent"
CostBasedIntentName="CostBasedIntentName"
CostBasedHostName="OVNHostName"
CostBasedPort="9028"
hpaIntent="hpaIntent"
trafficIntent="trafficIntent"
gpcHostName="gpcHostName"
gpcPort="9029"
hpaControllerIntentName="hpaControllerIntentName"
hpaHostName="hpaHostName"
hpaPort="9030"
trafficControllerIntentName="trafficControllerIntentName"
trafficHostName="trafficHostName"
trafficPort="9031"

# Setup
install_deps
populate_CSAR_composite_app_helm "$csar_id"

# BEGIN :: Delete statements are issued so that we clean up the 'orchestrator' collection
# and freshly populate the documents, also it serves as a direct test
# for all our DELETE APIs and an indirect test for all GET APIs
print_msg "Deleting controller ${genericPlacementIntent}"
delete_resource "${base_url}/controllers/${genericPlacementIntent}"

print_msg "Deleting controller ${hpaIntent}"
delete_resource "${base_url}/controllers/${hpaIntent}"

print_msg "Deleting controller ${trafficIntent}"
delete_resource "${base_url}/controllers/${trafficIntent}"

print_msg "Deleting controller ${CostBasedIntent}"
delete_resource "${base_url}/controllers/${CostBasedIntent}"

print_msg "Deleting controller ${OVNintent}"
delete_resource "${base_url}/controllers/${OVNintent}"

print_msg "Deleting intentToBeAddedinDeploymentIntentGroup"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents/${intentToBeAddedinDeploymentIntentGroup}"

print_msg "Deleting ${deploymentIntentGroupName}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}"

print_msg "Deleting ${appIntentNameForApp2}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp2}"

print_msg "Deleting ${appIntentNameForApp1}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents/${appIntentNameForApp1}"

print_msg "Deleting ${genericPlacementIntentName}"
delete_resource "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}"

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

call_api -F "metadata=$payload" \
         -F "file=@$app1_profile_path" \
         "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/composite-profiles/${main_composite_profile_name}/profiles"

print_msg "Registering profile with app2(prometheus)"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${sub_composite_profile_name2}",
      "description":"${sub_composite_profile_description}",
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
      "name":"${genericPlacementIntentName}",
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
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents"

print_msg "Adding placement intent for app2(prometheus)"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${appIntentNameForApp2}",
      "description":"${appIntentForApp2Desc}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "app-name":"${app2_name}",
      "intent":{
         "allOf":[
            {
               "provider-name":"${providerName1}",
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
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/generic-placement-intents/${genericPlacementIntentName}/app-intents"
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
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups"
# END: Registering DeploymentIntentGroup in the database

# BEGIN: Adding intents to an intent group
print_msg "Adding all the intents to the deploymentIntent group"
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
call_api -d "${payload}" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/intents"
# END: Adding intents to an intent group

# BEGIN: Adding controllers
print_msg "Adding CostBased placement contoller"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${CostBasedIntent}",
      "description":"${CostBasedIntentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${CostBasedHostName}",
      "port": ${CostBasedPort},
      "type": "placement",
      "priority": 3
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/controllers"

print_msg "Adding HPA contoller"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${hpaIntent}",
      "description":"${hpaControllerIntentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${hpaHostName}",
      "port": ${hpaPort},
      "type": "placement",
      "priority": 2
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/controllers"

print_msg "Adding traffic contoller"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${trafficIntent}",
      "description":"${trafficControllerIntentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${trafficHostName}",
      "port": ${trafficPort},
      "type": "action",
      "priority": 3
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/controllers"

print_msg "Adding OVN action contoller"
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${OVNintent}",
      "description":"${OVNintentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${OVNHostName}",
      "port": ${OVNPort},
      "type": "action",
      "priority": 2
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/controllers"
# END: Adding controllers

#BEGIN: Instantiation
print_msg "Getting the sorted templates for each of the apps.."
call_api -d "" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/approve"
call_api -d "" "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/deployment-intent-groups/${deploymentIntentGroupName}/instantiate"
# END: Instantiation
