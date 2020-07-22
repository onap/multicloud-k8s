#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# Script name: ./test_controllers.sh
# Purpose: to determine whether or not cost based, HPA, traffic, and OVN controllers work.

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

# Setup
install_deps

# Clean up
delete_all

#testing costBased placement controller
print_msg "Adding CostBased placement controller"
CostBasedIntent=""
payload="$(cat <<EOF
{
   "metadata":{
      "name":"${CostBasedIntent}",
      "description":"${CostBasedIntentName}",
      "userData1":"${userData1}",
      "userData2":"${userData2}"
   },
   "spec":{
      "host": "${CostBasedHostName}",
      "port": ${CostBasedPort},
      "type": "placement",
      "priority": 3
   }
}
EOF
)"

# Test-1
# register null CostBasedIntent
call_api_negative -d "${payload}" "${base_url}/controllers"
if [ $return_status == 405 ] ;then
    print_msg "Test:cost_based_controller-post-1 with null cost_based_controller name. Expected = 405, Actual = $return_status PASSED"
else
    print_msg "Test:cost_based_controller-post1 with null cost_based_controller name. Expected = 405, Actual = $return_status FAILED"
fi

# Test-2
# delete null CostBasedIntent
CostBasedIntent=""
print_msg "Deleting controller ${CostBasedIntent}"
delete_resource_negative "${base_url}/controllers/${CostBasedIntent}"
if [ $return_status == 400 ] ;then
    print_msg "Test:cost_based_controller-delete-1 with null project name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:cost_based_controller-delete-1 with null project name. Expected = 400, Actual = $return_status FAILED"
fi

# Test-3
# get null CostBasedIntent
CostBasedIntent=""
print_msg "getting controller ${CostBasedIntent}"
get_resource_negative "${base_url}/controllers/${CostBasedIntent}"
if [ $return_status == 404 ] ;then
    print_msg "Test:cost_based_controller-get-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:cost_based_controller-get-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi

#END
