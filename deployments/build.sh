#!/bin/bash

# Copyright 2018 Intel Corporation.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o nounset
set -o pipefail
set -o xtrace

function generate_binary {
    GOPATH=$(go env GOPATH)
    rm -f k8plugin
    rm -f *.so
    $GOPATH/bin/dep ensure -v
    for plugin in deployment namespace service; do
        go build -buildmode=plugin -o ./$plugin.so ../src/k8splugin/plugins/$plugin/plugin.go
    done
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w' -o ./k8plugin ../src/k8splugin/cmd/main.go
}

function build_image {
    echo "Start build docker image."
    docker-compose build --no-cache
}

generate_binary
build_image
