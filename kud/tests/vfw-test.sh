#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

source _functions.sh

base_url_clm=${base_url_clm:-"http://10.10.10.6:31856/v2"}
base_url_ncm=${base_url_ncm:-"http://10.10.10.6:32737/v2"}
base_url_orchestrator=${base_url_orchestrator:-"http://10.10.10.6:31298/v2"}
base_url_ovnaction=${base_url_ovnaction:-"http://10.10.10.6:31181/v2"}

# add clusters to clm
# TODO one is added by default, add more if vfw demo is
#      extended to multiple clusters
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

clustername="edge01"
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

# set $kubeconfigfile before running script to point to the desired config file
kubeconfigfile=${kubeconfigfile:-"oops"}

# TODO consider demo of cluster label based placement
#      could use to onboard multiple clusters for vfw
#      but still deploy to just 1 cluster based on label
labelname="LabelA"
labeldata="$(cat<<EOF
{"label-name": "$labelname"}
EOF
)"

clustername2="edge02"
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

# set $kubeconfigfile2 before running script to point to the desired config file
kubeconfigfile2=${kubeconfigfile2:-"oops"}

# TODO consider demo of cluster label based placement
#      could use to onboard multiple clusters for vfw
#      but still deploy to just 1 cluster based on label
labelname2="LabelA"
labeldata2="$(cat<<EOF
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
    "host": "${rsynccontrollername}",
    "port": 9041 
  }
}
EOF
)"

# add the rsync controller entry
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
    "host": "${ovnactioncontrollername}",
    "type": "action",
    "priority": 1,
    "port": 9053 
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
packetgen_helm_chart=${packetgen_helm_path:-"oops"}
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
firewall_helm_chart=${firewall_helm_path:-"oops"}
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
sink_helm_chart=${sink_helm_path:-"oops"}
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
packetgen_profile_file=${packetgen_profile_targz:-"oops"}
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
firewall_profile_file=${firewall_profile_targz:-"oops"}
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
sink_profile_file=${sink_profile_targz:-"oops"}
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
      "logical-cloud":"unused_logical_cloud",
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
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent"

    call_api -d "${packetgen_workload_intent_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents"
    call_api -d "${firewall_workload_intent_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents"
    call_api -d "${sink_workload_intent_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents"

    call_api -d "${packetgen_emco_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces"
    call_api -d "${packetgen_unprotected_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces"

    call_api -d "${firewall_emco_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces"
    call_api -d "${firewall_unprotected_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces"
    call_api -d "${firewall_protected_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces"

    call_api -d "${sink_emco_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces"
    call_api -d "${sink_protected_interface_data}" \
             "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces"
}

function createOrchData {
    print_msg "creating controller entries"
    call_api -d "${rsynccontrollerdata}" "${base_url_orchestrator}/controllers"
    call_api -d "${ovnactioncontrollerdata}" "${base_url_orchestrator}/controllers"

    print_msg "creating project entry"
    call_api -d "${projectdata}" "${base_url_orchestrator}/projects"

    print_msg "creating vfw composite app entry"
    call_api -d "${vfw_compositeapp_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps"

    print_msg "adding vfw apps to the composite app"
    call_api -F "metadata=${packetgen_app_data}" \
             -F "file=@${packetgen_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps"
    call_api -F "metadata=${firewall_app_data}" \
             -F "file=@${firewall_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps"
    call_api -F "metadata=${sink_app_data}" \
             -F "file=@${sink_helm_chart}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps"

    print_msg "creating vfw composite profile entry"
    call_api -d "${vfw_composite_profile_data}" "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles"

    print_msg "adding vfw app profiles to the composite profile"
    call_api -F "metadata=${packetgen_profile_data}" \
             -F "file=@${packetgen_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles"
    call_api -F "metadata=${firewall_profile_data}" \
             -F "file=@${firewall_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles"
    call_api -F "metadata=${sink_profile_data}" \
             -F "file=@${sink_profile_file}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles"

    print_msg "create the deployment intent group"
    call_api -d "${deployment_intent_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups"
    call_api -d "${deployment_intents_in_group_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents"

    print_msg "create the generic placement intent"
    call_api -d "${generic_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents"

    print_msg "add the vfw app placement intents to the generic placement intent"
    call_api -d "${packetgen_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents"
    call_api -d "${firewall_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents"
    call_api -d "${sink_placement_intent_data}" \
             "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents"

    createOvnactionData

}

function createNcmData {
    print_msg "Creating cluster provider and cluster"
    call_api -d "${clusterproviderdata}" "${base_url_clm}/cluster-providers"
    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata" -F "file=@$kubeconfigfile" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"
    call_api -d "${labeldata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels"
    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata2" -F "file=@$kubeconfigfile2" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"
    call_api -d "${labeldata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/labels"

    print_msg "Creating provider network and network intents"
    call_api -d "${emcoprovidernetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks"
    call_api -d "${unprotectedprovidernetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks"
    call_api -d "${protectednetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/networks"

    call_api -d "${emcoprovidernetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/provider-networks"
    call_api -d "${unprotectedprovidernetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/provider-networks"
    call_api -d "${protectednetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/networks"
}


function createData {
    createNcmData
    createOrchData  # this will call createOvnactionData
}

function getOvnactionData {
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}"

    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}"

    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces/${packetgen_emco_interface_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces/${packetgen_unprotected_interface_name}"

    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_emco_interface_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_unprotected_interface_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_protected_interface_name}"

    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces/${sink_emco_interface_name}"
    call_api_nox "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces/${sink_protected_interface_name}"
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

    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}"

    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${packetgen_placement_intent_name}"
    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${firewall_placement_intent_name}"
    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${sink_placement_intent_name}"

    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}"
    call_api_nox "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents/${deployment_intents_in_group_name}"
}

function getNcmData {
    call_api_nox "${base_url_clm}/cluster-providers/${clusterprovidername}"
    call_api_nox -H "Accept: application/json" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}"
    call_api_nox -H "Accept: application/json" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}"
    call_api_nox "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels/${labelname}"
    call_api_nox "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters?label=${labelname}"

    call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks/${emcoprovidernetworkname}"
    call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks/${unprotectedprovidernetworkname}"
    call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/networks/${protectednetworkname}"

    call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/provider-networks/${emcoprovidernetworkname}"
    call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/provider-networks/${unprotectedprovidernetworkname}"
    call_api_nox "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/networks/${protectednetworkname}"
}

function getData {
    getNcmData
    getOrchData
    getOvnactionData
}

function deleteOvnactionData {
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces/${sink_protected_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}/interfaces/${sink_emco_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_protected_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_unprotected_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}/interfaces/${firewall_emco_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces/${packetgen_unprotected_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}/interfaces/${packetgen_emco_interface_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${sink_workload_intent_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${firewall_workload_intent_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}/workload-intents/${packetgen_workload_intent_name}"
    delete_resource "${base_url_ovnaction}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/network-controller-intent/${vfw_ovnaction_intent_name}"
}

function deleteOrchData {
    delete_resource "${base_url_orchestrator}/controllers/${rsynccontrollername}"
    delete_resource "${base_url_orchestrator}/controllers/${ovnactioncontrollername}"


    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${sink_placement_intent_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${firewall_placement_intent_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}/app-intents/${packetgen_placement_intent_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/generic-placement-intents/${generic_placement_intent_name}"

    deleteOvnactionData

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/intents/${deployment_intents_in_group_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${sink_profile_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${firewall_profile_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}/profiles/${packetgen_profile_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/composite-profiles/${vfw_composite_profile_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${sink_app_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${firewall_app_name}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/apps/${packetgen_app_name}"

    delete_resource "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}"
    delete_resource "${base_url_orchestrator}/projects/${projectname}"
}

function deleteNcmData {
    delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/networks/${protectednetworkname}"
    delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks/${unprotectedprovidernetworkname}"
    delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks/${emcoprovidernetworkname}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels/${labelname}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}"
    delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/networks/${protectednetworkname}"
    delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/provider-networks/${unprotectedprovidernetworkname}"
    delete_resource "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/provider-networks/${emcoprovidernetworkname}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/labels/${labelname}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}"
    delete_resource "${base_url_clm}/cluster-providers/${clusterprovidername}"
}

function deleteData {
    deleteNcmData
    deleteOrchData
}

# apply the network and providernetwork to an appcontext and instantiate with rsync
function applyNcmData {
    call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/apply"
    call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/apply"
}

# deletes the network resources from the clusters and the associated appcontext entries
function terminateNcmData {
    call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/terminate"
    call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername2}/terminate"
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
    call_api "${base_url_orchestrator}/projects/${projectname}/composite-apps/${vfw_compositeapp_name}/${vfw_compositeapp_version}/deployment-intent-groups/${deployment_intent_group_name}/status${query}"
}

function usage {
    echo "Usage: $0  create|get|delete|apply|terminate|instantiate"
    echo "    create - creates all ncm, ovnaction, clm resources needed for vfw"
    echo "             following env variables need to be set for create:"
    echo "                 kubeconfigfile=<path of kubeconfig file for destination cluster>"
    echo "                 kubeconfigfile2=<path of kubeconfig file for second destination cluster>"
    echo "                 packetgen_helm_path=<path to helm chart file for the packet generator>"
    echo "                 firewall_helm_path=<path to helm chart file for the firewall>"
    echo "                 sink_helm_path=<path to helm chart file for the sink>"
    echo "                 packetgen_profile_targz=<path to profile tar.gz file for the packet generator>"
    echo "                 firewall_profile_targz=<path to profile tar.gz file for the firewall>"
    echo "                 sink_profile_targz=<path to profile tar.gz file for the sink>"
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

    exit
}

function check_for_env_settings {
    ok=""
    if [ "${kubeconfigfile}" == "oops" ] ; then
        echo -e "ERROR - kubeconfigfile environment variable needs to be set"
        ok="no"
    fi
    if [ "${kubeconfigfile2}" == "oops" ] ; then
        echo -e "ERROR - kubeconfigfile2 environment variable needs to be set"
        ok="no"
    fi
    if [ "${packetgen_helm_chart}" == "oops" ] ; then
        echo -e "ERROR - packetgen_helm_path environment variable needs to be set"
        ok="no"
    fi
    if [ "${firewall_helm_chart}" == "oops" ] ; then
        echo -e "ERROR - firewall_helm_path environment variable needs to be set"
        ok="no"
    fi
    if [ "${sink_helm_chart}" == "oops" ] ; then
        echo -e "ERROR - sink_helm_path environment variable needs to be set"
        ok="no"
    fi
    if [ "${packetgen_profile_file}" == "oops" ] ; then
        echo -e "ERROR - packetgen_profile_targz environment variable needs to be set"
        ok="no"
    fi
    if [ "${firewall_profile_file}" == "oops" ] ; then
        echo -e "ERROR - firewall_profile_targz environment variable needs to be set"
        ok="no"
    fi
    if [ "${sink_profile_file}" == "oops" ] ; then
        echo -e "ERROR - sink_profile_targz environment variable needs to be set"
        ok="no"
    fi
    if [ "${ok}" == "no" ] ; then
        echo ""
        usage
    fi
}

if [ "$#" -lt 1 ] ; then
    usage
fi

case "$1" in
    "create" )
        check_for_env_settings
        createData
        ;;
    "get" )    getData ;;
    "delete" ) deleteData ;;
    "apply" ) applyNcmData ;;
    "instantiate" ) instantiateVfw ;;
    "terminate" ) terminateVfw ;;
    "status" )
    query=""
    if [ "$#" -eq 2 ] ; then
        query="?$2"
    fi
    statusVfw ${query} ;;
    *) usage ;;
esac
