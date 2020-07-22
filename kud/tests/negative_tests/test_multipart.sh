#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# Script name: ./test_multipart.sh
# Purpose: To ascertain whether or not POST/DELETE/GET API is able to register a null name
# Description, userdata1, and userdata2 have values that I assigned

set -o errexit
set -o nounset
set -o pipefail

source _test_functions.sh

if [ ${1:+1} ]; then
    if [ "$1" == "--external" ]; then
        master_ip=$(kubectl cluster-info | grep "Kubernetes master" | \
            awk -F ":" '{print $2}' | awk -F "//" '{print $2}')
        onap_svc_node_port=30498
        base_url="http://$master_ip:$onap_svc_node_port/v1"
    fi
fi

# Cleanup
delete_all

# Register project
create_project

# Register composite-app
create_composite_app

# Create entries for app1&app2 in the database
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

# Test-1 
# registering null app1_helm_path
app1_helm_path=""

call_api_negative -F "metadata=$payload" \
        -F "file=@$app1_helm_path" \
    "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps"
if [ $return_status == 400 ] ;then
    print_msg "Test:multipart-post-1 with null name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:multipart pos-1 with null name. Expected = 400, Actual = $return_status FAILED"
fi

# Test-2
# deleting a null app name
app1_name=""
print_msg "Deleting ${app1_name}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app1_name}"
if [ $return_status == 400 ] ;then
    print_msg "Test:multipart delete-1 with null name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:multipart delete-1 with null name. Expected = 400, Actual = $return_status FAILED"
fi

# Test-3
# geting a null app name
app1_name=""
print_msg "getting ${app1_name}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}/apps/${app1_name}"
if [ $return_status == 400 ] ;then
    print_msg "Test:multipart get-1 with null name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:multipart get-1 with null name. Expected = 400, Actual = $return_status FAILED"
fi

#END
