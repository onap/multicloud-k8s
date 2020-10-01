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

dcm_addr="http://localhost:9077"

# parameters
project="test-project"
description="test-description"
logical_cloud_name="lc1"
namespace="ns1"
api_groups=""
user="user-1"
permission="permission-1"
cluster_provider_name="cp-1"
cluster_1_name="c1"
cluster_2_name="c2"
lc_cluster_1_name="lc-cl-1"
lc_cluster_2_name="lc-cl-2"
quota_name="quota-1"

# endpoints
logical_cloud_url="$dcm_addr/v2/projects/${project}/logical-clouds"
quota_url="${logical_cloud_url}/${logical_cloud_name}/cluster-quotas"
cluster_url="${logical_cloud_url}/${logical_cloud_name}/cluster-references"


logical_cloud_data="$(cat << EOF
{
 "metadata" : {
    "name": "${logical_cloud_name}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "namespace" : "${namespace}",
    "user" : {
    "user-name" : "${user}",
    "type" : "certificate",
    "user-permissions" : [
       { "permission-name" : "${permission}",
         "apiGroups" : ["${api_groups}"],
         "resources" : ["secrets", "pods"],
         "verbs" : ["get", "watch", "list", "create"]
       }
    ]
  }
 }
}
EOF
)"

cluster_1_data="$(cat << EOF
{
 "metadata" : {
    "name": "${lc_cluster_1_name}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },

 "spec" : {
    "cluster-provider": "${cluster_provider_name}",
    "cluster-name": "${cluster_1_name}",
    "loadbalancer-ip" : "0.0.0.0"
  }
}
EOF
)"

cluster_2_data="$(cat << EOF
{
 "metadata" : {
    "name": "${lc_cluster_2_name}",
    "description": "${description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },

 "spec" : {
    "cluster-provider": "${cluster_provider_name}",
    "cluster-name": "${cluster_2_name}",
    "loadbalancer-ip" : "0.0.0.1"
  }
}
EOF
)"

quota_data="$(cat << EOF
{
    "metadata" : {
      "name" : "${quota_name}",
      "description": "${description}"
     },
    "spec" : {
      "limits.cpu": "400",
      "limits.memory": "1000Gi",
      "requests.cpu": "300",
      "requests.memory": "900Gi",
      "requests.storage" : "500Gi",
      "requests.ephemeral-storage": "500",
      "limits.ephemeral-storage": "500",
      "persistentvolumeclaims" : "500",
      "pods": "500",
      "configmaps" : "1000",
      "replicationcontrollers": "500",
      "resourcequotas" : "500",
      "services": "500",
      "services.loadbalancers" : "500",
      "services.nodeports" : "500",
      "secrets" : "500",
      "count/replicationcontrollers" : "500",
      "count/deployments.apps" : "500",
      "count/replicasets.apps" : "500",
      "count/statefulsets.apps" : "500",
      "count/jobs.batch" : "500",
      "count/cronjobs.batch" : "500",
      "count/deployments.extensions" : "500"
    }
}
EOF
)"

# Cleanup (delete created resources)
if [ "$1" == "clean" ]; then
    printf "\n\nTerminating logical cloud...\n\n"
    curl -X POST "${logical_cloud_url}/${logical_cloud_name}/terminate"
    sleep 10

    printf "\n\nDeleting Quota info for the logical cloud\n\n"
    curl -X DELETE "${quota_url}/${quota_name}"

    printf "\n\nDeleting the two clusters from logical cloud\n\n"
    curl -X DELETE ${cluster_url}/${lc_cluster_1_name}
    curl -X DELETE ${cluster_url}/${lc_cluster_2_name}

    printf "\n\nDeleting logical cloud data\n\n"
    curl -X DELETE ${logical_cloud_url}/${logical_cloud_name}
elif [ "$1" == "kube" ]; then
    printf "\n\nFetching kubeconfig for cluster 1:\n\n"
    curl -X GET "${logical_cloud_url}/${logical_cloud_name}/cluster-references/${lc_cluster_1_name}/kubeconfig" > kubeconfig-${lc_cluster_1_name}

    printf "\n\nFetching kubeconfig for cluster 2:\n\n"
    curl -X GET "${logical_cloud_url}/${logical_cloud_name}/cluster-references/${lc_cluster_2_name}/kubeconfig" > kubeconfig-${lc_cluster_2_name}
else
    printf "\n\nCreating logical cloud data\n\n"
    curl -d "${logical_cloud_data}" -X POST ${logical_cloud_url}

    printf "\n\nAdding two clusters to logical cloud\n\n"
    curl -d "${cluster_1_data}" -X POST ${cluster_url}
    curl -d "${cluster_2_data}" -X POST ${cluster_url}

    printf "\n\nAdding resource quota for the logical cloud\n\n"
    curl -d "${quota_data}" -X POST ${quota_url}

    printf "\n\nGetting logical cloud\n\n"
    curl -X GET "${logical_cloud_url}/${logical_cloud_name}"

    printf "\n\nGetting clusters info for logical cloud\n\n"
    curl -X GET ${cluster_url}

    printf "\n\nGetting first cluster of logical cloud\n"
    curl -X GET ${cluster_url}/${lc_cluster_1_name}

    printf "\n\nGetting second cluster of logical cloud\n"
    curl -X GET ${cluster_url}/${lc_cluster_2_name}

    printf "\n\nGetting Quota info for the logical cloud\n\n"
    curl -X GET "${quota_url}/${quota_name}"

    printf "\n\nApplying logical cloud...\n\n"
    curl -X POST "${logical_cloud_url}/${logical_cloud_name}/apply"
fi
