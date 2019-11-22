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
set -o nounset
set -o pipefail

echo "Create pods ..."
kubectl apply -f sdwan/ovn-pod.yml
kubectl apply -f sdwan/sdwan-openwrt-ovn.yml

bash sdwan/test.sh

echo "Clear pods ..."
kubectl delete -f sdwan/ovn-pod.yml
kubectl delete -f sdwan/sdwan-openwrt-ovn.yml

echo "Test Completed!"
