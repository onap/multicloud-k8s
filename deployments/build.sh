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

VERSION="0.11.0-SNAPSHOT"
GO_VERSION="1.14"
export IMAGE_NAME="nexus3.onap.org:10003/onap/multicloud/k8s"

function _compile_src {
    echo "Compiling source code"
    pushd $k8s_path/src/k8splugin/
    pwd
    # mount directory and build in container (thus not relying on the state of the runner)
    docker run --rm -v "$PWD":/usr/src/myapp -w /usr/src/myapp nexus3.onap.org:10001/golang:${GO_VERSION} make
    popd
}

function _move_bin {
    echo "Moving binaries"
    rm -f k8plugin *so
    mv $k8s_path/src/k8splugin/k8plugin .
    mv $k8s_path/src/k8splugin/plugins/*.so .
}

function _cleanup {
    echo "Cleaning previous execution"
    docker-compose kill
    image=$(grep "image.*k8plugin" docker-compose.yml)
    if [[ -n ${image} ]]; then
        docker images ${image#*:} -q | xargs docker rmi -f
    fi

    exited_containers=$(docker ps -a --filter "status=exited" -q)
    if [[ -n "$exited_containers" ]]; then
        echo "Removing exited containers..."
        echo "$exited_containers" | xargs docker rm
    else
        echo "Nothing to remove"
    fi
}

function _build_docker {
    echo "Building docker image"
    apt-get update && apt-get install -y docker-compose-plugin
    docker-compose build --no-cache
}

function _push_image {
    local timestamp=$(date -u +%Y%m%dT%H%M%SZ)
    local tag_name=${IMAGE_NAME}:${1:-latest}
    local tag_with_timestamp=${IMAGE_NAME}:${input_tag}-${timestamp}

    echo "Pushing ${tag_name}"
    docker tag ${IMAGE_NAME}:latest ${tag_name}
    docker push ${tag_name}

    docker tag ${IMAGE_NAME}:latest ${tag_with_timestamp}

    echo "Pushing ${tag_with_timestamp}"
    docker push ${tag_with_timestamp}
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
