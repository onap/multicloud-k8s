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
        function ubuntu_deps {
            sudo apt-get install -y jq
        }
        install_packages "" ubuntu_deps ""
    fi
}

# install_ovn_deps() - Install dependencies required for tests that require OVN
function install_ovn_deps {
    if ! $(yq --version &>/dev/null); then
        install_deps # jq needed as it's dependency of yq
        sudo -E pip install yq
    fi
    if ! $(ovn-nbctl --version &>/dev/null); then
        function ovn_ubuntu_deps {
            sudo apt-get install -y apt-transport-https
            echo "deb https://packages.wand.net.nz $(lsb_release -sc) main" | sudo tee /etc/apt/sources.list.d/wand.list
            sudo curl https://packages.wand.net.nz/keyring.gpg -o /etc/apt/trusted.gpg.d/wand.gpg
            sudo apt-get update
            sudo apt install -y ovn-common
        }
        install_packages "" ovn_ubuntu_deps ""
    fi
}

function install_packages {
    local suse_packages=$1
    local ubuntu_debian_packages=$2
    local rhel_centos_packages=$3
    source /etc/os-release || source /usr/lib/os-release
    case ${ID,,} in
        *suse)
            ($suse_packages)
        ;;
        ubuntu|debian)
            ($ubuntu_debian_packages)
        ;;
        rhel|centos|fedora)
            ($rhel_centos_packages)
        ;;
    esac
}
