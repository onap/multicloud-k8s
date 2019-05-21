#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright 2019 © Samsung Electronics Co., Ltd.
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
# Start k8splugin from compiled binaries to foreground. This is usable for development use.
#
source /etc/environment
k8s_path="$(git rev-parse --show-toplevel)"

stop_all
start_mongo

echo "Compiling source code"
pushd $k8s_path/src/k8splugin/
generate_k8sconfig
make all
./k8plugin
popd
