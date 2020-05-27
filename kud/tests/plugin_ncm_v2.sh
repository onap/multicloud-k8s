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

base_url=${base_url:-"http://localhost:9019/v2"}

kubeconfig_path="$HOME/.kube/config"

cluster_provider_name1="cluster_provider1"
cluster_provider_name2="cluster_provider2"
cluster_provider_desc1="cluster_provider1_Desc"
cluster_provider_desc2="cluster_provider2_Desc"
userData1="user1"
userData2="user2"

clusterName1="clusterName1"
cluster_desc1="cluster_desc1"
clusterName2="clusterName2"
cluster_desc2="cluster_desc2"
#clusterName3 and clusterName4 shall be added with clusterLabel1 and clusterLabel2
# under cluster_provider1 and cluster_provider2 respectively
clusterName3="clusterName3"
cluster_desc3="cluster_desc3"
clusterName4="clusterName4"
cluster_desc4="cluster_desc4"
clusterName5="clusterName5"
cluster_desc5="cluster_desc5"
clusterName6="clusterName6"
cluster_desc6="cluster_desc6"

clusterLabel1="clusterLabel1"
clusterLabel2="clusterLabel2"

# BEGIN :: Delete statements are issued so that we clean up the 'cluster' collection
# and freshly populate the documents, also it serves as a direct test
# for all our DELETE APIs and an indirect test for all GET APIs
print_msg "Deleting the clusterLabel1 and clusterLabel2, if they were existing"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name1}/clusters/${clusterName3}/labels/${clusterLabel1}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name2}/clusters/${clusterName4}/labels/${clusterLabel2}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name1}/clusters/${clusterName5}/labels/${clusterLabel1}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name2}/clusters/${clusterName6}/labels/${clusterLabel2}"
# Above statements delete the clusterLabel1 and clusterLabel2 which are linked to cluster3 and cluster4

print_msg "Deleting the cluster1, cluster2, cluster3, cluster4 if they were existing"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name1}/clusters/${clusterName1}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name2}/clusters/${clusterName2}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name1}/clusters/${clusterName3}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name2}/clusters/${clusterName4}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name1}/clusters/${clusterName5}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name2}/clusters/${clusterName6}"

print_msg "Deleting the cluster-providers, if they were existing"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name1}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name2}"

# END :: Delete statements are issued so that we clean up the 'cluster' collection
# and freshly populate the documents, also it serves as a direct test
# for all our DELETE APIs and an indirect test for all GET APIs

# BEGIN: Register cluster_provider_name1 and cluster_provider_name2
print_msg "Deleting the cluster-providers, if they were existing"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name1}"
delete_resource "${base_url}/cluster-providers/${cluster_provider_name2}"

print_msg "Registering cluster_provider_name1"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${cluster_provider_name1}",
    "description": "${cluster_provider_desc1}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/cluster-providers"

print_msg "Registering cluster_provider_name2"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${cluster_provider_name2}",
    "description": "${cluster_provider_desc2}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -d "${payload}" "${base_url}/cluster-providers"
# END: Register cluster_provider_name1 and cluster_provider_name2

# BEGIN : Register cluster1, cluster2, cluster3 and cluster4
print_msg "Registering cluster1"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${clusterName1}",
    "description": "${cluster_desc1}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -F "metadata=$payload" \
         -F "file=@$kubeconfig_path" \
         "${base_url}/cluster-providers/${cluster_provider_name1}/clusters" >/dev/null #massive output


print_msg "Registering cluster2"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${clusterName2}",
    "description": "${cluster_desc2}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -F "metadata=$payload" \
         -F "file=@$kubeconfig_path" \
         "${base_url}/cluster-providers/${cluster_provider_name2}/clusters" >/dev/null #massive output


print_msg "Registering cluster3"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${clusterName3}",
    "description": "${cluster_desc3}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -F "metadata=$payload" \
         -F "file=@$kubeconfig_path" \
         "${base_url}/cluster-providers/${cluster_provider_name1}/clusters" >/dev/null #massive output


print_msg "Registering cluster4"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${clusterName4}",
    "description": "${cluster_desc4}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -F "metadata=$payload" \
         -F "file=@$kubeconfig_path" \
         "${base_url}/cluster-providers/${cluster_provider_name2}/clusters" >/dev/null #massive output

print_msg "Registering cluster5"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${clusterName5}",
    "description": "${cluster_desc5}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -F "metadata=$payload" \
         -F "file=@$kubeconfig_path" \
         "${base_url}/cluster-providers/${cluster_provider_name1}/clusters" >/dev/null #massive output


print_msg "Registering cluster6"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${clusterName6}",
    "description": "${cluster_desc6}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api -F "metadata=$payload" \
         -F "file=@$kubeconfig_path" \
         "${base_url}/cluster-providers/${cluster_provider_name2}/clusters" >/dev/null #massive output

# END : Register cluster1, cluster2, cluster3 and cluster4


# BEGIN: adding labels to cluster3 and cluster4
print_msg "Adding label to cluster3"
payload="$(cat <<EOF
{
  "label-name" : "${clusterLabel1}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/cluster-providers/${cluster_provider_name1}/clusters/${clusterName3}/labels"

print_msg "Adding label to cluster4"
payload="$(cat <<EOF
{
  "label-name" : "${clusterLabel2}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/cluster-providers/${cluster_provider_name2}/clusters/${clusterName4}/labels"

# BEGIN: adding labels to cluster5 and cluster6. Cluster5 to label1 and cluster6 to label2
print_msg "Adding label to cluster5"
payload="$(cat <<EOF
{
  "label-name" : "${clusterLabel1}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/cluster-providers/${cluster_provider_name1}/clusters/${clusterName5}/labels"

print_msg "Adding label to cluster6"
payload="$(cat <<EOF
{
  "label-name" : "${clusterLabel2}"
}
EOF
)"
call_api -d "${payload}" "${base_url}/cluster-providers/${cluster_provider_name2}/clusters/${clusterName6}/labels"