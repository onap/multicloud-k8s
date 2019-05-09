#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright 2019 © Samsung Electronics Co., Ltd.
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

# install_deps() - Install dependencies required for functional tests
function install_deps {
    if ! $(jq --version &>/dev/null); then
        source /etc/os-release || source /usr/lib/os-release
        case ${ID,,} in
            *suse)
            ;;
            ubuntu|debian)
                sudo apt-get install -y jq
            ;;
            rhel|centos|fedora)
            ;;
        esac
    fi
}
