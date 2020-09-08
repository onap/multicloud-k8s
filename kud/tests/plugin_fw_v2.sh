#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

source _common_test.sh
source _functions.sh
source _functions.sh

kubeconfig_path="$HOME/.kube/config"

clusters="${KUD_PLUGIN_FW_CLUSTERS:-$(cat <<EOF
[
  {
    "metadata": {
      "name": "edge01",
      "description": "description of edge01",
      "userData1": "edge01 user data 1",
      "userData2": "edge01 user data 2"
    },
    "file": "$kubeconfig_path"
  }
]
EOF
)}"

function cluster_names {
    echo $clusters | jq -e -r '.[].metadata.name'
}

function cluster_metadata {
    cat<<EOF | jq .
{
  "metadata": $(echo $clusters | jq -e -r --arg name "$1" '.[]|select(.metadata.name==$name)|.metadata')
}
EOF
}

function cluster_file {
    echo $clusters | jq -e -r --arg name "$1" '.[]|select(.metadata.name==$name)|.file'
}

ARGS=()
while [[ $# -gt 0 ]]; do
    arg="$1"

    case $arg in
        "--external" )
            master_ip=$(kubectl cluster-info | grep "Kubernetes master" | \
                awk -F ":" '{print $2}' | awk -F "//" '{print $2}')
            base_url_clm=${base_url_clm:-"http://$master_ip:30461/v2"}
            base_url_ncm=${base_url_ncm:-"http://$master_ip:30431/v2"}
            base_url_orchestrator=${base_url_orchestrator:-"http://$master_ip:30415/v2"}
            base_url_ovnaction=${base_url_ovnaction:-"http://$master_ip:30471/v2"}
            rsync_service_port=30441
            rsync_service_host="$master_ip"
            ovnaction_service_port=30473
            ovnaction_service_host="$master_ip"
            shift
            ;;
        * )
            ARGS+=("$1")
            shift
            ;;
    esac
done
set -- "${ARGS[@]}" # restore positional parameters

base_url_clm=${base_url_clm:-"http://localhost:9061/v2"}
base_url_ncm=${base_url_ncm:-"http://localhost:9031/v2"}
base_url_orchestrator=${base_url_orchestrator:-"http://localhost:9015/v2"}
base_url_ovnaction=${base_url_ovnaction:-"http://localhost:9053/v2"}
rsync_service_port=${rsync_service_port:-9041}
rsync_service_host=${rsync_service_host:-"localhost"}
ovnaction_service_port=${ovnaction_service_port:-9053}
ovnaction_service_host=${ovnaction_service_host:-"localhost"}

CSAR_DIR="/opt/csar"
csar_id="4bf66240-a0be-4ce2-aebd-a01df7725f16"

packetgen_helm_path="$CSAR_DIR/$csar_id/packetgen.tar.gz"
packetgen_profile_targz="$CSAR_DIR/$csar_id/profile.tar.gz"
firewall_helm_path="$CSAR_DIR/$csar_id/firewall.tar.gz"
firewall_profile_targz="$CSAR_DIR/$csar_id/profile.tar.gz"
sink_helm_path="$CSAR_DIR/$csar_id/sink.tar.gz"
sink_profile_targz="$CSAR_DIR/$csar_id/profile.tar.gz"

demo_folder=$test_folder/../demo

function populate_CSAR_compositevfw_helm {
    _checks_args "$1"
    pushd "${CSAR_DIR}/$1"
    print_msg "Create Helm Chart Archives for compositevfw"
    rm -f *.tar.gz
    tar -czf packetgen.tar.gz -C $demo_folder/composite-firewall packetgen
    tar -czf firewall.tar.gz -C $demo_folder/composite-firewall firewall
    tar -czf sink.tar.gz -C $demo_folder/composite-firewall sink
    tar -czf profile.tar.gz -C $demo_folder/composite-firewall manifest.yaml override_values.yaml
    popd
}

function setup {
    install_deps
    populate_CSAR_compositevfw_helm "$csar_id"
}

clusterprovidername="vfw-cluster-provider"
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

labelname="LabelA"
labeldata="$(cat<<EOF
{"label-name": "$labelname"}
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
    "host": "${rsync_service_host}",
    "port": ${rsync_service_port}
  }
}
EOF
)"

# add the ovn action controller entry
ovnactioncontrollername="ovnaction"
ovnactioncontrollerdata="$(cat<<EOF
{
  "metadata": {
    "name": "$ovnactioncontrollername",
    "description": "description of $ovnactioncontrollername controller",
    "userData1": "user data 2 for $ovnactioncontrollername",
    "userData2": "user data 2 for $ovnactioncontrollername"
  },
  "spec": {
    "host": "${ovnaction_service_host}",
    "type": "action",
    "priority": 1,
    "port": ${ovnaction_service_port}
  }
}
EOF
)"

# define networks and providernetworks intents to ncm for the clusters
#      define emco-private-net and unprotexted-private-net as provider networks

emcoprovidernetworkname="emco-private-net"
emcoprovidernetworkdata="$(cat<<EOF
{
  "metadata": {
    "name": "$emcoprovidernetworkname",
    "description": "description of $emcoprovidernetworkname",
    "userData1": "user data 1 for $emcoprovidernetworkname",
    "userData2": "user data 2 for $emcoprovidernetworkname"
  },
  "spec": {
      "cniType": "ovn4nfv",
      "ipv4Subnets": [
          {
              "subnet": "10.10.20.0/24",
              "name": "subnet1",
              "gateway":  "10.10.20.1/24"
          }
      ],
      "providerNetType": "VLAN",
      "vlan": {
          "vlanId": "102",
          "providerInterfaceName": "eth1",
          "logicalInterfaceName": "eth1.102",
          "vlanNodeSelector": "specific",
          "nodeLabelList": [
              "kubernetes.io/hostname=localhost"
          ]
      }
  }
}
EOF
)"

unprotectedprovidernetworkname="unprotected-private-net"
unprotectedprovidernetworkdata="$(cat<<EOF
{
  "metadata": {
    "name": "$unprotectedprovidernetworkname",
    "description": "description of $unprotectedprovidernetworkname",
    "userData1": "user data 2 for $unprotectedprovidernetworkname",
    "userData2": "user data 2 for $unprotectedprovidernetworkname"
  },
  "spec": {
      "cniType": "ovn4nfv",
      "ipv4Subnets": [
          {
              "subnet": "192.168.10.0/24",
              "name": "subnet1",
              "gateway":  "192.168.10.1/24"
          }
      ],
      "providerNetType": "VLAN",
      "vlan": {
          "vlanId": "100",
          "providerInterfaceName": "eth1",
          "logicalInterfaceName": "eth1.100",
          "vlanNodeSelector": "specific",
          "nodeLabelList": [
              "kubernetes.io/hostname=localhost"
          ]
      }
  }
}
EOF
)"

protectednetworkname="protected-private-net"
protectednetworkdata="$(cat<<EOF
{
  "metadata": {
    "name": "$protectednetworkname",
    "description": "description of $protectednetworkname",
    "userData1": "user data 1 for $protectednetworkname",
    "userData2": "user data 1 for $protectednetworkname"
  },
  "spec": {
      "cniType": "ovn4nfv",
      "ipv4Subnets": [
          {
              "subnet": "192.168.20.0/24",
              "name": "subnet1",
              "gateway":  "192.168.20.100/32"
          }
      ]
  }
}
EOF
)"

# define a project
projectname="testvfw"
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
vfw_compositeapp_name="compositevfw"
vfw_compositeapp_version="v1"
vfw_compositeapp_data="$(cat <<EOF
{
  "metadata": {
    "name": "${vfw_compositeapp_name}",
    "description": "description of ${vfw_compositeapp_name}",
    "userData1": "user data 1 for ${vfw_compositeapp_name}",
    "userData2": "user data 2 for ${vfw_compositeapp_name}"
   },
   "spec":{
      "version":"${vfw_compositeapp_version}"
   }
}
EOF
)"

# define app entries for the composite application
#   includes the multipart tgz of the helm chart for vfw
# BEGIN: Create entries for app1&app2 in the database
packetgen_app_name="packetgen"
packetgen_helm_chart=${packetgen_helm_path}
packetgen_app_data="$(cat <<EOF
{
  "metadata": {
    "name": "${packetgen_app_name}",
    "description": "description for app ${packetgen_app_name}",
    "userData1": "user data 2 for ${packetgen_app_name}",
    "userData2": "user data 2 for ${packetgen_app_name}"
   }
}
EOF
)"

firewall_app_name="firewall"
firewall_helm_chart=${firewall_helm_path}
firewall_app_data="$(cat <<EOF
{
  "metadata": {
    "name": "${firewall_app_name}",
    "description": "description for app ${firewall_app_name}",
    "userData1": "user data 2 for ${firewall_app_name}",
    "userData2": "user data 2 for ${firewall_app_name}"
   }
}
EOF
)"

sink_app_name="sink"
sink_helm_chart=${sink_helm_path}
sink_app_data="$(cat <<EOF
{
  "metadata": {
    "name": "${sink_app_name}",
    "description": "description for app ${sink_app_name}",
    "userData1": "user data 2 for ${sink_app_name}",
    "userData2": "user data 2 for ${sink_app_name}"
   }
}
EOF
)"

# Add the composite profile
vfw_composite_profile_name="vfw_composite-profile"
vfw_composite_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${vfw_composite_profile_name}",
      "description":"description of ${vfw_composite_profile_name}",
      "userData1":"user data 1 for ${vfw_composite_profile_name}",
      "userData2":"user data 2 for ${vfw_composite_profile_name}"
   }
}
EOF
)"

# define the packetgen profile data
packetgen_profile_name="packetgen-profile"
packetgen_profile_file=${packetgen_profile_targz}
packetgen_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${packetgen_profile_name}",
      "description":"description of ${packetgen_profile_name}",
      "userData1":"user data 1 for ${packetgen_profile_name}",
      "userData2":"user data 2 for ${packetgen_profile_name}"
   },
   "spec":{
      "app-name":  "${packetgen_app_name}"
   }
}
EOF
)"

# define the firewall profile data
firewall_profile_name="firewall-profile"
firewall_profile_file=${firewall_profile_targz}
firewall_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${firewall_profile_name}",
      "description":"description of ${firewall_profile_name}",
      "userData1":"user data 1 for ${firewall_profile_name}",
      "userData2":"user data 2 for ${firewall_profile_name}"
   },
   "spec":{
      "app-name":  "${firewall_app_name}"
   }
}
EOF
)"

# define the sink profile data
sink_profile_name="sink-profile"
sink_profile_file=${sink_profile_targz}
sink_profile_data="$(cat <<EOF
{
   "metadata":{
      "name":"${sink_profile_name}",
      "description":"description of ${sink_profile_name}",
      "userData1":"user data 1 for ${sink_profile_name}",
      "userData2":"user data 2 for ${sink_profile_name}"
   },
   "spec":{
      "app-name":  "${sink_app_name}"
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

# define app placement intent for packetgen
packetgen_placement_intent_name="packetgen-placement-intent"
packetgen_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${packetgen_placement_intent_name}",
      "description":"description of ${packetgen_placement_intent_name}",
      "userData1":"user data 1 for ${packetgen_placement_intent_name}",
      "userData2":"user data 2 for ${packetgen_placement_intent_name}"
   },
   "spec":{
      "app-name":"${packetgen_app_name}",
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

# define app placement intent for firewall
firewall_placement_intent_name="firewall-placement-intent"
firewall_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${firewall_placement_intent_name}",
      "description":"description of ${firewall_placement_intent_name}",
      "userData1":"user data 1 for ${firewall_placement_intent_name}",
      "userData2":"user data 2 for ${firewall_placement_intent_name}"
   },
   "spec":{
      "app-name":"${firewall_app_name}",
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

# define app placement intent for sink
sink_placement_intent_name="sink-placement-intent"
sink_placement_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${sink_placement_intent_name}",
      "description":"description of ${sink_placement_intent_name}",
      "userData1":"user data 1 for ${sink_placement_intent_name}",
      "userData2":"user data 2 for ${sink_placement_intent_name}"
   },
   "spec":{
      "app-name":"${sink_app_name}",
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
release="fw0"
deployment_intent_group_name="vfw_deployment_intent_group"
deployment_intent_group_data="$(cat <<EOF
{
   "metadata":{
      "name":"${deployment_intent_group_name}",
      "description":"descriptiont of ${deployment_intent_group_name}",
      "userData1":"user data 1 for ${deployment_intent_group_name}",
      "userData2":"user data 2 for ${deployment_intent_group_name}"
   },
   "spec":{
      "profile":"${vfw_composite_profile_name}",
      "version":"${release}",
      "override-values":[
         {
            "app-name":"${packetgen_app_name}",
            "values": {
                  ".Values.service.ports.nodePort":"30888"
               }
         },
         {
            "app-name":"${firewall_app_name}",
            "values": {
                  ".Values.global.dcaeCollectorIp":"1.2.3.4",
                  ".Values.global.dcaeCollectorPort":"8888"
               }
         },
         {
            "app-name":"${sink_app_name}",
            "values": {
                  ".Values.service.ports.nodePort":"30677"
               }
         }
      ]
   }
}
EOF
)"

# define the network-control-intent for the vfw composite app
vfw_ovnaction_intent_name="vfw_ovnaction_intent"
vfw_ovnaction_intent_data="$(cat <<EOF
{
   "metadata":{
      "name":"${vfw_ovnaction_intent_name}",
      "description":"descriptionf of ${vfw_ovnaction_intent_name}",
      "userData1":"user data 1 for ${vfw_ovnaction_intent_name}",
      "userData2":"user data 2 for ${vfw_ovnaction_intent_name}"
   }
}
EOF
)"

# define the network workload intent for packetgen app
packetgen_workload_intent_name="packetgen_workload_intent"
packetgen_workload_intent_data="$(cat <<EOF
{
  "metadata": {
    "name": "${packetgen_workload_intent_name}",
    "description": "description of ${packetgen_workload_intent_name}",
    "userData1": "useer data 2 for ${packetgen_workload_intent_name}",
    "userData2": "useer data 2 for ${packetgen_workload_intent_name}"
  },
  "spec": {
    "application-name": "${packetgen_app_name}",
    "workload-resource": "${release}-${packetgen_app_name}",
    "type": "Deployment"
  }
}
EOF
)"

# define the network workload intent for firewall app
firewall_workload_intent_name="firewall_workload_intent"
firewall_workload_intent_data="$(cat <<EOF
{
  "metadata": {
    "name": "${firewall_workload_intent_name}",
    "description": "description of ${firewall_workload_intent_name}",
    "userData1": "useer data 2 for ${firewall_workload_intent_name}",
    "userData2": "useer data 2 for ${firewall_workload_intent_name}"
  },
  "spec": {
    "application-name": "${firewall_app_name}",
    "workload-resource": "${release}-${firewall_app_name}",
    "type": "Deployment"
  }
}
EOF
)"

# define the network workload intent for sink app
sink_workload_intent_name="sink_workload_intent"
sink_workload_intent_data="$(cat <<EOF
{
  "metadata": {
    "name": "${sink_workload_intent_name}",
    "description": "description of ${sink_workload_intent_name}",
    "userData1": "useer data 2 for ${sink_workload_intent_name}",
    "userData2": "useer data 2 for ${sink_workload_intent_name}"
  },
  "spec": {
    "application-name": "${sink_app_name}",
    "workload-resource": "${release}-${sink_app_name}",
    "type": "Deployment"
  }
}
EOF
)"

# define the network interface intents for the packetgen workload intent
packetgen_unprotected_interface_name="packetgen_unprotected_if"
packetgen_unprotected_interface_data="$(cat <<EOF
{
  "metadata": {
    "name": "${packetgen_unprotected_interface_name}",
    "description": "description of ${packetgen_unprotected_interface_name}",
    "userData1": "useer data 2 for ${packetgen_unprotected_interface_name}",
    "userData2": "useer data 2 for ${packetgen_unprotected_interface_name}"
  },
  "spec": {
    "interface": "eth1",
    "name": "${unprotectedprovidernetworkname}",
    "defaultGateway": "false",
    "ipAddress": "192.168.10.2"
  }
}
EOF
)"

packetgen_emco_interface_name="packetgen_emco_if"
packetgen_emco_interface_data="$(cat <<EOF
{
  "metadata": {
    "name": "${packetgen_emco_interface_name}",
    "description": "description of ${packetgen_emco_interface_name}",
    "userData1": "useer data 2 for ${packetgen_emco_interface_name}",
    "userData2": "useer data 2 for ${packetgen_emco_interface_name}"
  },
  "spec": {
    "interface": "eth2",
    "name": "${emcoprovidernetworkname}",
    "defaultGateway": "false",
    "ipAddress": "10.10.20.2"
  }
}
EOF
)"

# define the network interface intents for the firewall workload intent
firewall_unprotected_interface_name="firewall_unprotected_if"
firewall_unprotected_interface_data="$(cat <<EOF
{
  "metadata": {
    "name": "${firewall_unprotected_interface_name}",
    "description": "description of ${firewall_unprotected_interface_name}",
    "userData1": "useer data 2 for ${firewall_unprotected_interface_name}",
    "userData2": "useer data 2 for ${firewall_unprotected_interface_name}"
  },
  "spec": {
    "interface": "eth1",
    "name": "${unprotectedprovidernetworkname}",
    "defaultGateway": "false",
    "ipAddress": "192.168.10.3"
  }
}
EOF
)"

firewall_protected_interface_name="firewall_protected_if"
firewall_protected_interface_data="$(cat <<EOF
{
  "metadata": {
    "name": "${firewall_protected_interface_name}",
    "description": "description of ${firewall_protected_interface_name}",
    "userData1": "useer data 2 for ${firewall_protected_interface_name}",
    "userData2": "useer data 2 for ${firewall_protected_interface_name}"
  },
  "spec": {
    "interface": "eth2",
    "name": "${protectednetworkname}",
    "defaultGateway": "false",
    "ipAddress": "192.168.20.2"
  }
}
EOF
)"

firewall_emco_interface_name="firewall_emco_if"
firewall_emco_interface_data="$(cat <<EOF
{
  "metadata": {
    "name": "${firewall_emco_interface_name}",
    "description": "description of ${firewall_emco_interface_name}",
    "userData1": "useer data 2 for ${firewall_emco_interface_name}",
    "userData2": "useer data 2 for ${firewall_emco_interface_name}"
  },
  "spec": {
    "interface": "eth3",
    "name": "${emcoprovidernetworkname}",
    "defaultGateway": "false",
    "ipAddress": "10.10.20.3"
  }
}
EOF
)"

# define the network interface intents for the sink workload intent
sink_protected_interface_name="sink_protected_if"
sink_protected_interface_data="$(cat <<EOF
{
  "metadata": {
    "name": "${sink_protected_interface_name}",
    "description": "description of ${sink_protected_interface_name}",
    "userData1": "useer data 2 for ${sink_protected_interface_name}",
    "userData2": "useer data 2 for ${sink_protected_interface_name}"
  },
  "spec": {
    "interface": "eth1",
    "name": "${protectednetworkname}",
    "defaultGateway": "false",
    "ipAddress": "192.168.20.3"
  }
}
EOF
)"

sink_emco_interface_name="sink_emco_if"
sink_emco_interface_data="$(cat <<EOF
{
  "metadata": {
    "name": "${sink_emco_interface_name}",
    "description": "description of ${sink_emco_interface_name}",
    "userData1": "useer data 2 for ${sink_emco_interface_name}",
    "userData2": "useer data 2 for ${sink_emco_interface_name}"
  },
  "spec": {
    "interface": "eth2",
    "name": "${emcoprovidernetworkname}",
    "defaultGateway": "false",
    "ipAddress": "10.10.20.4"
  }
}
EOF
)"

# define the intents to be used by the group
deployment_intents_in_group_name="vfw_deploy_intents"
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
         "genericPlacementIntent":"${generic_placement_intent_name}",
         "ovnaction" : "${vfw_ovnaction_intent_name}"
      }
   }
}
EOF
)"

function createOvnactionData {
    call_api -d "${vfw_ovnaction_intent_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent"

    call_api -d "${packetgen_workload_intent_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents"
    call_api -d "${firewall_workload_intent_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents"
    call_api -d "${sink_workload_intent_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents"

    call_api -d "${packetgen_emco_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces"
    call_api -d "${packetgen_unprotected_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces"

    call_api -d "${firewall_emco_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces"
    call_api -d "${firewall_unprotected_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces"
    call_api -d "${firewall_protected_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces"

    call_api -d "${sink_emco_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces"
    call_api -d "${sink_protected_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces"
}

function createOrchData {
    print_msg "Creating controller entries"
    call_api -d "${rsynccontrollerdata}" "${base_url_orchestrator}/controllers"
    call_api -d "${ovnactioncontrollerdata}" "${base_url_orchestrator}/controllers"

    print_msg "Creating project entry"
    call_api -d "${projectdata}" "${base_url_orchestrator}/projects"

    print_msg "Creating vfw composite app entry"
    call_api -d "${vfw_compositeapp_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps"

    print_msg "Adding vfw apps to the composite app"
    call_api -F "metadata=${packetgen_app_data}" \
             -F "file=@${packetgen_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps"
    call_api -F "metadata=${firewall_app_data}" \
             -F "file=@${firewall_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps"
    call_api -F "metadata=${sink_app_data}" \
             -F "file=@${sink_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps"

    print_msg "Creating vfw composite profile entry"
    call_api -d "${vfw_composite_profile_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles"

    print_msg "Adding vfw app profiles to the composite profile"
    call_api -F "metadata=${packetgen_profile_data}" \
             -F "file=@${packetgen_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles"
    call_api -F "metadata=${firewall_profile_data}" \
             -F "file=@${firewall_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles"
    call_api -F "metadata=${sink_profile_data}" \
             -F "file=@${sink_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles"

    print_msg "Create the generic placement intent"
    call_api -d "${generic_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents"

    print_msg "Add the vfw app placement intents to the generic placement intent"
    call_api -d "${packetgen_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents"
    call_api -d "${firewall_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents"
    call_api -d "${sink_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents"

    createOvnactionData

    print_msg "Create the deployment intent group"
    call_api -d "${deployment_intent_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups"
    call_api -d "${deployment_intents_in_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents"
}

function createNcmData {
    print_msg "Creating cluster provider ${clusterprovidername}"
    call_api -d "${clusterproviderdata}" "${base_url_clm}/cluster-providers"

    for name in $(cluster_names); do
        metadata=$(cluster_metadata "$name")
        file=$(cluster_file "$name")
        print_msg "Creating cluster ${name}"
        call_api -H "Content-Type: multipart/form-data" -F "metadata=$metadata" -F "file=@$file" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"
        call_api -d "${labeldata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${name}/labels"

        print_msg "Creating provider network and network intents for ${name}"
        call_api -d "${emcoprovidernetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/provider-networks"
        call_api -d "${unprotectedprovidernetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/provider-networks"
        call_api -d "${protectednetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/networks"
    done
}

function createData {
    setup
    createNcmData
    createOrchData  # this will call createOvnactionData
}

function getOvnactionData {
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}"

    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}"

    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces/${packetgen_emco_interface_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces/${packetgen_unprotected_interface_name}"

    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_emco_interface_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_unprotected_interface_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_protected_interface_name}"

    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces/${sink_emco_interface_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces/${sink_protected_interface_name}"
}

function getOrchData {
    call_api_nox "${base_url_orchestrator}/controllers/${rsynccontrollername}"
    call_api_nox "${base_url_orchestrator}/controllers/${ovnactioncontrollername}"

    call_api_nox "${base_url_orchestrator}/projects/${projectname}"
    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}"

    call_api_nox -H "Accept: application/json" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${packetgen_app_name}"
    call_api_nox -H "Accept: application/json" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${firewall_app_name}"
    call_api_nox -H "Accept: application/json" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${sink_app_name}"

    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}"

    call_api_nox -H "Accept: application/json" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${packetgen_profile_name}"
    call_api_nox -H "Accept: application/json" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${firewall_profile_name}"
    call_api_nox -H "Accept: application/json" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${sink_profile_name}"

    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}"

    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${packetgen_placement_intent_name}"
    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${firewall_placement_intent_name}"
    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${sink_placement_intent_name}"

    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}"
    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents/${deployment_intents_in_group_name}"
}

function getNcmData {
    call_api_nox "${base_url_clm}/cluster-providers/${clusterprovidername}"
    call_api_nox "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters?label=${labelname}"

    for name in $(cluster_names); do
        call_api_nox -H "Accept: application/json" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${name}"
        call_api_nox "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${name}/labels/${labelname}"
        call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/provider-networks/${emcoprovidernetworkname}"
        call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/provider-networks/${unprotectedprovidernetworkname}"
        call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/networks/${protectednetworkname}"
    done
}

function getData {
    getNcmData
    getOrchData
    getOvnactionData
}

function deleteOvnactionData {
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces/${sink_protected_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces/${sink_emco_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_protected_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_unprotected_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_emco_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces/${packetgen_unprotected_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces/${packetgen_emco_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/network-controller-intent/${vfw_ovnaction_intent_name}"
}

function deleteOrchData {
    delete_resource "${base_url_orchestrator}/controllers/${rsynccontrollername}"
    delete_resource "${base_url_orchestrator}/controllers/${ovnactioncontrollername}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents/${deployment_intents_in_group_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${sink_placement_intent_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${firewall_placement_intent_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${packetgen_placement_intent_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/generic-placement-intents/${generic_placement_intent_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${sink_profile_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${firewall_profile_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${packetgen_profile_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${sink_app_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${firewall_app_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${packetgen_app_name}"

    deleteOvnactionData

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}"
}

function deleteNcmData {
    for name in $(cluster_names); do
        delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/networks/${protectednetworkname}"
        delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/provider-networks/${unprotectedprovidernetworkname}"
        delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/provider-networks/${emcoprovidernetworkname}"
        delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${name}/labels/${labelname}"
        delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${name}"
    done

    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}"
}

function deleteData {
    deleteNcmData
    deleteOrchData
}

# apply the network and providernetwork to an appcontext and instantiate with rsync
function applyNcmData {
    for name in $(cluster_names); do
        call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/apply"
    done
}

# deletes the network resources from the clusters and the associated appcontext entries
function terminateNcmData {
    for name in $(cluster_names); do
        call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${name}/terminate"
    done
}

# terminates the vfw resources
function terminateOrchData {
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/terminate"
}

# terminates the vfw and ncm resources
function terminateVfw {
    terminateOrchData
    terminateNcmData
}

function instantiateVfw {
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/approve"
    call_api -d "{ }" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/instantiate"
}

function statusVfw {
    call_api "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/status"
}

function waitForVfw {
    wait_for_deployment_status "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/status" $1
}

function usage {
    echo "Usage: $0  create|get|delete|apply|terminate|instantiate"
    echo "    create - creates all ncm, ovnaction, clm resources needed for vfw"
    echo "    get - queries all resources in ncm, ovnaction, clm resources created for vfw"
    echo "    delete - deletes all resources in ncm, ovnaction, clm resources created for vfw"
    echo "    apply - applys the network intents - e.g. networks created in ncm"
    echo "    instantiate - approves and instantiates the composite app via the generic deployment intent"
    echo "    status - get status of deployed resources"
    echo "    terminate - remove the vFW composite app resources and network resources create by 'instantiate' and 'apply'"
    echo ""
    echo "    a reasonable test sequence:"
    echo "    1.  create"
    echo "    2.  apply"
    echo "    3.  instantiate"
    echo "    4.  status"
    echo "    5.  terminate"
    echo "    6.  destroy"

    exit
}

if [[ "$#" -gt 0 ]] ; then
    case "$1" in
        "create" ) createData ;;
        "get" ) getData ;;
        "apply" ) applyNcmData ;;
        "instantiate" ) instantiateVfw ;;
        "status" ) statusVfw ;;
        "wait" ) waitForVfw "Instantiated" ;;
        "terminate" ) terminateVfw ;;
        "delete" ) deleteData ;;
        *) usage ;;
    esac
else
    createData
    applyNcmData
    instantiateVfw

    print_msg "[BEGIN] Basic checks for instantiated resource"
    print_msg "Wait for deployment to be instantiated"
    waitForVfw "Instantiated"
    for name in $(cluster_names); do
        print_msg "Check that networks were created on cluster $name"
        file=$(cluster_file "$name")
        KUBECONFIG=$file kubectl get network protected-private-net
        KUBECONFIG=$file kubectl get providernetwork emco-private-net
        KUBECONFIG=$file kubectl get providernetwork unprotected-private-net
    done
    for name in $(cluster_names); do
        print_msg "Wait for all pods to start on cluster $name"
        file=$(cluster_file "$name")
        KUBECONFIG=$file wait_for_pod -l app=sink
        KUBECONFIG=$file wait_for_pod -l app=firewall
        KUBECONFIG=$file wait_for_pod -l app=packetgen
    done
    # TODO: Provide some health check to verify vFW work
    print_msg "Not waiting for vFW to fully install as no further checks are implemented in testcase"
    #print_msg "Waiting 8minutes for vFW installation"
    #sleep 8m
    print_msg "[END] Basic checks for instantiated resource"

    terminateVfw
    waitForVfw "Terminated"
    deleteData
fi
