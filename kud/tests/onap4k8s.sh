#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################
set -o errexit
set -o pipefail

source _functions.sh
set +e

master_ip=$(kubectl cluster-info | grep "Kubernetes master" | \
    awk -F ":" '{print $2}' | awk -F "//" '{print $2}')
onap_svc_node_port=30498
declare -i timeout=18
declare -i interval=10

base_url="http://$master_ip:$onap_svc_node_port/v1"

function check_onap_svc {
    while ((timeout > 0)); do
        echo "try $timeout: Wait for $interval seconds to check for onap svc"
        sleep $interval
        call_api "$base_url/healthcheck"
        call_api_ret=$?
        if [[ $call_api_ret -eq 0 ]]; then
            echo "onap svc health check is success"
            exit 0
        fi
        ((timeout-=1))
    done
}

check_onap_svc
echo "Failed to check for onap svc"
exit 1
