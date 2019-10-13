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
set -ex

INSTALLER_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

function install_prerequisites {
#install package for docker images
    apt-get update
    apt-get install -y curl vim wget git \
        software-properties-common python-pip
    add-apt-repository ppa:longsleep/golang-backports
    apt-get update
    apt-get install -y golang-go rsync
}

# _install_ansible() - Install and Configure Ansible program
function _install_ansible {
    local version=$(grep "ansible_version" ${kud_playbooks}/kud-vars.yml | \
        awk -F ': ' '{print $2}')
    mkdir -p /etc/ansible/
    pip install ansible==$version
}

# install_k8s() - Install Kubernetes using kubespray tool
function install_kubespray {
    echo "Deploying kubernetes"
    version=$(grep "kubespray_version" ${kud_playbooks}/kud-vars.yml | \
        awk -F ': ' '{print $2}')
    local_release_dir=$(grep "local_release_dir" \
        $kud_inventory_folder/group_vars/k8s-cluster.yml | \
        awk -F "\"" '{print $2}')
    local tarball=v$version.tar.gz
    # install make to run mitogen target & unzip is mitogen playbook dependency
    apt-get install -y sshpass make unzip
    _install_ansible
    wget https://github.com/kubernetes-incubator/kubespray/archive/$tarball
    tar -C $dest_folder -xzf $tarball
    mv $dest_folder/kubespray-$version/ansible.cfg /etc/ansible/ansible.cfg
    chown -R root:root $dest_folder/kubespray-$version
    mkdir -p ${local_release_dir}/containers
    rm $tarball

    pushd $dest_folder/kubespray-$version/
    pip install -r ./requirements.txt
    make mitogen
    popd
    rm -f $kud_inventory_folder/group_vars/all.yml 2> /dev/null
    if [[ -n "${verbose:-}" ]]; then
        echo "kube_log_level: 5" | tee \
            $kud_inventory_folder/group_vars/all.yml
    else
        echo "kube_log_level: 2" | tee \
            $kud_inventory_folder/group_vars/all.yml
    fi
    echo "kubeadm_enabled: true" | \
        tee --append $kud_inventory_folder/group_vars/all.yml
    if [[ -n "${http_proxy:-}" ]]; then
        echo "http_proxy: \"$http_proxy\"" | tee --append \
            $kud_inventory_folder/group_vars/all.yml
    fi
    if [[ -n "${https_proxy:-}" ]]; then
        echo "https_proxy: \"$https_proxy\"" | tee --append \
            $kud_inventory_folder/group_vars/all.yml
    fi
}

function install_k8s {
    version=$(grep "kubespray_version" ${kud_playbooks}/kud-vars.yml | \
        awk -F ': ' '{print $2}')
    local cluster_name=$1
    ansible-playbook $verbose -i \
        $kud_inventory $dest_folder/kubespray-$version/cluster.yml \
        -e cluster_name=$cluster_name --become --become-user=root | \
        tee $cluster_log/setup-kubernetes.log

    # Configure environment
    mkdir -p $HOME/.kube
    cp $kud_inventory_folder/artifacts/admin.conf $HOME/.kube/config
    # Copy Kubespray kubectl to be usable in host running Ansible.
    # Requires kubectl_localhost: true in inventory/group_vars/k8s-cluster.yml
    if !(which kubectl); then
        cp $kud_inventory_folder/artifacts/kubectl /usr/local/bin/
    fi

    cp -rf $kud_inventory_folder/artifacts \
        /opt/kud/multi-cluster/$cluster_name/
}

# install_addons() - Install Kubenertes AddOns
function install_addons {
    source /etc/environment
    echo "Installing Kubernetes AddOns"
    ansible-galaxy install $verbose -r \
        $kud_infra_folder/galaxy-requirements.yml --ignore-errors

    ansible-playbook $verbose -i \
        $kud_inventory $kud_playbooks/configure-kud.yml | \
        tee $cluster_log/setup-kud.log
    for addon in ${KUD_ADDONS:-virtlet ovn4nfv nfd}; do
        echo "Deploying $addon using configure-$addon.yml playbook.."
        ansible-playbook $verbose -i \
            $kud_inventory $kud_playbooks/configure-${addon}.yml | \
            tee $cluster_log/setup-${addon}.log
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
    mkdir -p /opt/{kubeconfig,consul/config}
    cp $HOME/.kube/config /opt/kubeconfig/kud

    pushd $kud_folder/../../../deployments
    ./build.sh
    if [[ "${testing_enabled}" == "true" ]]; then
        ./start.sh
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
    KUBE_EDITOR="sed -i \"s|type\: ClusterIP|type\: NodePort|g\"" \
        kubectl -n kube-system edit service kubernetes-dashboard
    KUBE_EDITOR="sed -i \"s|nodePort\: .*|nodePort\: $node_port|g\"" \
        kubectl -n kube-system edit service kubernetes-dashboard

    master_ip=$(kubectl cluster-info | grep "Kubernetes master" | \
        awk -F ":" '{print $2}')

    printf "Kubernetes Info\n===============\n" > $k8s_info_file
    echo "Dashboard URL: https:$master_ip:$node_port" >> $k8s_info_file
    echo "Admin user: kube" >> $k8s_info_file
    echo "Admin password: secret" >> $k8s_info_file
}

verbose=""
if [[ -n "${KUD_DEBUG:-}" ]]; then
    set -o xtrace
    verbose="-vvv"
fi

# Configuration values
dest_folder=/opt
kud_folder=${INSTALLER_DIR}
kud_infra_folder=$kud_folder/../../deployment_infra
kud_playbooks=$kud_infra_folder/playbooks
kud_tests=$kud_folder/../../tests
k8s_info_file=$kud_folder/k8s_info.log
testing_enabled=${KUD_ENABLE_TESTS:-false}

mkdir -p /opt/csar

function install_pkg {
# Install dependencies
    apt-get update
    install_prerequisites
    install_kubespray
}

function install_cluster {
    install_k8s $1
    install_addons
    echo "installed the addons"
    if ${KUD_PLUGIN_ENABLED:-false}; then
        install_plugin
        echo "installed the install_plugin"
    fi
    _print_kubernetes_info
}


if [ "$1" == "--install_pkg" ]; then
    export kud_inventory_folder=$kud_folder/inventory
    kud_inventory=$kud_inventory_folder/hosts.ini
    install_pkg
    exit 0
fi

if [ "$1" == "--cluster" ]; then
    cluster_name=$2
    kud_multi_cluster_path=/opt/kud/multi-cluster
    cluster_path=$kud_multi_cluster_path/$cluster_name
    cluster_log=$kud_multi_cluster_path/$cluster_name/log
    export kud_inventory_folder=$kud_folder/inventory/$cluster_name
    kud_inventory=$kud_inventory_folder/hosts.ini

    mkdir -p $kud_inventory_folder
    mkdir -p $cluster_log
    cp $kud_multi_cluster_path/$cluster_name/hosts.ini $kud_inventory_folder/
    cp -rf $kud_folder/inventory/group_vars $kud_inventory_folder/

    install_cluster $cluster_name
    exit 0
fi
