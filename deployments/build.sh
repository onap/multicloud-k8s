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
set -o xtrace

k8s_path="$(git rev-parse --show-toplevel)"
export GOPATH=$k8s_path

echo "Compiling source code"
pushd $k8s_path/src/k8splugin/
make
popd

rm -f k8plugin *so
mv $k8s_path/src/k8splugin/k8plugin .
mv $k8s_path/src/k8splugin/plugins/*.so .

echo "Cleaning previous execution"
docker-compose kill
image=$(grep "image.*k8plugin" docker-compose.yml)
docker images ${image#*:} -q | xargs docker rmi -f
docker ps -a --filter "status=exited" -q | xargs docker rm

echo "Starting docker building process"
docker-compose build --no-cache
