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
rm -rf k8s
git clone https://git.onap.org/multicloud/k8s/
cd k8s/kud/hosting_providers/vagrant/
cat <<EOL > inventory/hosts.ini
[all]
localhost ansible_ssh_host=10.10.110.21 ansible_ssh_port=22
# The ansible_ssh_host IP is an example here. Please update the ansible_ssh_host IP accordingly.

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
rm -f ~/.ssh/id_rsa
echo -e "\n\n\n" | ssh-keygen -t rsa -N ""
cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
chmod og-wx ~/.ssh/authorized_keys

echo "Enabling nested-virtualization"
./node.sh

echo "Deploying KUD project"
./installer.sh | tee kud_installer.log
