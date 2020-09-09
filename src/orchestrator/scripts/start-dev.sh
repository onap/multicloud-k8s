#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright 2020 Intel Corporation.
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

source _functions.sh
k8s_path="$(git rev-parse --show-toplevel)"
#
# Start from compiled binaries to foreground. This is usable for development use.
#
source /etc/environment
opath="$(git rev-parse --show-toplevel)"/src/orchestrator

stop_all
start_mongo
start_etcd

echo "Compiling source code"
pushd $opath
generate_config
make all
./orchestrator
popd
