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
csar_id="cb009bfe-bbee-11e8-9766-525400435678"


app1_helm_path="$CSAR_DIR/$csar_id/prometheus-operator.tar.gz"
app1_profile_path="$CSAR_DIR/$csar_id/prometheus-operator_profile.tar.gz"
app2_helm_path="$CSAR_DIR/$csar_id/collectd.tar.gz"
app2_profile_path="$CSAR_DIR/$csar_id/collectd_profile.tar.gz"


# ---------BEGIN: SET CLM DATA---------------

clusterprovidername="collection-cluster-provider"
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

clustername="edge1"
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

kubeconfigedge1="/opt/kud/multi-cluster/edge1/artifacts/admin.conf"

labelname="LabelA"
labeldata="$(cat<<EOF
{"label-name": "$labelname"}
EOF
)"

clustername2="edge2"
clusterdata2="$(cat<<EOF
{
  "metadata": {
    "name": "$clustername2",
    "description": "description of $clustername2",
    "userData1": "$clustername2 user data 1",
    "userData2": "$clustername2 user data 2"
  }
}
EOF
)"

kubeconfigedge2="/opt/kud/multi-cluster/edge2/artifacts/admin.conf"

labelname2="LabelA"
labeldata2="$(cat<<EOF
{"label-name": "$labelname2"}
EOF
)"

clustername3="cluster1"
clusterdata3="$(cat<<EOF
{
  "metadata": {
    "name": "$clustername3",
    "description": "description of $clustername3",
    "userData1": "$clustername3 user data 1",
    "userData2": "$clustername3 user data 2"
  }
}
EOF
)"

kubeconfigcluster1="/opt/kud/multi-cluster/cluster1/artifacts/admin.conf"

labelname3="LabelForCluster1"
labeldata3="$(cat<<EOF
{"label-name": "$labelname3"}
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
projectname="TestProject"
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
collection_compositeapp_name="CollectionCompositeApp"
compositeapp_version="v1"
compositeapp_data="$(cat <<EOF
{
  "metadata": {
    "name": "${collection_compositeapp_name}",
    "description": "description of ${collection_compositeapp_name}",
    "userData1": "user data 1 for ${collection_compositeapp_name}",
    "userData2": "user data 2 for ${collection_compositeapp_name}"
   },
   "spec":{
      "version":"${compositeapp_version}"
   }
}
EOF
)"

# add app entries for the prometheus app into
# compositeApp

prometheus_app_name="prometheus-operator"
prometheus_helm_chart=${app1_helm_path}

prometheus_app_data="$(cat <<EOF
{
  "metadata": {
    "name": "${prometheus_app_name}",
    "description": "description for app ${prometheus_app_name}",
    "userData1": "user data 2 for ${prometheus_app_name}",
    "userData2": "user data 2 for ${prometheus_app_name}"
   }
}
EOF
)"

# add app entries for the collectd app into
# compositeApp

collectd_app_name="collectd"
collectd_helm_chart=${app2_helm_path}

collectd_app_data="$(cat <<EOF
{
  "metadata": {
    "name": "${collectd_app_name}",
    "description": "description for app ${collectd_app_name}",
    "userData1": "user data 2 for ${collectd_app_name}",
    "userData2": "user data 2 for ${collectd_app_name}"
   }
}
EOF
)"


# Add the composite profile
collection_composite_profile_name="collection_composite-profile"
collection_composite_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${collection_composite_profile_name}",
      "description":"description of ${collection_composite_profile_name}",
      "userData1":"user data 1 for ${collection_composite_profile_name}",
      "userData2":"user data 2 for ${collection_composite_profile_name}"
   }
}
EOF
)"

# Add the prometheus profile data into collection profile data
prometheus_profile_name="prometheus-profile"
prometheus_profile_file=$app1_profile_path
prometheus_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${prometheus_profile_name}",
      "description":"description of ${prometheus_profile_name}",
      "userData1":"user data 1 for ${prometheus_profile_name}",
      "userData2":"user data 2 for ${prometheus_profile_name}"
   },
   "spec":{
      "app-name":  "${prometheus_app_name}"
   }
}
EOF
)"

# Add the collectd profile data into collection profile data
collectd_profile_name="collectd-profile"
collectd_profile_file=$app2_profile_path
collectd_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${collectd_profile_name}",
      "description":"description of ${collectd_profile_name}",
      "userData1":"user data 1 for ${collectd_profile_name}",
      "userData2":"user data 2 for ${collectd_profile_name}"
   },
   "spec":{
      "app-name":  "${collectd_app_name}"
   }
}
EOF
)"


# define the generic placement intent
generic_placement_intent_name="generic-placement-intent"
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

# define app placement intent for prometheus
prometheus_placement_intent_name="prometheus-placement-intent"
prometheus_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${prometheus_placement_intent_name}",
      "description":"description of ${prometheus_placement_intent_name}",
      "userData1":"user data 1 for ${prometheus_placement_intent_name}",
      "userData2":"user data 2 for ${prometheus_placement_intent_name}"
   },
   "spec":{
      "app-name":"${prometheus_app_name}",
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

# define app placement intent for collectd
collectd_placement_intent_name="collectd-placement-intent"
collectd_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${collectd_placement_intent_name}",
      "description":"description of ${collectd_placement_intent_name}",
      "userData1":"user data 1 for ${collectd_placement_intent_name}",
      "userData2":"user data 2 for ${collectd_placement_intent_name}"
   },
   "spec":{
      "app-name":"${collectd_app_name}",
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
release="collection"
deployment_intent_group_name="collection_deployment_intent_group"
deployment_intent_group_data="$(cat <<EOF
{
   "metadata":{
      "name":"${deployment_intent_group_name}",
      "description":"descriptiont of ${deployment_intent_group_name}",
      "userData1":"user data 1 for ${deployment_intent_group_name}",
      "userData2":"user data 2 for ${deployment_intent_group_name}"
   },
   "spec":{
      "profile":"${collection_composite_profile_name}",
      "version":"${release}",
      "override-values":[]
   }
}
EOF
)"

# define the intents to be used by the group
deployment_intents_in_group_name="collection_deploy_intents"
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

    print_msg "creating collection composite app entry"
    call_api -d "${compositeapp_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps"

    print_msg "adding prometheus app to the composite app"
    call_api -F "metadata=${prometheus_app_data}" \
             -F "file=@${prometheus_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/apps"

    print_msg "adding collectd app to the composite app"
    call_api -F "metadata=${collectd_app_data}" \
             -F "file=@${collectd_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/apps"

    print_msg "creating collection composite profile entry"
    call_api -d "${collection_composite_profile_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles"

    print_msg "adding prometheus app profiles to the composite profile"
    call_api -F "metadata=${prometheus_profile_data}" \
             -F "file=@${prometheus_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles/${collection_composite_profile_name}/profiles"

    print_msg "adding collectd app profiles to the composite profile"
    call_api -F "metadata=${collectd_profile_data}" \
             -F "file=@${collectd_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles/${collection_composite_profile_name}/profiles"

    print_msg "create the generic placement intent"
    call_api -d "${generic_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/generic-placement-intents"

    print_msg "add the prometheus app placement intent to the generic placement intent"
    call_api -d "${prometheus_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents"

    print_msg "add the collectd app placement intent to the generic placement intent"
    call_api -d "${collectd_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents"


    print_msg "create the deployment intent group"
    call_api -d "${deployment_intent_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups"
    call_api -d "${deployment_intents_in_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents"

}

function deleteOrchestratorData {
   # TODO- delete rsync controller and any other controller
    delete_resource "${base_url_orchestrator}/controllers/${rsynccontrollername}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents/${deployment_intents_in_group_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${prometheus_placement_intent_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${collectd_placement_intent_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles/${collection_composite_profile_name}/profiles/${prometheus_profile_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles/${collection_composite_profile_name}/profiles/${collectd_profile_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles/${collection_composite_profile_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/apps/${prometheus_app_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/apps/${collectd_app_name}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}"
}


function createClmData {
    print_msg "Creating cluster provider and cluster"
    call_api -d "${clusterproviderdata}" "${base_url_clm}/cluster-providers"

    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata" -F "file=@$kubeconfigedge1" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"

    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata2" -F "file=@$kubeconfigedge2" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"

    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata3" -F "file=@$kubeconfigcluster1" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"

    call_api -d "${labeldata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels"

    call_api -d "${labeldata2}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/labels"

    call_api -d "${labeldata3}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername3}/labels"
}

function deleteClmData {

    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels/${labelname}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/labels/${labelname2}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername3}/labels/${labelname3}"

    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername3}"

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
# function ApplyNcmData {
#    call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/apply"
#     call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/apply"
# }

# function terminateNcmData {
#     call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/terminate"
#     call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/terminate"
# }

function instantiate {
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/approve"
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/instantiate"
}


function terminateOrchData {
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/terminate"
}

# Setup
install_deps
populate_CSAR_composite_app_helm "$csar_id"

#terminateOrchData
deleteData
createData
instantiate

