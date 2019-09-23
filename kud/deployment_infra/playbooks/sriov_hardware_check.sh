#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o pipefail

source /etc/environment

ethernet_adpator_version=$( lspci | grep "Ethernet Controller X710" | head -n 1 | cut -d " " -f 8 )
if [ -z "$ethernet_adpator_version" ]; then
    echo "False"
    exit 0
fi
SRIOV_ENABLED=${ethernet_adpator_version:-"false"}
#checking for the right hardware version of NIC on the machine
if [ "$ethernet_adpator_version" == "X710" ]; then
    echo "True"
else
    echo "False"
fi
