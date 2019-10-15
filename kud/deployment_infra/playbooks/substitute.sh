#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

for file in $(find /etc/*.conf -type f -name "c6xxvf_dev*.conf"); do
    device_id=$( echo $file | cut -d '_' -f 2 | tr -cd '[[:digit:]]')
    echo $device_id
    cat /etc/c6xxvf_dev${device_id}.conf
    sed -i "s/\[SSL\]/\[SSL${device_id}\]/g" /etc/c6xxvf_dev${device_id}.conf
done

for file in $(find /etc/*.conf -type f -name "c6xx_dev*.conf"); do
    dev_id=$( echo $file | cut -d '_' -f 2 | tr -cd '[[:digit:]]')
    echo $dev_id
    cat /etc/c6xx_dev${dev_id}.conf
    sed -i "s/\[SSL\]/\[SSL${dev_id}\]/g" /etc/c6xx_dev${dev_id}.conf
done
