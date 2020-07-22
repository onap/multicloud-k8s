#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# Script name: ./test_composite_app.sh
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

# Clean up
delete_all

# Register project
create_project

# TEST-1 
# Register a null composite app
composite_app_name=""
composite_app_description="test_project_description"
composite_app_version="test_composite_app_version"
userData1="user1"
userData2="user2"

print_msg "Registering composite-app"
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${composite_app_name}",
    "description": "${composite_app_description}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   },
   "spec":{
      "version":"${composite_app_version}"
   }
}
EOF
)"
call_api_negative -d "${payload}" "${base_url}/projects/${project_name}/composite-apps"
if [ $return_status == 405 ] ;then
print_msg "Test: composite application post expected value = 405 actual value = $return_status PASSED"
else
print_msg "Test: composite application post expected value = 405 actual value = $return_status FAILED"
fi

# TEST-2
# Delete a non existing composite app
composite_app_name=""
print_msg "Deleting ${composite_app_name}/${composite_app_version}"
delete_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}"
if [ $return_status == 400 ] ;then
print_msg "Test: composite application delete expected value = 400 actual value = $return_status PASSED"
else
print_msg "Test: composite application delete expected value = 400 actual value = $return_status FAILED"
fi

# Test-3
# Delete a non existing composite app
composite_app_name=""
print_msg "Deleting ${composite_app_name}/${composite_app_version}"
get_resource_negative "${base_url}/projects/${project_name}/composite-apps/${composite_app_name}/${composite_app_version}"
if [ $return_status == 400 ] ;then
print_msg "Test: composite application get expected value = 400 actual value = $return_status PASSED"
else
print_msg "Test: composite application get expected value = 400 actual value = $return_status FAILED"
fi

#END
