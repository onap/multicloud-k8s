#!/bin/bash

# Based on:
# https://gerrit.akraino.org/r/#/c/icn/+/1359/1/deploy/kud-plugin-addons/device-plugins/sriov/driver/install_iavf_drivers.sh

nic_models=(X710 XL710 X722)
nic_drivers=(i40e)
device_checkers=(is_not_used is_driver_match is_model_match)

function install_iavf_driver {
    local ifname=$1

    echo "Installing modules..."
    echo "Installing i40evf blacklist file..."
    mkdir -p "/etc/modprobe.d/"
    echo "blacklist i40evf" > "/etc/modprobe.d/iavf-blacklist-i40evf.conf"

    kver=`uname -a | awk '{print $3}'`
    install_mod_dir=/lib/modules/$kver/updates/drivers/net/ethernet/intel/iavf/
    echo "Installing driver in $install_mod_dir"
    mkdir -p $install_mod_dir
    cp iavf.ko $install_mod_dir

    echo "Installing kernel module i40evf..."
    depmod -a
    modprobe i40evf
    modprobe iavf

    echo "Enabling VF on interface $ifname..."
    echo "/sys/class/net/$ifname/device/sriov_numvfs"
    echo '8' > /sys/class/net/$ifname/device/sriov_numvfs
}

function is_not_used {
    local ifname=$1
    route_info=`ip route show | grep $ifname`
    if [ -z "$route_info" ]; then
        return 1
    else
        return 0
    fi
}

function is_driver_match {
    local ifname=$1
    driver=`cat /sys/class/net/$ifname/device/uevent | grep DRIVER | cut -f2 -d "="`
    if [ ! -z "$driver" ]; then
        for nic_driver in ${nic_drivers[@]}; do
            if [ "$driver" = "$nic_driver" ]; then
                return 1
            fi
        done
    fi
    return 0
}

function is_model_match {
    local ifname=$1
    pci_addr=`cat /sys/class/net/$ifname/device/uevent | grep PCI_SLOT_NAME | cut -f2 -d "=" | cut -f2,3 -d ":"`
    if [ ! -z "$pci_addr" ]; then
        for nic_model in ${nic_models[@]}; do
            model_match=$(lspci | grep $pci_addr | grep $nic_model)
            if [ ! -z "$model_match" ]; then
                return 1
            fi
        done
    fi
    return 0
}

function get_sriov_ifname {
    for net_device in /sys/class/net/*/ ; do
        if [ -e $net_device/device/sriov_numvfs ] ; then
            ifname=$(basename $net_device)
            for device_checker in ${device_checkers[@]}; do
                eval $device_checker $ifname
                if [ "$?" = "0" ]; then
                    ifname=""
                    break
                fi
            done
            if [ ! -z "$ifname" ]; then
                echo $ifname
                return
            fi
        fi
    done
    echo ''
}

if [ $# -ne 1 ] ; then
    ifname=$(get_sriov_ifname)
    if [ -z "$ifname" ]; then
        echo "Cannot find Nic with SRIOV support."
    else
        install_iavf_driver $ifname
    fi
else
    ifname=$1
    if [ ! -e /sys/class/net/$ifname/device/sriov_numvfs ] ; then
        echo "${ifname} is not a valid sriov interface"
    else
        install_iavf_driver $ifname
    fi
fi
