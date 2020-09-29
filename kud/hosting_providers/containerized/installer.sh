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
    echo "Removing ppa for jonathonf/python-3.6"
    ls /etc/apt/sources.list.d/ || true
    find /etc/apt/sources.list.d -maxdepth 1 -name '*jonathonf*' -delete || true
    apt-get update
    apt-get install -y curl vim wget git \
        software-properties-common python-pip sudo
    add-apt-repository -y ppa:longsleep/golang-backports
    apt-get update
    apt-get install -y golang-go rsync
}

# _install_ansible() - Install and Configure Ansible program
function _install_ansible {
    local version=$(grep "ansible_version" ${kud_playbooks}/kud-vars.yml |
        awk -F ': ' '{print $2}')
    mkdir -p /etc/ansible/
    pip install --no-cache-dir ansible==$version
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
    pip install --no-cache-dir -r ./requirements.txt
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
    if [ ${1:+1} ]; then
        local plugins_name="$1"
        echo "additional addons plugins $1"
    else
        local plugins_name=""
        echo "no additional addons pluigns"
    fi

    source /etc/environment
    echo "Installing Kubernetes AddOns"
    ansible-galaxy install $verbose -r \
        $kud_infra_folder/galaxy-requirements.yml --ignore-errors

    ansible-playbook $verbose -i \
        $kud_inventory $kud_playbooks/configure-kud.yml | \
        tee $cluster_log/setup-kud.log
    for addon in ${KUD_ADDONS:-virtlet ovn4nfv nfd sriov cmk $plugins_name}; do
        echo "Deploying $addon using configure-$addon.yml playbook.."
        ansible-playbook $verbose -i \
            $kud_inventory $kud_playbooks/configure-${addon}.yml | \
            tee $cluster_log/setup-${addon}.log
    done

    echo "Run the test cases if testing_enabled is set to true."
    if [[ "${testing_enabled}" == "true" ]]; then
        for addon in ${KUD_ADDONS:-virtlet ovn4nfv nfd sriov cmk $plugins_name}; do
            pushd $kud_tests
            bash ${addon}.sh
            case $addon in
                "onap4k8s" )
                    echo "Test the onap4k8s plugin installation"
                    for functional_test in plugin_edgex plugin_fw plugin_eaa; do
                        bash ${functional_test}.sh --external
                    done
                    ;;
                "emco" )
                    echo "Test the emco plugin installation"
                    for functional_test in plugin_fw_v2; do
                        bash ${functional_test}.sh --external
                    done
                    ;;
            esac
            popd
        done
    fi
    echo "Add-ons deployment complete..."
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
export CSAR_DIR=/opt/csar

function install_pkg {
# Install dependencies
    apt-get update
    install_prerequisites
    install_kubespray
}

function install_cluster {
    install_k8s $1
    if [ ${2:+1} ]; then
        echo "install default addons and $2"
        install_addons "$2"
    else
        install_addons
    fi
    echo "installed the addons"

    _print_kubernetes_info
}

function usage {
    echo "installer usage:"
    echo "./installer.sh --install_pkg - Install the required softwarepackage"
    echo "./installer.sh --cluster <cluster name> \
- Install k8s cluster with default plugins"
    echo "./installer.sh --cluster <cluster name> \
--plugins <plugin_1 plugin_2> - Install k8s cluster with default plugins \
and additional plugins such as onap4k8s."
}

if [ $# -eq 0 ]; then
    echo "Error: No arguments supplied"
    usage
    exit 1
fi

if [ -z "$1" ]; then
    echo "Error: Null argument passed"
    usage
    exit 1
fi

if [ "$1" == "--install_pkg" ]; then
    export kud_inventory_folder=$kud_folder/inventory
    kud_inventory=$kud_inventory_folder/hosts.ini
    install_pkg
    echo "install pkg"
    exit 0
fi

if [ "$1" == "--cluster" ]; then
    if [ -z "${2-}"  ]; then
        echo "Error: Cluster name is null"
        usage
        exit 1
    fi

    cluster_name=$2
    kud_multi_cluster_path=/opt/kud/multi-cluster
    cluster_path=$kud_multi_cluster_path/$cluster_name
    echo $cluster_path
    if [ ! -d "${cluster_path}" ]; then
        echo "Error: cluster_path ${cluster_path} doesn't exit"
        usage
        exit 1
    fi

    cluster_log=$kud_multi_cluster_path/$cluster_name/log
    export kud_inventory_folder=$kud_folder/inventory/$cluster_name
    kud_inventory=$kud_inventory_folder/hosts.ini

    mkdir -p $kud_inventory_folder
    mkdir -p $cluster_log
    cp $kud_multi_cluster_path/$cluster_name/hosts.ini $kud_inventory_folder/
    cp -rf $kud_folder/inventory/group_vars $kud_inventory_folder/

    if [ ${3:+1} ]; then
        if [ "$3" == "--plugins" ]; then
            if [ -z "${4-}"  ]; then
                echo "Error: plugins arguments is null; Refer the usage"
                usage
                exit 1
            fi
            plugins_name=${@:4:$#}
            install_cluster $cluster_name "$plugins_name"
            exit 0
        else
            echo "Error: cluster argument should have plugins; \
                Refer the usage"
            usage
            exit 1
        fi
    fi
    install_cluster $cluster_name
    exit 0
fi

echo "Error: Refer the installer usage"
usage
exit 1
