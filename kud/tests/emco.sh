#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
# Copyright (c) 2020 Intel Corporation

set -o errexit
set -o nounset
set -o pipefail

source _common_test.sh
source _functions.sh
#source _common.sh

# TODO Workaround for MULTICLOUD-1202
function delete_resource_nox {
    call_api_nox -X DELETE "$1"
    ! call_api -X GET "$1" >/dev/null
}

master_ip=$(kubectl cluster-info | grep "Kubernetes master" | \
    awk -F ":" '{print $2}' | awk -F "//" '{print $2}')
rsync_service_port=30441
rsync_service_host="$master_ip"
base_url_orchestrator=${base_url_orchestrator:-"http://$master_ip:30415/v2"}
base_url_clm=${base_url_clm:-"http://$master_ip:30461/v2"}
base_url_dcm=${base_url_dcm:-"http://$master_ip:30477/v2"}

CSAR_DIR="/opt/csar"
csar_id="cb009bfe-bbee-11e8-9766-525400435678"

app1_helm_path="$CSAR_DIR/$csar_id/prometheus-operator.tar.gz"
app1_profile_path="$CSAR_DIR/$csar_id/prometheus-operator_profile.tar.gz"
app2_helm_path="$CSAR_DIR/$csar_id/collectd.tar.gz"
app2_profile_path="$CSAR_DIR/$csar_id/collectd_profile.tar.gz"

kubeconfig_path="$HOME/.kube/config"

function populate_CSAR_composite_app_helm {
    _checks_args "$1"
    pushd "${CSAR_DIR}/$1"
    print_msg "Create Helm Chart Archives for compositeApp"
    rm -f *.tar.gz
    tar -czf collectd.tar.gz -C $test_folder/vnfs/comp-app/collection/app1/helm .
    tar -czf prometheus-operator.tar.gz -C $test_folder/vnfs/comp-app/collection/app2/helm .
    tar -czf collectd_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/app1/profile .
    tar -czf prometheus-operator_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/app2/profile .
    popd
}


# ---------BEGIN: SET CLM DATA---------------

clusterprovidername="sanity-test-cluster-provider"
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

clustername="LocalEdge1"
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


labelname="LocalLabel"
labeldata="$(cat<<EOF
{"label-name": "$labelname"}
EOF
)"

admin_logical_cloud_name="lcadmin"
admin_logical_cloud_data="$(cat << EOF
{
 "metadata" : {
    "name": "${admin_logical_cloud_name}",
    "description": "logical cloud description",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "level": "0"
  }
 }
}
EOF
)"

lc_cluster_1_name="lc1-c1"
cluster_1_data="$(cat << EOF
{
 "metadata" : {
    "name": "${lc_cluster_1_name}",
    "description": "logical cloud cluster 1 description",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },

 "spec" : {
    "cluster-provider": "${clusterprovidername}",
    "cluster-name": "${clustername}",
    "loadbalancer-ip" : "0.0.0.0"
  }
}
EOF
)"

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
    "host": "$rsync_service_host",
    "port": $rsync_service_port
  }
}
EOF
)"

# ------------END: SET CLM DATA--------------


#-------------BEGIN:SET ORCH DATA------------------

# define a project
projectname="Sanity-Test-Project"
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
generic_placement_intent_name="test-generic-placement-intent"
generic_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${generic_placement_intent_name}",
      "description":"${generic_placement_intent_name}",
      "userData1":"${generic_placement_intent_name}",
      "userData2":"${generic_placement_intent_name}"
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
release="test-collection"
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
      "override-values":[],
      "logical-cloud":"${admin_logical_cloud_name}"
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

    createLogicalCloudData

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

    print_msg "create the deployment intent group"
    call_api -d "${deployment_intent_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups"

    print_msg "create the generic placement intent"
    call_api -d "${generic_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents"

    print_msg "add the prometheus app placement intent to the generic placement intent"
    call_api -d "${prometheus_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents"

    print_msg "add the collectd app placement intent to the generic placement intent"
    call_api -d "${collectd_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents"

    call_api -d "${deployment_intents_in_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents"

}

function deleteOrchestratorData {

    print_msg "Begin deleteOrchestratorData"

    delete_resource_nox "${base_url_orchestrator}/controllers/${rsynccontrollername}"

    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents/${deployment_intents_in_group_name}"

    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${prometheus_placement_intent_name}"
    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${collectd_placement_intent_name}"
    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}"

    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles/${collection_composite_profile_name}/profiles/${prometheus_profile_name}"
    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles/${collection_composite_profile_name}/profiles/${collectd_profile_name}"

    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}"

    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/composite-profiles/${collection_composite_profile_name}"

    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/apps/${prometheus_app_name}"

    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/apps/${collectd_app_name}"


    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}"

    deleteLogicalCloud

    delete_resource_nox "${base_url_orchestrator}/projects/${projectname}"

    print_msg "deleteOrchestratorData done"
}


function createClmData {
    print_msg "Creating cluster provider and cluster"
    call_api -d "${clusterproviderdata}" "${base_url_clm}/cluster-providers"

    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata" -F "file=@$kubeconfig_path" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"

    call_api -d "${labeldata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels"


}

function deleteClmData {
    print_msg "begin deleteClmData"
    delete_resource_nox "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels/${labelname}"
    delete_resource_nox "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}"
    delete_resource_nox "${base_url_clm}/cluster-providers/${clusterprovidername}"
    print_msg "deleteClmData done"
}

function createLogicalCloudData {
    print_msg "creating logical cloud"
    call_api -d "${admin_logical_cloud_data}" "${base_url_dcm}/projects/${projectname}/logical-clouds"
    call_api -d "${cluster_1_data}" "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/cluster-references"
}

function getLogicalCloudData {
    call_api_nox "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}"
    call_api_nox "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/cluster-references/${lc_cluster_1_name}"
}

function deleteLogicalCloud {
    delete_resource_nox "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/cluster-references/${lc_cluster_1_name}"
    delete_resource_nox "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}"
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
    call_api -d "{ }" "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/instantiate"
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/approve"
    # instantiate may fail due to the logical cloud not yet instantiated, so retry
    try=0
    until call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/instantiate"; do
        if [[ $try -lt 10 ]]; then
            sleep 1s
        else
            return 1
        fi
        try=$((try + 1))
    done
    return 0
}

function terminateOrchData {
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/terminate"
    call_api -d "{ }" "${base_url_dcm}/projects/${projectname}/logical-clouds/${admin_logical_cloud_name}/terminate"
}

function status {
    call_api "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/status"
}

function waitFor {
    wait_for_deployment_status "${base_url_orchestrator}/projects/${projectname}/composite-apps/${collection_compositeapp_name}/${compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/status" $1
}

# Setup

function setupEmcoTest {
    install_deps
    populate_CSAR_composite_app_helm "$csar_id"
}

function start {
    setupEmcoTest
    deleteData
    print_msg "Before creating, deleting the data success"
    createData
    print_msg "creating the data success"
    instantiate
    print_msg "instantiate success"
    waitFor "Instantiated"
}

function stop {
    terminateOrchData
    print_msg "terminated the resources"
    waitFor "Terminated"
    deleteData
    print_msg "deleting the data success"
}

function usage {
    echo ""
    echo "    Usage: $0  start | stop"
    echo ""
    echo "    start - creates the orchstrator and cluster management data, instantiates the resources for collectd and prometheus and then deploys them on the local cluster"
    echo ""
    echo "    stop  - terminates the resources for collectd and prometheus and uninstalls the compositeApp"
    echo ""
    exit
}

if [[ "$#" -gt 0 ]] ; then
    case "$1" in
        "setup" ) setupEmcoTest ;;
        "start" ) start ;;
        "stop" ) stop ;;
        "create" ) createData ;;
        "instantiate" ) instantiate ;;
        "status" ) status ;;
        "terminate" ) terminateOrchData ;;
        "delete" ) deleteData ;;
        *) usage ;;
    esac
else
    start
    stop
fi
