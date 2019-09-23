#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

modprobe -r iavf
kver=`uname -a | awk '{print $3}'`
rm -rf /lib/modules/$kver/updates/drivers/net/ethernet/intel/iavf/iavf.ko
depmod -a
sudo rm -rf /tmp/sriov
sudo rm -rf iavf-3.7.34.tar.gz
