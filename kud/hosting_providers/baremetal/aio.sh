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

# For aio inventory by default get ovn central ip from local host default interface.
# This variable used only in this file, but env variable defined to enable user to override it prior calling aio.sh.
OVN_CENTRAL_IP_ADDRESS=${OVN_CENTRAL_IP_ADDRESS:-$(hostname -I | cut -d ' ' -f 1)}

cat <<EOL > inventory/hosts.ini
[all]
localhost ansible_ssh_host=${OVN_CENTRAL_IP_ADDRESS} ansible_ssh_port=22

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

rm -f ~/.ssh/id_rsa
echo -e "\n\n\n" | ssh-keygen -t rsa -N ""
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
chmod og-wx ~/.ssh/authorized_keys

echo "Enabling nested-virtualization"
sudo ./node.sh

echo "Deploying KUD project"
./installer.sh | tee kud_installer.log
