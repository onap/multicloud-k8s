#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright 2020 Intel Corporation.
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

function stop_all {
    docker-compose kill
    docker-compose down
}

function start_mongo {
    docker-compose up -d mongo
    export DATABASE_IP=$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' $(docker ps -aqf "name=mongo"))
    export no_proxy=${no_proxy:-},${DATABASE_IP}
    export NO_PROXY=${NO_PROXY:-},${DATABASE_IP}
}

function start_etcd {
    docker-compose up -d etcd
}


function generate_config {
cat << EOF > config.json
{
    "ca-file": "ca.cert",
    "server-cert": "server.cert",
    "server-key": "server.key",
    "password": "",
    "database-ip": "${DATABASE_IP}",
    "database-type": "mongo",
    "plugin-dir": "plugins",
    "etcd-ip": "127.0.0.1",
    "etcd-cert": "",
    "etcd-key": "",
    "etcd-ca-file": "",
    "service-port": "9015"
}
EOF
}

function start_all {
    docker-compose up -d
}
