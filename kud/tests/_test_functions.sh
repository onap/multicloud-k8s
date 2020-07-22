#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# Additional functions to run negative tests
# Aditya Sharoff <aditya.sharoff@intel.com> 07/16/2020

set -o errexit
set -o nounset
set -o pipefail

FUNCTIONS_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

source /etc/environment
source $FUNCTIONS_DIR/_common_test.sh

function call_api_negative {
    #Runs curl with passed flags and provides
    #additional error handling and debug information

    #Function outputs server response body
    #and performs validation of http_code

    local status
    local curl_response_file="$(mktemp -p /tmp)"
    local curl_common_flags=(-s -w "%{http_code}" -o "${curl_response_file}")
    local command=(curl "${curl_common_flags[@]}" "$@")

    echo "[INFO] Running '${command[@]}'" >&2
    if ! status="$("${command[@]}")"; then
        echo "[ERROR] Internal curl error! '$status'" >&2
        cat "${curl_response_file}"
        rm "${curl_response_file}"
        return 2
    else
        echo "[INFO] Server replied with status: ${status}" >&2
        cat "${curl_response_file}"
        rm "${curl_response_file}"
	return_status=$status
    fi
}


function delete_resource_negative {
    #Issues DELETE http call to provided endpoint
    #and further validates by following GET request

    call_api_negative -X DELETE "$1"
    #! call_api -X GET "$1" >/dev/null
}

function get_resource_negative {
    #! call_api_negative -X GET "$1" >/dev/null
    ! call_api_negative -X GET "$1" 
    echo $return_status
}
