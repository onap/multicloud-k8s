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
export GOPATH=$k8s_path

export CSAR_DIR=/opt/csar
export KUBE_CONFIG_DIR=/opt/kubeconfig
export DATABASE_TYPE=consul
export DATABASE_IP=consul-svr
export PLUGINS_DIR=$k8s_path/src/k8splugin/plugins

echo "Starting consul services"
docker-compose kill
docker-compose up -d consul
echo "Compiling source code"

pushd $k8s_path/src/k8splugin/
make clean
make plugins
go run cmd/main.go
popd
