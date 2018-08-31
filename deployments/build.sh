#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018 Intel Corporation
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o nounset
set -o pipefail

k8s_path="$(git rev-parse --show-toplevel)"

echo "Compiling source code"
pushd $k8s_path/src/k8splugin/
make
popd

pushd $k8s_path/deployments
for file in k8plugin *so; do
    rm -f $file
    mv $k8s_path/src/k8splugin/$file .
done

echo "Starting docker building process"
docker-compose build --no-cache
popd
