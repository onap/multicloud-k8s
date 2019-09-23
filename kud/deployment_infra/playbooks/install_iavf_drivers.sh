#!/bin/bash

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

function is_used {
    local ifname=$1
    route_info=`ip route show | grep $ifname`
    if [ -z "$route_info" ]; then
        return 0
    else
        return 1
    fi
}

function get_sriov_ifname {
    for net_device in /sys/class/net/*/ ; do
        if [ -e $net_device/device/sriov_numvfs ] ; then
            ifname=$(basename $net_device)
            is_used $ifname
            if [ "$?" = "0" ]; then
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
