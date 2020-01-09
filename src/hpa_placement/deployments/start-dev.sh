#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright 2020 Intel Corporation
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

source _functions.sh

#
# Start hpa_placement from compiled binaries to foreground. This is usable for development use.
#
source /etc/environment
k8s_path="$(git rev-parse --show-toplevel)"

stop_all
start_mongo

echo "Compiling source code"
pushd $k8s_path/src/hpa_placement/
generate_config
make all
./hpa_placement
popd
