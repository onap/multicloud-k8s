#!/bin/bash
#SPDX-license-identifier: Apache-2.0
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

INSTALLER_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

source ${INSTALLER_DIR}/../../tests/_functions.sh

# _install_go() - Install GoLang package
function _install_go {
    version=$(grep "go_version" ${kud_playbooks}/kud-vars.yml | awk -F "'" '{print $2}')
    local tarball=go$version.linux-amd64.tar.gz

    if $(go version &>/dev/null); then
        return
    fi

    wget https://dl.google.com/go/$tarball
    sudo tar -C /usr/local -xzf $tarball
    rm $tarball

    export PATH=$PATH:/usr/local/go/bin
    sudo sed -i "s|^PATH=.*|PATH=\"$PATH\"|" /etc/environment
}

# _install_pip() - Install Python Package Manager
function _install_pip {
    if $(pip --version &>/dev/null); then
        sudo -E pip install --upgrade pip
    else
        sudo apt-get install -y python-dev
        curl -sL https://bootstrap.pypa.io/get-pip.py | sudo python
    fi
}

# _install_ansible() - Install and Configure Ansible program
function _install_ansible {
    if $(ansible --version &>/dev/null); then
        sudo pip uninstall -y ansible
    fi
    _install_pip
    local version=$(grep "ansible_version" ${kud_playbooks}/kud-vars.yml | awk -F ': ' '{print $2}')
    sudo mkdir -p /etc/ansible/
    sudo -E pip install ansible==$version
}

# _install_docker() - Download and install docker-engine
function _install_docker {
    local max_concurrent_downloads=${1:-3}

    if $(docker version &>/dev/null); then
        return
    fi
    sudo apt-get install -y apt-transport-https ca-certificates curl
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
    sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
    sudo apt-get update
    sudo apt-get install -y docker-ce

    sudo mkdir -p /etc/systemd/system/docker.service.d
    if [ ${http_proxy:-} ]; then
        echo "[Service]" | sudo tee /etc/systemd/system/docker.service.d/http-proxy.conf
        echo "Environment=\"HTTP_PROXY=$http_proxy\"" | sudo tee --append /etc/systemd/system/docker.service.d/http-proxy.conf
    fi
    if [ ${https_proxy:-} ]; then
        echo "[Service]" | sudo tee /etc/systemd/system/docker.service.d/https-proxy.conf
        echo "Environment=\"HTTPS_PROXY=$https_proxy\"" | sudo tee --append /etc/systemd/system/docker.service.d/https-proxy.conf
    fi
    if [ ${no_proxy:-} ]; then
        echo "[Service]" | sudo tee /etc/systemd/system/docker.service.d/no-proxy.conf
        echo "Environment=\"NO_PROXY=$no_proxy\"" | sudo tee --append /etc/systemd/system/docker.service.d/no-proxy.conf
    fi
    sudo systemctl daemon-reload
    echo "DOCKER_OPTS=\"-H tcp://0.0.0.0:2375 -H unix:///var/run/docker.sock --max-concurrent-downloads $max_concurrent_downloads \"" | sudo tee --append /etc/default/docker
    if [[ -z $(groups | grep docker) ]]; then
        sudo usermod -aG docker $USER
    fi

    sudo systemctl restart docker
    sleep 10
}

function _set_environment_file {
    # By default ovn central interface is the first active network interface on localhost. If other wanted, need to export this variable in aio.sh or Vagrant file.
    OVN_CENTRAL_INTERFACE="${OVN_CENTRAL_INTERFACE:-$(ip addr show | awk '/inet.*brd/{print $NF; exit}')}"
    echo "export OVN_CENTRAL_INTERFACE=${OVN_CENTRAL_INTERFACE}" | sudo tee --append /etc/environment
    echo "export OVN_CENTRAL_ADDRESS=$(get_ovn_central_address)" | sudo tee --append /etc/environment
    echo "export KUBE_CONFIG_DIR=/opt/kubeconfig" | sudo tee --append /etc/environment
    echo "export CSAR_DIR=/opt/csar" | sudo tee --append /etc/environment
}

# install_k8s() - Install Kubernetes using kubespray tool
function install_k8s {
    echo "Deploying kubernetes"
    local dest_folder=/opt
    version=$(grep "kubespray_version" ${kud_playbooks}/kud-vars.yml | awk -F ': ' '{print $2}')
    local_release_dir=$(grep "local_release_dir" $kud_inventory_folder/group_vars/k8s-cluster.yml | awk -F "\"" '{print $2}')
    local tarball=v$version.tar.gz
    sudo apt-get install -y sshpass make unzip # install make to run mitogen target and unzip is mitogen playbook dependency
    _install_docker
    _install_ansible
    wget https://github.com/kubernetes-incubator/kubespray/archive/$tarball
    sudo tar -C $dest_folder -xzf $tarball
    sudo mv $dest_folder/kubespray-$version/ansible.cfg /etc/ansible/ansible.cfg
    sudo chown -R $USER $dest_folder/kubespray-$version
    sudo mkdir -p ${local_release_dir}/containers
    rm $tarball

    pushd $dest_folder/kubespray-$version/
    sudo -E pip install -r ./requirements.txt
    make mitogen
    popd
    rm -f $kud_inventory_folder/group_vars/all.yml 2> /dev/null
    if [[ -n "${verbose:-}" ]]; then
        echo "kube_log_level: 5" | tee $kud_inventory_folder/group_vars/all.yml
    else
        echo "kube_log_level: 2" | tee $kud_inventory_folder/group_vars/all.yml
    fi
    echo "kubeadm_enabled: true" | tee --append $kud_inventory_folder/group_vars/all.yml
    if [[ -n "${http_proxy:-}" ]]; then
        echo "http_proxy: \"$http_proxy\"" | tee --append $kud_inventory_folder/group_vars/all.yml
    fi
    if [[ -n "${https_proxy:-}" ]]; then
        echo "https_proxy: \"$https_proxy\"" | tee --append $kud_inventory_folder/group_vars/all.yml
    fi
    ansible-playbook $verbose -i $kud_inventory $dest_folder/kubespray-$version/cluster.yml --become --become-user=root | sudo tee $log_folder/setup-kubernetes.log

    # Configure environment
    mkdir -p $HOME/.kube
    cp $kud_inventory_folder/artifacts/admin.conf $HOME/.kube/config
    # Copy Kubespray kubectl to be usable in host running Ansible. Requires kubectl_localhost: true in inventory/group_vars/k8s-cluster.yml
    sudo cp $kud_inventory_folder/artifacts/kubectl /usr/local/bin/
}

# install_addons() - Install Kubenertes AddOns
function install_addons {
    source /etc/environment
    echo "Installing Kubernetes AddOns"
    _install_ansible
    sudo ansible-galaxy install $verbose -r $kud_infra_folder/galaxy-requirements.yml --ignore-errors

    ansible-playbook $verbose -i $kud_inventory $kud_playbooks/configure-kud.yml | sudo tee $log_folder/setup-kud.log
    for addon in ${KUD_ADDONS:-virtlet ovn4nfv nfd}; do
        echo "Deploying $addon using configure-$addon.yml playbook.."
        ansible-playbook $verbose -i $kud_inventory $kud_playbooks/configure-${addon}.yml | sudo tee $log_folder/setup-${addon}.log
        if [[ "${testing_enabled}" == "true" ]]; then
            pushd $kud_tests
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
    sudo -E pip install docker-compose

    sudo mkdir -p /opt/{kubeconfig,consul/config}
    sudo cp $HOME/.kube/config /opt/kubeconfig/kud

    pushd $kud_folder/../../../deployments
    sudo ./build.sh
    if [[ "${testing_enabled}" == "true" ]]; then
        sudo ./start.sh
        pushd $kud_tests
        for functional_test in plugin plugin_edgex plugin_fw; do
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

sudo -k # forgot sudo password
if ! sudo -n "true"; then
    echo ""
    echo "passwordless sudo is needed for '$(id -nu)' user."
    echo "Please fix your /etc/sudoers file. You likely want an"
    echo "entry like the following one..."
    echo ""
    echo "$(id -nu) ALL=(ALL) NOPASSWD: ALL"
    exit 1
fi

verbose=""
if [[ -n "${KUD_DEBUG:-}" ]]; then
    set -o xtrace
    verbose="-vvv"
fi

# Configuration values
log_folder=/var/log/kud
kud_folder=${INSTALLER_DIR}
kud_infra_folder=$kud_folder/../../deployment_infra
export kud_inventory_folder=$kud_folder/inventory
kud_inventory=$kud_inventory_folder/hosts.ini
kud_playbooks=$kud_infra_folder/playbooks
kud_tests=$kud_folder/../../tests
k8s_info_file=$kud_folder/k8s_info.log
testing_enabled=${KUD_ENABLE_TESTS:-false}

sudo mkdir -p $log_folder
sudo mkdir -p /opt/csar
sudo chown -R $USER /opt/csar

# Install dependencies
# Setup proxy variables
if [ -f $kud_folder/sources.list ]; then
    sudo mv /etc/apt/sources.list /etc/apt/sources.list.backup
    sudo cp $kud_folder/sources.list /etc/apt/sources.list
fi
sudo apt-get update
install_k8s
_set_environment_file
install_addons
if ${KUD_PLUGIN_ENABLED:-false}; then
    install_plugin
fi
_print_kubernetes_info
