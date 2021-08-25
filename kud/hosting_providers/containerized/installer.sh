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
    apt-get update
    apt-get install -y software-properties-common
    add-apt-repository -y ppa:longsleep/golang-backports
    apt-get update
    apt-get install -y \
            curl \
            gettext-base \
            git \
            golang-go \
            make \
            python3-pip \
            rsync \
            sshpass \
            sudo \
            unzip \
            vim \
            wget
    update-alternatives --install /usr/bin/pip pip /usr/bin/pip3 1
}

# _install_ansible() - Install and Configure Ansible program
function _install_ansible {
    local version=$(grep "ansible_version" ${kud_playbooks}/kud-vars.yml |
        awk -F ': ' '{print $2}')
    mkdir -p /etc/ansible/
    pip install --no-cache-dir ansible==$version
}

function install_kubespray {
    echo "Deploying kubernetes"
    version=$(grep "kubespray_version" ${kud_playbooks}/kud-vars.yml | \
        awk -F ': ' '{print $2}')
    local_release_dir=$(grep "local_release_dir" \
        $kud_inventory_folder/group_vars/k8s-cluster.yml | \
        awk -F "\"" '{print $2}')
    local tarball=v$version.tar.gz
    _install_ansible
    wget https://github.com/kubernetes-incubator/kubespray/archive/$tarball
    tar -C $dest_folder -xzf $tarball
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

# install_k8s() - Install Kubernetes using kubespray tool including Kata
function install_k8s {
    local cluster_name=$1
    ansible-playbook $verbose -i \
        $kud_inventory $kud_playbooks/preconfigure-kubespray.yml \
        --become --become-user=root | \
        tee $cluster_log/setup-kubernetes.log
    if [ "$container_runtime" == "docker" ]; then
        echo "Docker will be used as the container runtime interface"
        ansible-playbook $verbose -i \
            $kud_inventory $dest_folder/kubespray-$version/cluster.yml \
            -e cluster_name=$cluster_name --become --become-user=root | \
            tee $cluster_log/setup-kubernetes.log
    elif [ "$container_runtime" == "containerd" ]; then
        echo "Containerd will be used as the container runtime interface"
        ansible-playbook $verbose -i \
            $kud_inventory $dest_folder/kubespray-$version/cluster.yml \
            -e $kud_kata_override_variables -e cluster_name=$cluster_name \
            --become --become-user=root | \
            tee $cluster_log/setup-kubernetes.log
        #Install Kata Containers in containerd scenario
        ansible-playbook $verbose -i \
            $kud_inventory -e "base_dest=$HOME" \
            $kud_playbooks/configure-kata.yml | \
            tee $cluster_log/setup-kata.log
    else
        echo "Only Docker or Containerd are supported container runtimes"
        exit 1
    fi

    # Configure environment
    # Requires kubeconfig_localhost and kubectl_localhost to be true
    # in inventory/group_vars/k8s-cluster.yml
    mkdir -p $HOME/.kube
    cp $kud_inventory_folder/artifacts/admin.conf $HOME/.kube/config
    if !(which kubectl); then
        cp $kud_inventory_folder/artifacts/kubectl /usr/local/bin/
    fi
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
        $kud_inventory -e "base_dest=$HOME" $kud_playbooks/configure-kud.yml \
        | tee $cluster_log/setup-kud.log

    kud_addons="${KUD_ADDONS:-} ${plugins_name}"

    for addon in ${kud_addons}; do
        echo "Deploying $addon using configure-$addon.yml playbook.."
        ansible-playbook $verbose -i \
            $kud_inventory -e "base_dest=$HOME" \
            $kud_playbooks/configure-${addon}.yml | \
            tee $cluster_log/setup-${addon}.log
    done

    echo "Run the test cases if testing_enabled is set to true."
    if [[ "${testing_enabled}" == "true" ]]; then
        failed_kud_tests=""
        # Run Kata test first if Kata was installed
        if [ "$container_runtime" == "containerd" ]; then
            #Install Kata webhook for test pods
            ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" \
                -e "kata_webhook_runtimeclass=$kata_webhook_runtimeclass" \
                $kud_playbooks/configure-kata-webhook.yml \
                --become --become-user=root | \
                sudo tee $cluster_log/setup-kata-webhook.log
            kata_webhook_deployed=true
            pushd $kud_tests
            bash kata.sh || failed_kud_tests="${failed_kud_tests} kata"
            popd
        fi
        #Run other plugin tests
        for addon in ${kud_addons}; do
            pushd $kud_tests
            bash ${addon}.sh || failed_kud_tests="${failed_kud_tests} ${addon}"
            case $addon in
                "onap4k8s" )
                    echo "Test the onap4k8s plugin installation"
                    for functional_test in plugin_edgex plugin_fw plugin_eaa; do
                        bash ${functional_test}.sh --external || failed_kud_tests="${failed_kud_tests} ${functional_test}"
                    done
                    ;;
                "emco" )
                    echo "Test the emco plugin installation"
                    # TODO plugin_fw_v2 requires virtlet and a patched multus to succeed
                    # for functional_test in plugin_fw_v2; do
                    #     bash ${functional_test}.sh --external || failed_kud_tests="${failed_kud_tests} ${functional_test}"
                    # done
                    ;;
            esac
            popd
        done
        # Remove Kata webhook if user didn't want it permanently installed
        if ! [ "$enable_kata_webhook" == "true" ] && [ "$kata_webhook_deployed" == "true" ]; then
            ansible-playbook $verbose -i $kud_inventory -e "base_dest=$HOME" \
                -e "kata_webhook_runtimeclass=$kata_webhook_runtimeclass" \
                $kud_playbooks/configure-kata-webhook-reset.yml \
                --become --become-user=root | \
                sudo tee $cluster_log/kata-webhook-reset.log
            kata_webhook_deployed=false
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
            sudo tee $cluster_log/setup-kata-webhook.log
    fi

    echo "Add-ons deployment complete..."
}

function master_ip {
    kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}' | awk -F '[:/]' '{print $4}'
}

# Copy installation artifacts to be usable in host running Ansible
function install_host_artifacts {
    local -r cluster_name=$1
    local -r host_dir="/opt/kud/multi-cluster"
    local -r host_addons_dir="${host_dir}/addons"
    local -r host_artifacts_dir="${host_dir}/${cluster_name}/artifacts"

    for addon in cdi cdi-operator cpu-manager kubevirt kubevirt-operator multus-cni node-feature-discovery ovn4nfv ovn4nfv-network qat-device-plugin sriov-network sriov-network-operator; do
        mkdir -p ${host_addons_dir}/${addon}/{helm,profile}
        cp -r ${kud_infra_folder}/helm/${addon} ${host_addons_dir}/${addon}/helm
        cp -r ${kud_infra_folder}/profiles/${addon}/* ${host_addons_dir}/${addon}/profile
        tar -czf ${host_addons_dir}/${addon}.tar.gz -C ${host_addons_dir}/${addon}/helm .
        tar -czf ${host_addons_dir}/${addon}_profile.tar.gz -C ${host_addons_dir}/${addon}/profile .
    done

    mkdir -p ${host_addons_dir}/tests
    for test in _common _common_test _functions topology-manager-sriov kubevirt multus ovn4nfv nfd sriov-network qat cmk; do
        cp ${kud_tests}/${test}.sh ${host_addons_dir}/tests
    done
    cp ${kud_tests}/plugin_fw_v2.sh ${host_addons_dir}/tests
    cp ${kud_tests}/plugin_fw_v2.yaml ${host_addons_dir}/tests
    cp -r ${kud_tests}/../demo/composite-firewall ${host_addons_dir}/tests

    mkdir -p ${host_artifacts_dir}
    cp -rf ${kud_inventory_folder}/artifacts/* ${host_artifacts_dir}

    mkdir -p ${host_artifacts_dir}/addons
    for yaml in ${kud_infra_folder}/emco/examples/*.yaml; do
        cp ${yaml} ${host_artifacts_dir}/addons
    done
    for template in addons/*.tmpl; do
        CLUSTER_NAME="${cluster_name}" \
        HOST_IP="$(master_ip)" \
        KUBE_PATH="${host_artifacts_dir}/admin.conf" \
        PACKAGES_PATH="${host_addons_dir}" \
        envsubst <${template} >${host_artifacts_dir}/${template%.tmpl}
    done
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

    printf "Kubernetes Info\n===============\n" > $k8s_info_file
    echo "Dashboard URL: https://$(master_ip):$node_port" >> $k8s_info_file
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
container_runtime=${CONTAINER_RUNTIME:-docker}
enable_kata_webhook=${ENABLE_KATA_WEBHOOK:-false}
kata_webhook_runtimeclass=${KATA_WEBHOOK_RUNTIMECLASS:-kata-qemu}
kata_webhook_deployed=false
# For containerd the etcd_deployment_type: docker is the default and doesn't work.
# You have to use either etcd_kubeadm_enabled: true or etcd_deployment_type: host
# See https://github.com/kubernetes-sigs/kubespray/issues/5713
kud_kata_override_variables="container_manager=containerd \
    -e etcd_deployment_type=host -e kubelet_cgroup_driver=cgroupfs"

mkdir -p /opt/csar
export CSAR_DIR=/opt/csar

function install_pkg {
    install_prerequisites
    install_kubespray
}

function install_cluster {
    version=$(grep "kubespray_version" ${kud_playbooks}/kud-vars.yml | \
        awk -F ': ' '{print $2}')
    export ANSIBLE_CONFIG=$dest_folder/kubespray-$version/ansible.cfg
    install_k8s $1
    if [ ${2:+1} ]; then
        echo "install default addons and $2"
        install_addons "$2"
    else
        install_addons
    fi
    echo "installed the addons"

    install_host_artifacts $1

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
