#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020 Intel Corporation
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o nounset
set -o pipefail

k8s_path="$(git rev-parse --show-toplevel)"

VERSION="0.0.1-SNAPSHOT"
export IMAGE_NAME="nexus3.onap.org:10003/onap/multicloud/k8s/orchestrator"

function _compile_src {
    echo "Compiling source code"
    pushd $k8s_path/src/orchestrator/
    make
    popd
}

function _move_bin {
    echo "Moving binaries"
    mv $k8s_path/src/orchestrator/orchestrator .
}

function _cleanup {
    echo "Cleaning previous execution"
    docker-compose kill
    image=$(grep "image.*orchestrator" docker-compose.yml)
    if [[ -n ${image} ]]; then
        docker images ${image#*:} -q | xargs docker rmi -f
    fi
    docker ps -a --filter "status=exited" -q | xargs docker rm
}

function _build_docker {
    echo "Building docker image"
    docker-compose build --no-cache
}

function _push_image {
    local tag_name=${IMAGE_NAME}:${1:-latest}

    echo "Start push {$tag_name}"
    docker push ${IMAGE_NAME}:latest
    docker tag ${IMAGE_NAME}:latest ${tag_name}
    docker push ${tag_name}
}

if [[ -n "${JENKINS_HOME+x}" ]]; then
    set -o xtrace
    _compile_src
    _move_bin
    _build_docker
    _push_image $VERSION
else
    source /etc/environment

    _compile_src
    _move_bin
    _cleanup
    _build_docker
fi

