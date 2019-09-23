#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o nounset
set -o pipefail

ethernet_adpator_version=$( lspci | grep "Ethernet Controller X710" | head -n 1 | cut -d " " -f 8 )
SRIOV_ENABLED=${ethernet_adpator_version:-"false"}
#checking for the right hardware version of NIC on the machine
if [ "$ethernet_adpator_version" == "X710" ]; then
    SRIOV_ENABLED=true
    echo "NIC card specs match. SRIOV option avaiable for this version"
else
    echo "Failed. The version supplied does not match."
    exit 1
fi
