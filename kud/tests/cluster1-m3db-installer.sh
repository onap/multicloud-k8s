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


source _common_test.sh
source _functions.sh
source _common.sh


base_url_clm=${base_url_clm:-"http://192.168.121.29:30073/v2"}
base_url_ncm=${base_url_ncm:-"http://192.168.121.29:31955/v2"}
base_url_orchestrator=${base_url_orchestrator:-"http://192.168.121.29:32447/v2"}
base_url_rysnc=${base_url_orchestrator:-"http://192.168.121.29:32002/v2"}

CSAR_DIR="/opt/csar"
csar_id="m3db-cb009bfe-bbee-11e8-9766-525400435678"

app1_helm_path="$CSAR_DIR/$csar_id/m3db.tar.gz"
app1_profile_path="$CSAR_DIR/$csar_id/m3db_profile.tar.gz"

# ---------BEGIN: SET CLM DATA---------------

clusterprovidername="collection-m3db-installer-cluster1-provider"
clusterproviderdata="$(cat<<EOF
{
  "metadata": {
    "name": "$clusterprovidername",
    "description": "description of $clusterprovidername",
    "userData1": "$clusterprovidername user data 1",
    "userData2": "$clusterprovidername user data 2"
  }
}
EOF
)"

clustername="cluster1"
clusterdata="$(cat<<EOF
{
  "metadata": {
    "name": "$clustername",
    "description": "description of $clustername",
    "userData1": "$clustername user data 1",
    "userData2": "$clustername user data 2"
  }
}
EOF
)"

kubeconfigcluster1="/opt/kud/multi-cluster/cluster1/artifacts/admin.conf"

labelname="LabelCluster2"
labeldata="$(cat<<EOF
{"label-name": "$labelname"}
EOF
)"

#--TODO--Creating provider network and network intents----

# add the rsync controller entry
rsynccontrollername="rsync"
rsynccontrollerdata="$(cat<<EOF
{
  "metadata": {
    "name": "rsync",
    "description": "description of $rsynccontrollername controller",
    "userData1": "user data 1 for $rsynccontrollername",
    "userData2": "user data 2 for $rsynccontrollername"
  },
  "spec": {
    "host": "${rsynccontrollername}",
    "port": 9041
  }
}
EOF
)"

# ------------END: SET CLM DATA--------------

#-------------BEGIN:SET ORCH DATA------------------

# define a project
projectname="M3dbInstaller-cluster1Old-Project"
projectdata="$(cat<<EOF
{
  "metadata": {
    "name": "$projectname",
    "description": "description of $projectname controller",
    "userData1": "$projectname user data 1",
    "userData2": "$projectname user data 2"
  }
}
EOF
)"

# define a composite application
m3db_compositeapp_name="OperatorsCompositeApp"
compositeapp_version="v1"
compositeapp_data="$(cat <<EOF
{
  "metadata": {
    "name": "${m3db_compositeapp_name}",
    "description": "description of ${m3db_compositeapp_name}",
    "userData1": "user data 1 for ${m3db_compositeapp_name}",
    "userData2": "user data 2 for ${m3db_compositeapp_name}"
   },
   "spec":{
      "version":"${compositeapp_version}"
   }
}
EOF
)"


# add m3db app into compositeApp

m3db_app_name="m3db"
m3db_helm_chart=${app1_helm_path}

m3db_app_data="$(cat <<EOF
{
  "metadata": {
    "name": "${m3db_app_name}",
    "description": "description for app ${m3db_app_name}",
    "userData1": "user data 2 for ${m3db_app_name}",
    "userData2": "user data 2 for ${m3db_app_name}"
   }
}
EOF
)"

# Add the composite profile
m3db_composite_profile_name="m3db_composite-profile"
m3db_composite_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${m3db_composite_profile_name}",
      "description":"description of ${m3db_composite_profile_name}",
      "userData1":"user data 1 for ${m3db_composite_profile_name}",
      "userData2":"user data 2 for ${m3db_composite_profile_name}"
   }
}
EOF
)"


# Add the m3db profile data into composite profile data
m3db_profile_name="m3db-profile"
m3db_profile_file=$app1_profile_path
m3db_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${m3db_profile_name}",
      "description":"description of ${m3db_profile_name}",
      "userData1":"user data 1 for ${m3db_profile_name}",
      "userData2":"user data 2 for ${m3db_profile_name}"
   },
   "spec":{
      "app-name":  "${m3db_app_name}"
   }
}
EOF
)"


# define the generic placement intent
generic_placement_intent_name="M3db-generic-placement-intent"
generic_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${generic_placement_intent_name}",
      "description":"${generic_placement_intent_name}",
      "userData1":"${generic_placement_intent_name}",
      "userData2":"${generic_placement_intent_name}"
   },
   "spec":{
      "logical-cloud":"unused_logical_cloud"
   }
}
EOF
)"


# define placement intent for m3db as sub-app
m3db_placement_intent_name="m3db-placement-intent"
m3db_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${m3db_placement_intent_name}",
      "description":"description of ${m3db_placement_intent_name}",
      "userData1":"user data 1 for ${m3db_placement_intent_name}",
      "userData2":"user data 2 for ${m3db_placement_intent_name}"
   },
   "spec":{
      "app-name":"${m3db_app_name}",
      "intent":{
         "allOf":[
            {  "provider-name":"${clusterprovidername}",
               "cluster-label-name":"${labelname}"
            }
         ]
      }
   }
}
EOF
)"


# define a deployment intent group
release="m3db"
deployment_intent_group_name="m3db_deployment_intent_group"
deployment_intent_group_data="$(cat <<EOF
{
   "metadata":{
      "name":"${deployment_intent_group_name}",
      "description":"descriptiont of ${deployment_intent_group_name}",
      "userData1":"user data 1 for ${deployment_intent_group_name}",
      "userData2":"user data 2 for ${deployment_intent_group_name}"
   },
   "spec":{
      "profile":"${m3db_composite_profile_name}",
      "version":"${release}",
      "override-values":[]
   }
}
EOF
)"

# define the intents to be used by the group
deployment_intents_in_group_name="m3db_deploy_intents"
deployment_intents_in_group_data="$(cat <<EOF
{
   "metadata":{
      "name":"${deployment_intents_in_group_name}",
      "description":"descriptionf of ${deployment_intents_in_group_name}",
      "userData1":"user data 1 for ${deployment_intents_in_group_name}",
      "userData2":"user data 2 for ${deployment_intents_in_group_name}"
   },
   "spec":{
      "intent":{
         "genericPlacementIntent":"${generic_placement_intent_name}"
      }
   }
}
EOF
)"

#---------END: SET ORCH DATA--------------------

function createOrchestratorData {
    print_msg "creating controller entries"
    call_api -d "${rsynccontrollerdata}" "${base_url_orchestrator}/controllers"


    print_msg "creating project entry"
    call_api -d "${projectdata}" "${base_url_orchestrator}/projects"


    print_msg "creating m3db composite app entry"
    call_api -d "${compositeapp_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps"

    print_msg "adding m3db sub-app to the composite app"
    call_api -F "metadata=${m3db_app_data}" \
             -F "file=@${m3db_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/apps"


    print_msg "creating m3db composite profile entry"
    call_api -d "${m3db_composite_profile_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/composite-profiles"

    print_msg "adding m3db sub-app profile to the composite profile"
    call_api -F "metadata=${m3db_profile_data}" \
             -F "file=@${m3db_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/composite-profiles/${m3db_composite_profile_name}/profiles"



    print_msg "create the generic placement intent"
    call_api -d "${generic_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/generic-placement-intents"
    print_msg "add the m3db app placement intent to the generic placement intent"
    call_api -d "${m3db_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents"

    print_msg "create the deployment intent group"
    call_api -d "${deployment_intent_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/deployment-intent-groups"
    call_api -d "${deployment_intents_in_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents"
    print_msg "finished orch data creation"
}

function deleteOrchestratorData {

    # TODO- delete rsync controller and any other controller
    delete_resource "${base_url_orchestrator}/controllers/${rsynccontrollername}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents/${deployment_intents_in_group_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${m3db_placement_intent_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/composite-profiles/${m3db_composite_profile_name}/profiles/${m3db_profile_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/composite-profiles/${m3db_composite_profile_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/apps/${m3db_app_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}"
    print_msg "finished orch data deletion"

}

function createClmData {
    print_msg "Creating cluster provider and cluster"
    call_api -d "${clusterproviderdata}" "${base_url_clm}/cluster-providers"
    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata" -F "file=@$kubeconfigcluster1" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"
    call_api -d "${labeldata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels"

}


function deleteClmData {
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels/${labelname}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}"
}


function createData {
    createClmData
    createOrchestratorData
}

function deleteData {
    deleteClmData
    deleteOrchestratorData
}

function instantiate {
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/instantiate"
}

function terminateOrchData {
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${m3db_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/terminate"
}


# Setup
install_deps
populate_CSAR_m3db_helm "$csar_id"

#terminateOrchData
deleteData
createData
instantiate


