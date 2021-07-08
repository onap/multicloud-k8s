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

    #gcc is required for go apps compilation
    if ! which gcc; then
        sudo apt-get install -y gcc
    fi

    if $(go version &>/dev/null); then
        return
    fi

    wget https://dl.google.com/go/$tarball
    sudo tar -C /usr/local -xzf $tarball
    rm $tarball

    export PATH=$PATH:/usr/local/go/bin
    sudo sed -i "s|^PATH=.*|PATH=\"$PATH\"|" /etc/environment
    #allow golang to work with sudo
    sudo sed -i 's|secure_path="\([^"]\+\)"|secure_path="\1:/usr/local/go/bin"|' /etc/sudoers
}

# _install_ansible() - Install and Configure Ansible program
function _install_ansible {
    sudo apt-get install -y python3 python3-pip
    sudo update-alternatives --install /usr/bin/python python /usr/bin/python3 1 --force
    sudo update-alternatives --install /usr/bin/pip pip /usr/bin/pip3 1 --force
    sudo -E pip install --no-cache-dir --upgrade pip
    if $(ansible --version &>/dev/null); then
        sudo pip uninstall -y ansible
    fi
    local version=$(grep "ansible_version" ${kud_playbooks}/kud-vars.yml | awk -F ': ' '{print $2}')
    sudo mkdir -p /etc/ansible/
    sudo -E pip install --no-cache-dir ansible==$version
}

function _set_environment_file {
    # By default ovn central interface is the first active network interface on localhost. If other wanted, need to export this variable in aio.sh or Vagrant file.
    OVN_CENTRAL_INTERFACE="${OVN_CENTRAL_INTERFACE:-$(ip addr show | awk '/inet.*brd/{print $NF; exit}')}"
    echo "export OVN_CENTRAL_INTERFACE=${OVN_CENTRAL_INTERFACE}" | sudo tee --append /etc/environment
    echo "export OVN_CENTRAL_ADDRESS=$(get_ovn_central_address)" | sudo tee --append /etc/environment
    echo "export KUBE_CONFIG_DIR=/opt/kubeconfig" | sudo tee --append /etc/environment
    echo "export CSAR_DIR=/opt/csar" | sudo tee --append /etc/environment
    echo "export ANSIBLE_CONFIG=${ANSIBLE_CONFIG}" | sudo tee --append /etc/environment
}

# install_k8s() - Install Kubernetes using kubespray tool
function install_k8s {
    echo "Deploying kubernetes"
    local dest_folder=/opt
    version=$(grep "kubespray_version" ${kud_playbooks}/kud-vars.yml | awk -F ': ' '{print $2}')
    local_release_dir=$(grep "local_release_dir" $kud_inventory_folder/group_vars/k8s-cluster.yml | awk -F "\"" '{print $2}')
    local tarball=v$version.tar.gz
    sudo apt-get install -y sshpass make unzip # install make to run mitogen target and unzip is mitogen playbook dependency
    sudo apt-get install -y gnupg2 software-properties-common
    _install_ansible
    wget https://github.com/kubernetes-incubator/kubespray/archive/$tarball
    sudo tar -C $dest_folder -xzf $tarball
    sudo chown -R $USER $dest_folder/kubespray-$version
    sudo mkdir -p ${local_release_dir}/containers
    rm $tarball

    pushd $dest_folder/kubespray-$version/
    sudo -E pip install --no-cache-dir -r ./requirements.txt
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
    export ANSIBLE_CONFIG=$dest_folder/kubespray-$version/ansible.cfg

    ansible-playbook $verbose -i $kud_inventory \
        $kud_playbooks/preconfigure-kubespray.yml --become --become-user=root \
        | sudo tee $log_folder/setup-kubernetes.log
    if [ "$container_runtime" == "docker" ]; then
        /bin/echo -e "\n\e[1;42mDocker will be used as the container runtime interface\e[0m"
        ansible-playbook $verbose -i $kud_inventory \
            $dest_folder/kubespray-$version/cluster.yml --become \
            --become-user=root | sudo tee $log_folder/setup-kubernetes.log
    elif [ "$container_runtime" == "containerd" ]; then
        /bin/echo -e "\n\e[1;42mContainerd will be used as the container runtime interface\e[0m"
        # Because the kud_kata_override_variable has its own quotations in it
        # a eval command is needed to properly execute the ansible script
        ansible_kubespray_cmd="ansible-playbook $verbose -i $kud_inventory \
            $dest_folder/kubespray-$version/cluster.yml \
            -e ${kud_kata_override_variables} --become --become-user=root | \
            sudo tee $log_folder/setup-kubernetes.log"
        eval $ansible_kubespray_cmd
        ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" \
            $kud_playbooks/configure-kata.yml --become --become-user=root | \
            sudo tee $log_folder/setup-kata.log
    else
        echo "Only Docker or Containerd are supported container runtimes"
        exit 1
    fi

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
    ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" $kud_playbooks/configure-kud.yml | sudo tee $log_folder/setup-kud.log

    # The order of KUD_ADDONS is important: some plugins (sriov, qat)
    # require nfd to be enabled. Some addons are not currently supported with containerd
    if [ "${container_runtime}" == "docker" ]; then
        kud_addons=${KUD_ADDONS:-virtlet ovn4nfv nfd sriov \
            qat optane cmk}
    elif [ "${container_runtime}" == "containerd" ]; then
        kud_addons=${KUD_ADDONS:-ovn4nfv nfd}
    fi

    for addon in ${kud_addons}; do
        echo "Deploying $addon using configure-$addon.yml playbook.."
        ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" \
            $kud_playbooks/configure-${addon}.yml | \
            sudo tee $log_folder/setup-${addon}.log
    done

    echo "Run the test cases if testing_enabled is set to true."
    if [[ "${testing_enabled}" == "true" ]]; then
        failed_kud_tests=""
        # Run Kata test first if Kata was installed
        if [ "${container_runtime}" == "containerd" ]; then
            #Install Kata webhook for test pods
            ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" \
                -e "kata_webhook_runtimeclass=$kata_webhook_runtimeclass" \
                $kud_playbooks/configure-kata-webhook.yml \
                --become --become-user=root | \
                sudo tee $log_folder/setup-kata-webhook.log
            kata_webhook_deployed=true
            pushd $kud_tests
            bash kata.sh || failed_kud_tests="${failed_kud_tests} kata"
            popd
        fi
        # Run other plugin tests
        # The topology-manager is added to the tests here as it is
        # enabled via kubelet config, not an addon
        for addon in topology-manager ${kud_addons}; do
            pushd $kud_tests
            bash ${addon}.sh || failed_kud_tests="${failed_kud_tests} ${addon}"
            popd
        done
        # Remove Kata webhook if user didn't want it permanently installed
        if ! [ "${enable_kata_webhook}" == "true" ] && [ "${kata_webhook_deployed}" == "true" ]; then
            ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" \
                -e "kata_webhook_runtimeclass=$kata_webhook_runtimeclass" \
                $kud_playbooks/configure-kata-webhook-reset.yml \
                --become --become-user=root | \
                sudo tee $log_folder/kata-webhook-reset.log
        fi
        if [[ ! -z "$failed_kud_tests" ]]; then
            echo "Test cases failed:${failed_kud_tests}"
            return 1
        fi
    fi
    # Check if Kata webhook should be installed and isn't already installed
    if [ "$enable_kata_webhook" == "true" ] && ! [ "$kata_webhook_deployed" == "true" ]; then
        ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" \
            -e "kata_webhook_runtimeclass=$kata_webhook_runtimeclass" \
            $kud_playbooks/configure-kata-webhook.yml \
            --become --become-user=root | \
            sudo tee $log_folder/setup-kata-webhook.log
    fi
    echo "Add-ons deployment complete..."
}

# install_plugin() - Install ONAP Multicloud Kubernetes plugin
function install_plugin {
    echo "Installing multicloud/k8s plugin"
    sudo -E pip install --no-cache-dir docker-compose

    sudo mkdir -p /opt/{kubeconfig,consul/config}
    sudo cp $HOME/.kube/config /opt/kubeconfig/kud

    pushd $kud_folder/../../../deployments
    sudo ./build.sh
    if [[ "${testing_enabled}" == "true" ]]; then
        sudo ./start.sh
        pushd $kud_tests
        for functional_test in plugin plugin_edgex plugin_fw plugin_eaa; do
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

    master_ip=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' | awk -F '[:/]' '{print $4}')

    printf "Kubernetes Info\n===============\n" > $k8s_info_file
    echo "Dashboard URL: https://$master_ip:$node_port" >> $k8s_info_file
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
container_runtime=${CONTAINER_RUNTIME:-docker}
enable_kata_webhook=${ENABLE_KATA_WEBHOOK:-false}
kata_webhook_runtimeclass=${KATA_WEBHOOK_RUNTIMECLASS:-kata-clh}
kata_webhook_deployed=false
# For containerd the etcd_deployment_type: docker is the default and doesn't work.
# You have to use either etcd_kubeadm_enabled: true or etcd_deployment_type: host
# See https://github.com/kubernetes-sigs/kubespray/issues/5713
kud_kata_override_variables="container_manager=containerd \
    -e etcd_deployment_type=host -e kubelet_cgroup_driver=cgroupfs \
    -e \"{'download_localhost': false}\" -e \"{'download_run_once': false}\""

sudo mkdir -p $log_folder
sudo mkdir -p /opt/csar
sudo chown -R $USER /opt/csar
# Install dependencies
# Setup proxy variables
if [ -f $kud_folder/sources.list ]; then
    sudo mv /etc/apt/sources.list /etc/apt/sources.list.backup
    sudo cp $kud_folder/sources.list /etc/apt/sources.list
fi
echo "Removing ppa for jonathonf/python-3.6"
sudo ls /etc/apt/sources.list.d/ || true
sudo find /etc/apt/sources.list.d -maxdepth 1 -name '*jonathonf*' -delete || true
sudo apt-get update
_install_go
install_k8s
_set_environment_file
install_addons
if ${KUD_PLUGIN_ENABLED:-false}; then
    install_plugin
fi
_print_kubernetes_info
