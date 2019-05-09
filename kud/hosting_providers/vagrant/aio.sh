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

if [[ $(whoami) != 'root' ]];then
    echo "This bash script must be executed as root user"
    exit 1
fi

echo "Cloning and configuring KUD project..."
if [ -d "k8s" ]; then rm -rf k8s; fi
git clone https://git.onap.org/multicloud/k8s/
cd k8s/kud/hosting_providers/baremetal/
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
sed -i '/andrewrothstein.kubectl/d' ../../deployment_infra/playbooks/configure-*.yml
if [ -f ~/.ssh/id_rsa ]; then rm -f ~/.ssh/id_rsa; fi
echo -e "\n\n\n" | ssh-keygen -t rsa -N ""
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
chmod og-wx ~/.ssh/authorized_keys

echo "Enabling nested-virtualization"
./node.sh

echo "Deploying KRD project"
./installer.sh | tee kud_installer.log
