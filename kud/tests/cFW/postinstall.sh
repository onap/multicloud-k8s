#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o nounset
set -o pipefail
set -o xtrace

# install_docker() - Download and install docker-engine
function install_docker {
    local max_concurrent_downloads=${1:-3}

    if $(docker version &>/dev/null); then
        return
    fi
    apt-get install -y software-properties-common linux-image-extra-$(uname -r) linux-image-extra-virtual apt-transport-https ca-certificates curl
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu  $(lsb_release -cs) stable"
    apt-get update
    apt-get install -y docker-ce

    mkdir -p /etc/systemd/system/docker.service.d
    if [ $http_proxy ]; then
        cat <<EOL > /etc/systemd/system/docker.service.d/http-proxy.conf
[Service]
Environment="HTTP_PROXY=$http_proxy"
EOL
    fi
    if [ $https_proxy ]; then
        cat <<EOL > /etc/systemd/system/docker.service.d/https-proxy.conf
[Service]
Environment="HTTPS_PROXY=$https_proxy"
EOL
    fi
    if [ $no_proxy ]; then
        cat <<EOL > /etc/systemd/system/docker.service.d/no-proxy.conf
[Service]
Environment="NO_PROXY=$no_proxy"
EOL
    fi
    systemctl daemon-reload
    echo "DOCKER_OPTS=\"-H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock --max-concurrent-downloads $max_concurrent_downloads \"" >> /etc/default/docker
    usermod -aG docker $USER

    systemctl restart docker
    sleep 10
}

# install_docker_compose() - Installs docker compose python module
function install_docker_compose {
    if ! which pip; then
        curl -sL https://bootstrap.pypa.io/get-pip.py | python
    fi
    pip install --upgrade pip
    pip install docker-compose
}

echo 'vm.nr_hugepages = 1024' >> /etc/sysctl.conf
sysctl -p

install_docker
install_docker_compose

cd /vagrant
# build vpp docker image
BUILD_ARGS="--no-cache"
if [ $HTTP_PROXY ]; then
    BUILD_ARGS+=" --build-arg HTTP_PROXY=${HTTP_PROXY}"
fi
if [ $HTTPS_PROXY ]; then
    BUILD_ARGS+=" --build-arg HTTPS_PROXY=${HTTPS_PROXY}"
fi
pushd vpp
docker build ${BUILD_ARGS} -t electrocucaracha/vpp:latest .
popd

docker-compose up -d
