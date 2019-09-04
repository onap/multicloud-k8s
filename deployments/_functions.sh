#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright 2019 © Samsung Electronics Co., Ltd.
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

function generate_k8sconfig {
cat << EOF > k8sconfig.json
{
    "database-address": "${DATABASE_IP}",
    "database-type": "mongo",
    "plugin-dir": "plugins",
    "service-port": "9015"
}
EOF
}

function start_all {
    docker-compose up -d
}
