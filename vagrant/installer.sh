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
set -o pipefail

# _install_go() - Install GoLang package
function _install_go {
    version=$(grep "go_version" ${krd_playbooks}/krd-vars.yml | awk -F "'" '{print $2}')
    local tarball=go$version.linux-amd64.tar.gz

    if $(go version &>/dev/null); then
        return
    fi

    wget https://dl.google.com/go/$tarball
    tar -C /usr/local -xzf $tarball
    rm $tarball

    export PATH=$PATH:/usr/local/go/bin
    sed -i "s|^PATH=.*|PATH=\"$PATH\"|" /etc/environment
    export INSTALL_DIRECTORY=/usr/local/bin
    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
}

# _install_pip() - Install Python Package Manager
function _install_pip {
    if $(pip --version &>/dev/null); then
        return
    fi
    apt-get install -y python-dev
    curl -sL https://bootstrap.pypa.io/get-pip.py | python
    pip install --upgrade pip
}

# _install_ansible() - Install and Configure Ansible program
function _install_ansible {
    mkdir -p /etc/ansible/
    if $(ansible --version &>/dev/null); then
        return
    fi
    _install_pip
    pip install ansible
}

# _install_docker() - Download and install docker-engine
function _install_docker {
    local max_concurrent_downloads=${1:-3}

    if $(docker version &>/dev/null); then
        return
    fi
    apt-get install -y software-properties-common linux-image-extra-$(uname -r) linux-image-extra-virtual apt-transport-https ca-certificates curl
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
    add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
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
    echo "DOCKER_OPTS=\"-H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock --max-concurrent-downloads $max_concurrent_downloads \"" | tee --append /etc/default/docker
    usermod -aG docker $USER

    systemctl restart docker
    sleep 10
}

# install_k8s() - Install Kubernetes using kubespray tool
function install_k8s {
    echo "Deploying kubernetes"
    local dest_folder=/opt
    version=$(grep "kubespray_version" ${krd_playbooks}/krd-vars.yml | awk -F ': ' '{print $2}')
    local tarball=v$version.tar.gz

    apt-get install -y sshpass
    _install_ansible
    wget https://github.com/kubernetes-incubator/kubespray/archive/$tarball
    tar -C $dest_folder -xzf $tarball
    mv $dest_folder/kubespray-$version/ansible.cfg /etc/ansible/ansible.cfg
    rm $tarball

    pip install -r $dest_folder/kubespray-$version/requirements.txt
    rm -f $krd_inventory_folder/group_vars/all.yml 2> /dev/null
    if [[ -n "${verbose}" ]]; then
        echo "kube_log_level: 5" | tee $krd_inventory_folder/group_vars/all.yml
    else
        echo "kube_log_level: 2" | tee $krd_inventory_folder/group_vars/all.yml
    fi
    if [[ -n "${http_proxy}" ]]; then
        echo "http_proxy: \"$http_proxy\"" | tee --append $krd_inventory_folder/group_vars/all.yml
    fi
    if [[ -n "${https_proxy}" ]]; then
        echo "https_proxy: \"$https_proxy\"" | tee --append $krd_inventory_folder/group_vars/all.yml
    fi
    ansible-playbook $verbose -i $krd_inventory $dest_folder/kubespray-$version/cluster.yml -b | tee $log_folder/setup-kubernetes.log

    # Configure environment
    mkdir -p $HOME/.kube
    mv $krd_inventory_folder/artifacts/admin.conf $HOME/.kube/config
}

# install_addons() - Install Kubenertes AddOns
function install_addons {
    echo "Installing Kubernetes AddOns"
    _install_ansible
    ansible-galaxy install $verbose -r $krd_folder/galaxy-requirements.yml --ignore-errors

    ansible-playbook $verbose -i $krd_inventory $krd_playbooks/configure-krd.yml | tee $log_folder/setup-krd.log
    for addon in ${KRD_ADDONS:-virtlet ovn-kubernetes multus}; do
        echo "Deploying $addon using configure-$addon.yml playbook.."
        ansible-playbook $verbose -i $krd_inventory $krd_playbooks/configure-${addon}.yml | tee $log_folder/setup-${addon}.log
        if [[ "${testing_enabled}" == "true" ]]; then
            pushd $krd_tests
            bash ${addon}.sh
            popd
        fi
    done
}

# install_plugin() - Install ONAP Multicloud Kubernetes plugin
function install_plugin {
    echo "Installing multicloud/k8s plugin"
    _install_go
    _install_docker
    pip install docker-compose

    mkdir -p /opt/{kubeconfig,consul/config}
    cp $HOME/.kube/config /opt/kubeconfig/krd
    export KUBE_CONFIG_DIR=/opt/kubeconfig
    echo "export KUBE_CONFIG_DIR=${KUBE_CONFIG_DIR}" >> /etc/environment

    GOPATH=$(go env GOPATH)
    pushd $GOPATH/src/k8-plugin-multicloud/deployments
    ./build.sh

    if [[ "${testing_enabled}" == "true" ]]; then
        docker-compose up -d
        pushd $krd_tests
        for functional_test in plugin plugin_edgex; do
            bash ${functional_test}.sh
        done
        popd
    fi
    popd
}

# _print_kubernetes_info() - Prints the login Kubernetes information
function _print_kubernetes_info {
    if ! $(kubectl version &>/dev/null); then
        return
    fi
    # Expose Dashboard using NodePort
    node_port=30080
    KUBE_EDITOR="sed -i \"s|type\: ClusterIP|type\: NodePort|g\"" kubectl -n kube-system edit service kubernetes-dashboard
    KUBE_EDITOR="sed -i \"s|nodePort\: .*|nodePort\: $node_port|g\"" kubectl -n kube-system edit service kubernetes-dashboard

    master_ip=$(kubectl cluster-info | grep "Kubernetes master" | awk -F ":" '{print $2}')

    printf "Kubernetes Info\n===============\n" > $k8s_info_file
    echo "Dashboard URL: https:$master_ip:$node_port" >> $k8s_info_file
    echo "Admin user: kube" >> $k8s_info_file
    echo "Admin password: secret" >> $k8s_info_file
}

if ! sudo -n "true"; then
    echo ""
    echo "passwordless sudo is needed for '$(id -nu)' user."
    echo "Please fix your /etc/sudoers file. You likely want an"
    echo "entry like the following one..."
    echo ""
    echo "$(id -nu) ALL=(ALL) NOPASSWD: ALL"
    exit 1
fi

if [[ -n "${KRD_DEBUG}" ]]; then
    set -o xtrace
    verbose="-vvv"
fi

# Configuration values
log_folder=/var/log/krd
krd_folder=$(pwd)
krd_inventory_folder=$krd_folder/inventory
krd_inventory=$krd_inventory_folder/hosts.ini
krd_playbooks=$krd_folder/playbooks
krd_tests=$krd_folder/tests
k8s_info_file=$krd_folder/k8s_info.log
testing_enabled=${KRD_ENABLE_TESTS:-false}

mkdir -p $log_folder
mkdir -p /opt/csar
export CSAR_DIR=/opt/csar
echo "export CSAR_DIR=${CSAR_DIR}" | tee --append /etc/environment

# Install dependencies
# Setup proxy variables
if [ -f $krd_folder/sources.list ]; then
    mv /etc/apt/sources.list /etc/apt/sources.list.backup
    cp $krd_folder/sources.list /etc/apt/sources.list
fi
apt-get update
install_k8s
install_addons
if [[ "${KRD_PLUGIN_ENABLED:-false}" ]]; then
    install_plugin
fi
_print_kubernetes_info
