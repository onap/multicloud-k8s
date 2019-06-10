#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

aio_dir=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd)
cd ${aio_dir}/../vagrant

echo "Preparing inventory for ansible"
cat <<EOL > inventory/hosts.ini
[all]
localhost

[kube-master]
localhost

[kube-node]
localhost

[etcd]
localhost

[ovn-central]
localhost

[ovn-controller]
localhost

[virtlet]
localhost

[k8s-cluster:children]
kube-node
kube-master
EOL

if ! [ -f ~/.ssh/id_rsa ]; then
        echo "Generating rsa key for this host"
        ssh-keygen -t rsa -N "" -f ~/.ssh/id_rsa <&-
fi
if ! grep -qF "$(ssh-keygen -y -f ~/.ssh/id_rsa)" ~/.ssh/authorized_keys; then
        echo "Allowing present ~/.ssh/id_rsa key to be used for login to this host"
        ssh-keygen -y -f ~/.ssh/id_rsa >> ~/.ssh/authorized_keys
fi
chmod og-wx ~/.ssh/authorized_keys

echo "Enabling nested-virtualization"
sudo ./node.sh

echo "Deploying KUD project"
./installer.sh | tee kud_installer.log
