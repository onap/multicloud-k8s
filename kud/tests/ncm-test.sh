#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

source _functions.sh

base_url_clm=${base_url:-"http://localhost:9019/v2"}
base_url_ncm=${base_url:-"http://localhost:9016/v2"}

# ncm data samples
clusterprovidername="cluster-provider-a"
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

clustername="cluster-a"
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
defaultkubeconfig="$(cat<<EOF
{
    "not a good kube config file"
}
EOF
)"
echo "$defaultkubeconfig" > /tmp/ncmkubeconfig

kubeconfigfile=${kubeconfigfile:-"/tmp/ncmkubeconfig"}

labelname="LabelA"
labeldata="$(cat<<EOF
{"label-name": "$labelname"}
EOF
)"

kvname="kva"
kvdata="$(cat<<EOF
{
  "metadata": {
    "name": "$kvname",
    "description": "this is key value $kvname",
    "userData1": "cluster $kvname pair data",
    "userData2": "cluster $kvname pair data a"
  },
  "spec": {
      "kv" : [
          {"keyA": "value A"},
          {"keyB": "value B"},
          {"keyC": "value C"}
       ]
   }
}
EOF
)"

networkname="network-a"
networkdata="$(cat<<EOF
{
  "metadata": {
    "name": "$networkname",
    "description": "Description of $networkname",
    "userData1": "$networkname data part A",
    "userData2": "$networkname data part B"
  },
  "spec": {
      "cniType": "ovn4nfv",
      "ipv4Subnets": [
          {
              "subnet": "172.16.33.0/24",
              "name": "subnet3",
              "gateway":  "172.16.33.1/32",
              "excludeIps": "172.16.33.2 172.16.33.5..172.16.33.10"
          },
          {
              "subnet": "172.16.34.0/24",
              "name": "subnet4",
              "gateway":  "172.16.34.1/32",
              "excludeIps": "172.16.34.2 172.16.34.5..172.16.34.10"
          }
      ]
  }
}
EOF
)"

providernetworkname="providernetwork-a"
providernetworkdata="$(cat<<EOF
{
  "metadata": {
    "name": "$providernetworkname",
    "description": "Description of $providernetworkname",
    "userData1": "$providernetworkname data part A",
    "userData2": "$providernetworkname data part B"
  },
  "spec": {
      "cniType": "ovn4nfv",
      "ipv4Subnets": [
          {
              "subnet": "172.16.31.0/24",
              "name": "subnet1",
              "gateway":  "172.16.31.1/32",
              "excludeIps": "172.16.31.2 172.16.31.5..172.16.31.10"
          },
          {
              "subnet": "172.16.32.0/24",
              "name": "subnet2",
              "gateway":  "172.16.32.1/32",
              "excludeIps": "172.16.32.2 172.16.32.5..172.16.32.10"
          }
      ],
      "providerNetType": "VLAN",
      "vlan": {
          "vlanId": "100",
          "providerInterfaceName": "eth1",
          "logicalInterfaceName": "eth1.100",
          "vlanNodeSelector": "specific",
          "nodeLabelList": [
              "kubernetes.io/hostname=localhost",
              "kubernetes.io/name=localhost"
          ]
      }
  }
}
EOF
)"


function createNcmData {
    call_api -d "${clusterproviderdata}" "${base_url_clm}/cluster-providers"
    call_api -H "Content-Type: multipart/form-data" -F "metadata=$clusterdata" -F "file=@$kubeconfigfile" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters"
    call_api -d "${labeldata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels"
    call_api -d "${kvdata}" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/kv-pairs"
    call_api -d "${providernetworkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks"
    call_api -d "${networkdata}" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/networks"
}

# apply the network and providernetwork to an appcontext and instantiate with resource synchronizer (when implemented)
function applyNcmData {
    call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/apply"
}

# deletes the appcontext (eventually will terminate from resource synchronizer when that funcationality is ready)
function terminateNcmData {
    call_api -d "{ }" "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/terminate"
}

function getNcmData {
    call_api "${base_url_clm}/cluster-providers/${clusterprovidername}" | jq .
    call_api -H "Accept: application/json" "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}" | jq .
    call_api "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters?label=${labelname}" | jq .
    call_api "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/kv-pairs/${kvname}" | jq .
    call_api "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/networks/${networkname}" | jq .
    call_api "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks/${providernetworkname}" | jq .

}

function deleteNcmData {
    call_api -X DELETE "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/networks/${networkname}"
    call_api -X DELETE "${base_url_ncm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/provider-networks/${providernetworkname}"
    call_api -X DELETE "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/labels/${labelname}"
    call_api -X DELETE "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}/kv-pairs/${kvname}"
    call_api -X DELETE "${base_url_clm}/cluster-providers/${clusterprovidername}/clusters/${clustername}"
    call_api -X DELETE "${base_url_clm}/cluster-providers/${clusterprovidername}"
}

function usage {
    echo "Usage: $0  create|apply|get|terminate|delete"
    exit
}

# Put in logic to select from a few choices:  create, get, delete
if [ "$#" -ne 1 ] ; then
    usage
fi

case "$1" in
    "create" ) createNcmData ;;
    "apply" ) applyNcmData ;;
    "terminate" ) terminateNcmData ;;
    "get" )    getNcmData ;;
    "delete" ) deleteNcmData ;;
    *) usage ;;
esac
