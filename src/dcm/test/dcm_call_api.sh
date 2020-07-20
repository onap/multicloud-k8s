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
project="test-project"
description="test-description"
logical_cloud_name="lc1"
cluster_1_name="lc-cl-1"
cluster_2_name="lc-cl-2"
logical_cloud_url="$dcm_addr/v2/projects/${project}/logical-clouds"
quota_name="quota-1"
quota_url="${logical_cloud_url}/${logical_cloud_name}/cluster-quotas"
cluster_url="${logical_cloud_url}/${logical_cloud_name}/cluster-references"


logical_cloud_data="$(cat << EOF
{
 "metadata" : {
    "name": "${logical_cloud_name}",
    "description": "${test-description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },
 "spec" : {
    "namespace" : "ns-1",
    "user" : {
    "user-name" : "user-1",
    "type" : "certificate",
    "user-permissions" : [
       { "permission-name" : "permission-1",
         "apiGroups" : ["stable.example.com"],
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
    "name": "${cluster_1_name}",
    "description": "${test-description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },

 "spec" : {
    "cluster-provider": "cp-1",
    "cluster-name": "c1",
    "loadbalancer-ip" : "0.0.0.0"
  }
}
EOF
)"

cluster_2_data="$(cat << EOF
{
 "metadata" : {
    "name": "${cluster_2_name}",
    "description": "${test-description}",
    "userData1":"<user data>",
    "userData2":"<user data>"
   },

 "spec" : {
    "cluster-provider": "cp-1",
    "cluster-name": "c2",
    "loadbalancer-ip" : "0.0.0.1"
  }
}
EOF
)"

quota_data="$(cat << EOF
{
    "metadata" : {
      "name" : "${quota_name}",
      "description": "${test-description}"
     },
    "spec" : {
      "limits.cpu": "400",
      "limits.memory": "1000Gi",
      "requests.cpu": "300",
      "requests.memory": "900Gi",
      "requests.storage" : "500Gi",
      "requests.ephemeral-storage": "",
      "limits.ephemeral-storage": "",
      "persistentvolumeclaims" : " ",
      "pods": "500",
      "configmaps" : "",
      "replicationcontrollers": "",
      "resourcequotas" : "",
      "services": "",
      "services.loadbalancers" : "",
      "services.nodeports" : "",
      "secrets" : "",
      "count/replicationcontrollers" : "",
      "count/deployments.apps" : "",
      "count/replicasets.apps" : "",
      "count/statefulsets.apps" : "",
      "count/jobs.batch" : "",
      "count/cronjobs.batch" : "",
      "count/deployments.extensions" : ""
    }
}
EOF
)"


# Create logical cloud
printf "\n\nCreating logical cloud data\n\n"
curl -d "${logical_cloud_data}" -X POST ${logical_cloud_url}

# Associate two clusters with the logical cloud
printf "\n\nAdding two clusters to logical cloud\n\n"3yy
curl -d "${cluster_1_data}" -X POST ${cluster_url}
curl -d "${cluster_2_data}" -X POST ${cluster_url}

# Add resource quota for the logical cloud
printf "\n\nAdding resource quota for the logical cloud\n\n"
curl -d "${quota_data}" -X POST ${quota_url}


# Get logical cloud data
printf "\n\nGetting logical cloud\n\n"
curl -X GET "${logical_cloud_url}/${logical_cloud_name}"

printf "\n\nGetting clusters info for logical cloud\n\n"
curl -X GET ${cluster_url}

printf "\n\nGetting first cluster of logical cloud\n"
curl -X GET ${cluster_url}/${cluster_1_name}

printf "\n\nGetting second cluster of logical cloud\n"
curl -X GET ${cluster_url}/${cluster_2_name}

printf "\n\nGetting Quota info for the logical cloud\n\n"
curl -X GET "${quota_url}/${quota_name}"

# Cleanup (delete created resources)
if [ "$1" == "clean" ]; then
    printf "\n\nDeleting Quota info for the logical cloud\n\n"
    curl -X DELETE "${quota_url}/${quota_name}"

    printf "\n\nDeleting the two clusters from logical cloud\n\n"3yy
    curl -X DELETE ${cluster_url}/${cluster_1_name}
    curl -X DELETE ${cluster_url}/${cluster_2_name}

    printf "\n\nDeleting logical cloud data\n\n"
    curl -X DELETE ${logical_cloud_url}/${logical_cloud_name}
fi
