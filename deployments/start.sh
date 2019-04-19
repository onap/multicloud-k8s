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

source /etc/environment

k8s_path="$(git rev-parse --show-toplevel)"
export GOPATH=$k8s_path
export GO111MODULE=on
CONFIG_FILE=$(pwd)/config/smsconfig.json

echo "Starting mongo services"
docker-compose kill
docker-compose up -d mongo
export DATABASE_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -aqf "name=mongo"))
export no_proxy=${no_proxy:-},$DATABASE_IP
export NO_PROXY=${NO_PROXY:-},$DATABASE_IP

echo "Compiling source code"
pushd $k8s_path/src/k8splugin/
cat << EOF > k8sconfig.json
{
    "database-address":     "$DATABASE_IP",
    "database-type": "mongo",
    "plugin-dir": "$(pwd)/plugins",
    "kube-config-dir": "$(pwd)/kubeconfigs"
}
EOF
make all
./k8plugin
popd
