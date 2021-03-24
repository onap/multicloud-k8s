#!/bin/bash
#
# The functions below are captured from the Makefile targets.  They
# cannot be run in a container as-is due to absolute paths, so they
# are recreated here.
#
# Note also that the portions of qat-driver-install that deal with
# rc.d are removed: they are intended to be handled by the deployed
# DaemonSet.  The rest is contained in _qat_service_start.
#
# The checks for loaded modules are moved to _qat_check_started.

BIN_LIST="qat_c3xxx.bin qat_c3xxx_mmp.bin qat_c62x.bin \
    qat_c62x_mmp.bin qat_mmp.bin qat_d15xx.bin qat_d15xx_mmp.bin \
    qat_200xx.bin qat_200xx_mmp.bin qat_895xcc.bin qat_895xcc_mmp.bin"

numDh895xDevicesP=$(lspci -n | grep -c "8086:0435") || true
numDh895xDevicesV=$(lspci -n | grep -c "8086:0443") || true
numC62xDevicesP=$(lspci -n | grep -c "8086:37c8") || true
numC62xDevicesV=$(lspci -n | grep -c "8086:37c9") || true
numD15xxDevicesP=$(lspci -n | grep -c "8086:6f54") || true
numD15xxDevicesV=$(lspci -n | grep -c "8086:6f55") || true
numC3xxxDevicesP=$(lspci -n | grep -c "8086:19e2") || true
numC3xxxDevicesV=$(lspci -n | grep -c "8086:19e3") || true
num200xxDevicesP=$(lspci -n | grep -c "8086:18ee") || true
num200xxDevicesV=$(lspci -n | grep -c "8086:18ef") || true

_qat_driver_install() {
    info "Installing drivers"
    if [[ -z "${KERNEL_MOD_SIGN_CMD}" ]]; then
        info "No driver signing required"
        INSTALL_MOD_PATH=${ROOT_MOUNT_DIR} make KDIR="${KERNEL_SRC_DIR}" -C "${QAT_INSTALL_DIR_CONTAINER}/quickassist/qat" mod_sign_cmd=":" modules_install
    else
        info "Driver signing is required"
        INSTALL_MOD_PATH=${ROOT_MOUNT_DIR} make KDIR="${KERNEL_SRC_DIR}" -C "${QAT_INSTALL_DIR_CONTAINER}/quickassist/qat" mod_sign_cmd="${KERNEL_MOD_SIGN_CMD}" modules_install
    fi
}

_adf_ctl_install() {
    info "Installing adf_ctl"
    install -D -m 750 "${QAT_INSTALL_DIR_CONTAINER}/quickassist/utilities/adf_ctl/adf_ctl" "${ROOT_MOUNT_DIR}/usr/local/bin/adf_ctl"
}

_adf_ctl_uninstall() {
    info "Uninstalling adf_ctl"
    # rm ${ROOT_MOUNT_DIR}/usr/local/bin/adf_ctl
    return 0
}

_rename_ssl_conf_section() {
    info "Renaming SSL section in conf files"
    restore_nullglob=$(shopt -p | grep nullglob)
    shopt -s nullglob
    for file in ${ROOT_MOUNT_DIR}/etc/dh895xcc_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/dh895xcc_dev${dev}.conf"
    done
    for file in ${ROOT_MOUNT_DIR}/etc/c6xx_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/c6xx_dev${dev}.conf"
    done
    for file in ${ROOT_MOUNT_DIR}/etc/d15xx_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/d15xx_dev${dev}.conf"
    done
    for file in ${ROOT_MOUNT_DIR}/etc/c3xxx_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/c3xxx_dev${dev}.conf"
    done
    for file in ${ROOT_MOUNT_DIR}/etc/200xx_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/200xx_dev${dev}.conf"
    done

    for file in ${ROOT_MOUNT_DIR}/etc/dh895xccvf_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/dh895xccvf_dev${dev}.conf"
    done
    for file in ${ROOT_MOUNT_DIR}/etc/c6xxvf_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/c6xxvf_dev${dev}.conf"
    done
    for file in ${ROOT_MOUNT_DIR}/etc/d15xxvf_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/d15xxvf_dev${dev}.conf"
    done
    for file in ${ROOT_MOUNT_DIR}/etc/c3xxxvf_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/c3xxxvf_dev${dev}.conf"
    done
    for file in ${ROOT_MOUNT_DIR}/etc/200xxvf_dev*.conf; do
        dev=$(echo "$file" | cut -d '_' -f 2 | tr -cd '[:digit:]')
        sed -i "s/\[SSL\]/\[SSL${dev}\]/g" "${ROOT_MOUNT_DIR}/etc/200xxvf_dev${dev}.conf"
    done
    $restore_nullglob
}

_qat_service_install() {
    local -r QAT_DH895XCC_NUM_VFS=32
    local -r QAT_DHC62X_NUM_VFS=16
    local -r QAT_DHD15XX_NUM_VFS=16
    local -r QAT_DHC3XXX_NUM_VFS=16
    local -r QAT_DH200XX_NUM_VFS=16
    local -r DEVICES="0435 0443 37c8 37c9 6f54 6f55 19e2 19e3 18ee 18ef"

    info "Installing service"
    pushd "${QAT_INSTALL_DIR_CONTAINER}/build" > /dev/null

    if [[ ! -d ${ROOT_MOUNT_DIR}/lib/firmware/qat_fw_backup ]]; then
        mkdir -p "${ROOT_MOUNT_DIR}/lib/firmware/qat_fw_backup"
    fi
    for bin in ${BIN_LIST}; do
        if [[ -e ${ROOT_MOUNT_DIR}/lib/firmware/${bin} ]]; then
            mv "${ROOT_MOUNT_DIR}/lib/firmware/${bin}" "${ROOT_MOUNT_DIR}/lib/firmware/qat_fw_backup/${bin}"
        fi
        if [[ -e ${bin} ]]; then
            install -D -m 750 "${bin}" "${ROOT_MOUNT_DIR}/lib/firmware/${bin}"
        fi
    done
    if [[ ! -d ${ROOT_MOUNT_DIR}/etc/qat_conf_backup ]]; then
        mkdir "${ROOT_MOUNT_DIR}/etc/qat_conf_backup"
    fi
    mv "${ROOT_MOUNT_DIR}/etc/dh895xcc*.conf" "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/" 2>/dev/null || true
    mv "${ROOT_MOUNT_DIR}/etc/c6xx*.conf" "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/" 2>/dev/null || true
    mv "${ROOT_MOUNT_DIR}/etc/d15xx*.conf" "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/" 2>/dev/null || true
    mv "${ROOT_MOUNT_DIR}/etc/c3xxx*.conf" "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/" 2>/dev/null || true
    mv "${ROOT_MOUNT_DIR}/etc/200xx*.conf" "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/" 2>/dev/null || true
    if [[ "${QAT_ENABLE_SRIOV}" != "guest" ]]; then
        for ((dev=0; dev<numDh895xDevicesP; dev++)); do
            install -D -m 640 dh895xcc_dev0.conf "${ROOT_MOUNT_DIR}/etc/dh895xcc_dev${dev}.conf"
        done
        for ((dev=0; dev<numC62xDevicesP; dev++)); do
            install -D -m 640 c6xx_dev$((dev%3)).conf "${ROOT_MOUNT_DIR}/etc/c6xx_dev${dev}.conf"
        done
        for ((dev=0; dev<numD15xxDevicesP; dev++)); do
            install -D -m 640 d15xx_dev$((dev%3)).conf "${ROOT_MOUNT_DIR}/etc/d15xx_dev${dev}.conf"
        done
        for ((dev=0; dev<numC3xxxDevicesP; dev++)); do
            install -D -m 640 c3xxx_dev0.conf "${ROOT_MOUNT_DIR}/etc/c3xxx_dev${dev}.conf"
        done
        for ((dev=0; dev<num200xxDevicesP; dev++)); do
            install -D -m 640 200xx_dev0.conf "${ROOT_MOUNT_DIR}/etc/200xx_dev${dev}.conf"
        done
    fi
    if [[ "${QAT_ENABLE_SRIOV}" == "host" ]]; then
        for ((dev=0; dev<numDh895xDevicesP; dev++)); do
            for ((vf_dev=0; vf_dev<QAT_DH895XCC_NUM_VFS; vf_dev++)); do
                vf_dev_num=$((dev * QAT_DH895XCC_NUM_VFS + vf_dev))
                install -D -m 640 dh895xccvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/dh895xccvf_dev${vf_dev_num}.conf"
            done
        done
        for ((dev=0; dev<numC62xDevicesP; dev++)); do
            for ((vf_dev=0; vf_dev<QAT_DHC62X_NUM_VFS; vf_dev++)); do
                vf_dev_num=$((dev * QAT_DHC62X_NUM_VFS + vf_dev))
                install -D -m 640 c6xxvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/c6xxvf_dev${vf_dev_num}.conf"
            done
        done
        for ((dev=0; dev<numD15xxDevicesP; dev++)); do
            for ((vf_dev=0; vf_dev<QAT_DHD15XX_NUM_VFS; vf_dev++)); do
                vf_dev_num=$((dev * QAT_DHD15XX_NUM_VFS + vf_dev))
                install -D -m 640 d15xxvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/d15xxvf_dev${vf_dev_num}.conf"
            done
        done
        for ((dev=0; dev<numC3xxxDevicesP; dev++)); do
            for ((vf_dev=0; vf_dev<QAT_DHC3XXX_NUM_VFS; vf_dev++)); do
                vf_dev_num=$((dev * QAT_DHC3XXX_NUM_VFS + vf_dev))
                install -D -m 640 c3xxxvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/c3xxxvf_dev${vf_dev_num}.conf"
            done
        done
        for ((dev=0; dev<num200xxDevicesP; dev++)); do
            for ((vf_dev=0; vf_dev<QAT_DH200XX_NUM_VFS; vf_dev++)); do
                vf_dev_num=$((dev * QAT_DH200XX_NUM_VFS + vf_dev))
                install -D -m 640 200xxvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/200xxvf_dev${vf_dev_num}.conf"
            done
        done
    else
        for ((dev=0; dev<numDh895xDevicesV; dev++)); do
            install -D -m 640 dh895xccvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/dh895xccvf_dev${dev}.conf"
        done
        for ((dev=0; dev<numC62xDevicesV; dev++)); do
            install -D -m 640 c6xxvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/c6xxvf_dev${dev}.conf"
        done
        for ((dev=0; dev<numD15xxDevicesV; dev++)); do
            install -D -m 640 d15xxvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/d15xxvf_dev${dev}.conf"
        done
        for ((dev=0; dev<numC3xxxDevicesV; dev++)); do
            install -D -m 640 c3xxxvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/c3xxxvf_dev${dev}.conf"
        done
        for ((dev=0; dev<num200xxDevicesV; dev++)); do
            install -D -m 640 200xxvf_dev0.conf.vm "${ROOT_MOUNT_DIR}/etc/200xxvf_dev${dev}.conf"
        done
    fi
    _rename_ssl_conf_section
    info "Creating startup and kill scripts"
    install -D -m 750 qat_service "${ROOT_MOUNT_DIR}/etc/init.d/qat_service"
    if [[ "${QAT_ENABLE_SRIOV}" == "host" ]]; then
        install -D -m 750 qat_service_vfs "${ROOT_MOUNT_DIR}/etc/init.d/qat_service_vfs"
    fi
    if [[ "${QAT_ENABLE_SRIOV}" == "host" || "${QAT_ENABLE_SRIOV}" == "guest" ]]; then
        echo "# Comment or remove next line to disable sriov" > "${ROOT_MOUNT_DIR}/etc/default/qat"
        echo "SRIOV_ENABLE=1" >> "${ROOT_MOUNT_DIR}/etc/default/qat"
    else
        echo "# Remove comment on next line to enable sriov" > "${ROOT_MOUNT_DIR}/etc/default/qat"
        echo "#SRIOV_ENABLE=1" >> "${ROOT_MOUNT_DIR}/etc/default/qat"
    fi
    echo "#LEGACY_LOADED=1" >> "${ROOT_MOUNT_DIR}/etc/default/qat"
    rm -f "${ROOT_MOUNT_DIR}/etc/modprobe.d/blacklist-qat-vfs.conf"
    if [[ "${QAT_ENABLE_SRIOV}" == "host" ]]; then
        if [[ ${numDh895xDevicesP} != 0 ]]; then
            echo "blacklist qat_dh895xccvf" >> "${ROOT_MOUNT_DIR}/etc/modprobe.d/blacklist-qat-vfs.conf"
        fi
        if [[ ${numC3xxxDevicesP} != 0 ]]; then
            echo "blacklist qat_c3xxxvf" >> "${ROOT_MOUNT_DIR}/etc/modprobe.d/blacklist-qat-vfs.conf"
        fi
        if [[ ${num200xxDevicesP} != 0 ]]; then
            echo "blacklist qat_200xxvf" >> "${ROOT_MOUNT_DIR}/etc/modprobe.d/blacklist-qat-vfs.conf"
        fi
        if [[ ${numC62xDevicesP} != 0 ]]; then
            echo "blacklist qat_c62xvf" >> "${ROOT_MOUNT_DIR}/etc/modprobe.d/blacklist-qat-vfs.conf"
        fi
        if [[ ${numD15xxDevicesP} != 0 ]]; then
            echo "blacklist qat_d15xxvf" >> "${ROOT_MOUNT_DIR}/etc/modprobe.d/blacklist-qat-vfs.conf"
        fi
    fi
    echo "#ENABLE_KAPI=1" >> "${ROOT_MOUNT_DIR}/etc/default/qat"
    info "Copying libqat_s.so to ${ROOT_MOUNT_DIR}/usr/local/lib"
    install -D -m 755 libqat_s.so "${ROOT_MOUNT_DIR}/usr/local/lib/libqat_s.so"
    info "Copying libusdm_drv_s.so to ${ROOT_MOUNT_DIR}/usr/local/lib"
    install -D -m 755 libusdm_drv_s.so "${ROOT_MOUNT_DIR}/usr/local/lib/libusdm_drv_s.so"
    echo /usr/local/lib > "${ROOT_MOUNT_DIR}/etc/ld.so.conf.d/qat.conf"
    ldconfig -r "${ROOT_MOUNT_DIR}"
    info "Copying usdm module to system drivers"
    if [[ ! -z "${KERNEL_MOD_SIGN_CMD}" ]]; then
        info "Need to sign driver usdm_drv.ko"
        ${KERNEL_MOD_SIGN_CMD} usdm_drv.ko
        info "Need to sign driver qat_api.ko"
        ${KERNEL_MOD_SIGN_CMD} qat_api.ko
    fi
    install usdm_drv.ko  "${ROOT_MOUNT_DIR}/lib/modules/$(uname -r)/kernel/drivers"
    install qat_api.ko  "${ROOT_MOUNT_DIR}/lib/modules/$(uname -r)/kernel/drivers"
    if [[ ! $(chroot "${ROOT_MOUNT_DIR}" getent group qat) ]]; then
        info "Creating qat group"
        groupadd -R "${ROOT_MOUNT_DIR}" qat
    else
        info "Group qat already exists"
    fi
    info "Creating udev rules"
    rm -f "${ROOT_MOUNT_DIR}/etc/udev/rules.d/00-qat.rules"
    {
        echo 'KERNEL=="qat_adf_ctl" MODE="0660" GROUP="qat"';
        echo 'KERNEL=="qat_dev_processes" MODE="0660" GROUP="qat"';
        echo 'KERNEL=="usdm_drv" MODE="0660" GROUP="qat"';
        echo 'ACTION=="add", DEVPATH=="/module/usdm_drv" SUBSYSTEM=="module" RUN+="/bin/mkdir /dev/hugepages/qat"';
        echo 'ACTION=="add", DEVPATH=="/module/usdm_drv" SUBSYSTEM=="module" RUN+="/bin/chgrp qat /dev/hugepages/qat"';
        echo 'ACTION=="add", DEVPATH=="/module/usdm_drv" SUBSYSTEM=="module" RUN+="/bin/chmod 0770 /dev/hugepages/qat"';
        echo 'ACTION=="remove", DEVPATH=="/module/usdm_drv" SUBSYSTEM=="module" RUN+="/bin/rmdir /dev/hugepages/qat"';
        for dev in ${DEVICES}; do
            echo 'KERNEL=="uio*", ATTRS{vendor}=="0x'"$(echo "8086" | tr -d \")"'", ATTRS{device}=="0x'"$(echo "${dev}" | tr -d \")"'"  MODE="0660" GROUP="qat"';
        done
    } > "${ROOT_MOUNT_DIR}/etc/udev/rules.d/00-qat.rules"
    info "Creating module.dep file for QAT released kernel object"
    info "This will take a few moments"
    depmod -a -b "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/depmod.d"

    popd > /dev/null
}

_qat_service_start() {
    if [[ $(lsmod | grep -c "usdm_drv") != "0" ]]; then
        rmmod usdm_drv
    fi
    info "Starting QAT service"
    info "... shutting down"
    chroot "${ROOT_MOUNT_DIR}" /etc/init.d/qat_service shutdown || true
    sleep 3
    info "... starting"
    chroot "${ROOT_MOUNT_DIR}" /etc/init.d/qat_service start
    if [[ "${QAT_ENABLE_SRIOV}" == "host" ]]; then
        modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" vfio-pci
        chroot "${ROOT_MOUNT_DIR}" /etc/init.d/qat_service_vfs start
    fi
    info "... started"
}

_qat_check_started() {
    if [[ $(lsmod | grep -c "usdm_drv") == "0" ]]; then
        error "usdm_drv module not installed"
        return "${RETCODE_ERROR}"
    fi
    if [[ ${numDh895xDevicesP} != 0 ]]; then
        if [[ $(lsmod | grep -c "qat_dh895xcc") == "0" ]]; then
            error "qat_dh895xcc module not installed"
            return "${RETCODE_ERROR}"
        fi
    fi
    if [[ ${numC62xDevicesP} != 0 ]]; then
        if [[ $(lsmod | grep -c "qat_c62x") == "0" ]]; then
            error "qat_c62x module not installed"
            return "${RETCODE_ERROR}"
        fi
    fi
    if [[ ${numD15xxDevicesP} != 0 ]]; then
        if [[ $(lsmod | grep -c "qat_d15xx") == "0" ]]; then
            error "qat_d15xx module not installed"
            return "${RETCODE_ERROR}"
        fi
    fi
    if [[ ${numC3xxxDevicesP} != 0 ]]; then
        if [[ $(lsmod | grep -c "qat_c3xxx") == "0" ]]; then
            error "qat_c3xxx module not installed"
            return "${RETCODE_ERROR}"
        fi
    fi
    if [[ ${num200xxDevicesP} != 0 ]]; then
        if [[ $(lsmod | grep -c "qat_200xx") == "0" ]]; then
            error "qat_200xx module not installed"
            return "${RETCODE_ERROR}"
        fi
    fi
    if [[ "${QAT_ENABLE_SRIOV}" == "guest" ]]; then
        if [[ ${numDh895xDevicesV} != 0 ]]; then
            if [[ $(lsmod | grep -c "qat_dh895xccvf") == "0" ]]; then
                error "qat_dh895xccvf module not installed"
                return "${RETCODE_ERROR}"
            fi
        fi
        if [[ ${numC62xDevicesV} != 0 ]]; then
            if [[ $(lsmod | grep -c "qat_c62xvf") == "0" ]]; then
                error "qat_c62xvf module not installed"
                return "${RETCODE_ERROR}"
            fi
        fi
        if [[ ${numD15xxDevicesV} != 0 ]]; then
            if [[ $(lsmod | grep -c "qat_d15xxvf") == "0" ]]; then
                error "qat_d15xxvf module not installed"
                return "${RETCODE_ERROR}"
            fi
        fi
        if [[ ${numC3xxxDevicesV} != 0 ]]; then
            if [[ $(lsmod | grep -c "qat_c3xxxvf") == "0" ]]; then
                error "qat_c3xxxvf module not installed"
                return "${RETCODE_ERROR}"
            fi
        fi
        if [[ ${num200xxDevicesV} != 0 ]]; then
            if [[ $(lsmod | grep -c "qat_200xxvf") == "0" ]]; then
                error "qat_200xxvf module not installed"
                return "${RETCODE_ERROR}"
            fi
        fi
    fi
    if [[ $("${ROOT_MOUNT_DIR}/usr/local/bin/adf_ctl" status | grep -c "state: down") != "0" ]]; then
        error "QAT driver not activated"
        return "${RETCODE_ERROR}"
    fi
}

_qat_service_shutdown() {
    info "Stopping service"
    if [[ $(lsmod | grep -c "qat") != "0" || -e ${ROOT_MOUNT_DIR}/lib/modules/$(uname -r)/updates/drivers/crypto/qat/qat_common/intel_qat.ko ]]; then
        if [[ $(lsmod | grep -c "usdm_drv") != "0" ]]; then
            rmmod usdm_drv
        fi
        if [[ -e ${ROOT_MOUNT_DIR}/etc/init.d/qat_service_upstream ]]; then
            until chroot "${ROOT_MOUNT_DIR}" /etc/init.d/qat_service_upstream shutdown; do
                sleep 1
            done
        elif [[ -e ${ROOT_MOUNT_DIR}/etc/init.d/qat_service ]]; then
            until chroot "${ROOT_MOUNT_DIR}" /etc/init.d/qat_service shutdown; do
                sleep 1
            done
        fi
    fi
}

_qat_service_uninstall() {
    info "Uninstalling service"
    if [[ $(lsmod | grep -c "qat") != "0" || -e ${ROOT_MOUNT_DIR}/lib/modules/$(uname -r)/updates/drivers/crypto/qat/qat_common/intel_qat.ko ]]; then
        info "Removing the QAT firmware"
        for bin in ${BIN_LIST}; do
            if [[ -e ${ROOT_MOUNT_DIR}/lib/firmware/${bin} ]]; then
                rm "${ROOT_MOUNT_DIR}/lib/firmware/${bin}"
            fi
            if [[ -e ${ROOT_MOUNT_DIR}/lib/firmware/qat_fw_backup/${bin} ]]; then
                mv "${ROOT_MOUNT_DIR}/lib/firmware/qat_fw_backup/${bin}" "${ROOT_MOUNT_DIR}/lib/firmware/${bin}"
            fi
        done

        if [[ -d ${ROOT_MOUNT_DIR}/lib/firmware/qat_fw ]]; then
            rm "${ROOT_MOUNT_DIR}/lib/firmware/qat_fw_backup"
        fi

        if [[ -e ${ROOT_MOUNT_DIR}/etc/init.d/qat_service_upstream ]]; then
            rm "${ROOT_MOUNT_DIR}/etc/init.d/qat_service_upstream"
            rm "${ROOT_MOUNT_DIR}/usr/local/bin/adf_ctl"
        elif [[ -e ${ROOT_MOUNT_DIR}/etc/init.d/qat_service ]]; then
            rm "${ROOT_MOUNT_DIR}/etc/init.d/qat_service"
            rm "${ROOT_MOUNT_DIR}/usr/local/bin/adf_ctl"
        fi
        rm -f "${ROOT_MOUNT_DIR}/etc/init.d/qat_service_vfs"
        rm -f "${ROOT_MOUNT_DIR}/etc/modprobe.d/blacklist-qat-vfs.conf"

        rm -f "${ROOT_MOUNT_DIR}/usr/local/lib/libqat_s.so"
        rm -f "${ROOT_MOUNT_DIR}/usr/local/lib/libusdm_drv_s.so"
        rm -f "${ROOT_MOUNT_DIR}/etc/ld.so.conf.d/qat.conf"
        ldconfig -r "${ROOT_MOUNT_DIR}"

        info "Removing config files"
        rm -f "${ROOT_MOUNT_DIR}/etc/dh895xcc*.conf"
        rm -f "${ROOT_MOUNT_DIR}/etc/c6xx*.conf"
        rm -f "${ROOT_MOUNT_DIR}/etc/d15xx*.conf"
        rm -f "${ROOT_MOUNT_DIR}/etc/c3xxx*.conf"
        rm -f "${ROOT_MOUNT_DIR}/etc/200xx*.conf"
        rm -f "${ROOT_MOUNT_DIR}/etc/udev/rules.d/00-qat.rules"

        mv -f "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/dh895xcc*.conf" "${ROOT_MOUNT_DIR}/etc/" 2>/dev/null || true
        mv -f "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/c6xx*.conf" "${ROOT_MOUNT_DIR}/etc/" 2>/dev/null || true
        mv -f "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/d15xx*.conf" "${ROOT_MOUNT_DIR}/etc/" 2>/dev/null || true
        mv -f "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/c3xxx*.conf" "${ROOT_MOUNT_DIR}/etc/" 2>/dev/null || true
        mv -f "${ROOT_MOUNT_DIR}/etc/qat_conf_backup/200xx*.conf" "${ROOT_MOUNT_DIR}/etc/" 2>/dev/null || true

        info "Removing drivers modules"
        rm -rf "${ROOT_MOUNT_DIR}/lib/modules/$(uname -r)/updates/drivers/crypto/qat"
        rm -f "${ROOT_MOUNT_DIR}/lib/modules/$(uname -r)/kernel/drivers/usdm_drv.ko"
        rm -f "${ROOT_MOUNT_DIR}/lib/modules/$(uname -r)/kernel/drivers/qat_api.ko"
        info "Creating module.dep file for QAT released kernel object"
        depmod -a -b "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/depmod.d"

        if [[ $(lsmod | grep -c "usdm_drv|intel_qat") != "0" ]]; then
            if [[ $(modinfo intel_qat | grep -c "updates") == "0" ]]; then
                info "In-tree driver loaded"
                info "Acceleration uninstall complete"
            else
                error "Some modules not removed properly"
                error "Acceleration uninstall failed"
            fi
        else
            info "Acceleration uninstall complete"
        fi
        if [[ ${numDh895xDevicesP} != 0 ]]; then
            lsmod | grep qat_dh895xcc >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_dh895xcc >/dev/null 2>&1 || true
        fi
        if [[ ${numC62xDevicesP} != 0 ]]; then
            lsmod | grep qat_c62x >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_c62x >/dev/null 2>&1 || true
        fi
        if [[ ${numD15xxDevicesP} != 0 ]]; then
            lsmod | grep qat_d15xx >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_d15xx >/dev/null 2>&1 || true
        fi
        if [[ ${numC3xxxDevicesP} != 0 ]]; then
            lsmod | grep qat_c3xxx >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_c3xxx >/dev/null 2>&1 || true
        fi
        if [[ ${num200xxDevicesP} != 0 ]]; then
            lsmod | grep qat_200xx >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_200xx >/dev/null 2>&1 || true
        fi
        if [[ ${numDh895xDevicesV} != 0 ]]; then
            lsmod | grep qat_dh895xccvf >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_dh895xccvf >/dev/null 2>&1 || true
        fi
        if [[ ${numC62xDevicesV} != 0 ]]; then
            lsmod | grep qat_c62xvf >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_c62xvf >/dev/null 2>&1 || true
        fi
        if [[ ${numD15xxDevicesV} != 0 ]]; then
            lsmod | grep qat_d15xxvf >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_d15xxvf >/dev/null 2>&1 || true
        fi
        if [[ ${numC3xxxDevicesV} != 0 ]]; then
            lsmod | grep qat_c3xxxvf >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_c3xxxvf >/dev/null 2>&1 || true
        fi
        if [[ ${num200xxDevicesV} != 0 ]]; then
            lsmod | grep qat_200xxvf >/dev/null 2>&1 || modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" -b -q qat_200xxvf >/dev/null 2>&1 || true
        fi
    else
        info "Acceleration package not installed"
    fi
}

_qat_sample_install() {
    info "Installing samples"
    if [[ -f ${QAT_INSTALL_DIR_CONTAINER}/quickassist/utilities/libusdm_drv/linux/build/linux_2.6/user_space/libusdm_drv.a ]]; then
        ICP_ROOT="${QAT_INSTALL_DIR_CONTAINER}" make perf_user -C "${QAT_INSTALL_DIR_CONTAINER}/quickassist/lookaside/access_layer/src/sample_code"
        cp "${QAT_INSTALL_DIR_CONTAINER}/quickassist/lookaside/access_layer/src/sample_code/performance/build/linux_2.6/user_space/cpa_sample_code" "${QAT_INSTALL_DIR_CONTAINER}/build"
        ICP_ROOT="${QAT_INSTALL_DIR_CONTAINER}" KERNEL_SOURCE_ROOT="${KERNEL_SRC_DIR}" make perf_kernel -C "${QAT_INSTALL_DIR_CONTAINER}/quickassist/lookaside/access_layer/src/sample_code"
        cp "${QAT_INSTALL_DIR_CONTAINER}/quickassist/lookaside/access_layer/src/sample_code/performance/build/linux_2.6/kernel_space/cpa_sample_code.ko" "${QAT_INSTALL_DIR_CONTAINER}/build"
    else
        error "No libusdm_drv library found - build the project (make all) before samples"
        return "${RETCODE_ERROR}"
    fi

    if [[ ! -d ${ROOT_MOUNT_DIR}/lib/firmware ]]; then
        mkdir "${ROOT_MOUNT_DIR}/lib/firmware"
    fi

    cp "${QAT_INSTALL_DIR_CONTAINER}/quickassist/lookaside/access_layer/src/sample_code/performance/compression/calgary" "${ROOT_MOUNT_DIR}/lib/firmware"
    cp "${QAT_INSTALL_DIR_CONTAINER}/quickassist/lookaside/access_layer/src/sample_code/performance/compression/calgary32" "${ROOT_MOUNT_DIR}/lib/firmware"
    cp "${QAT_INSTALL_DIR_CONTAINER}/quickassist/lookaside/access_layer/src/sample_code/performance/compression/canterbury" "${ROOT_MOUNT_DIR}/lib/firmware"
    if [[ ! -z "${KERNEL_MOD_SIGN_CMD}" ]]; then
        if [[ -f ${QAT_INSTALL_DIR_CONTAINER}/build/cpa_sample_code.ko ]]; then
            echo "Need to sign sample code ${QAT_INSTALL_DIR_CONTAINER}/build/cpa_sample_code.ko."
            "${KERNEL_MOD_SIGN_CMD}" "${QAT_INSTALL_DIR_CONTAINER}/build/cpa_sample_code.ko"
        fi
    fi

    install -D -m 750 "${QAT_INSTALL_DIR_CONTAINER}/build/cpa_sample_code" "${ROOT_MOUNT_DIR}/usr/local/bin/cpa_sample_code"
    install -D -m 750 "${QAT_INSTALL_DIR_CONTAINER}/build/cpa_sample_code.ko" "${ROOT_MOUNT_DIR}/usr/local/bin/cpa_sample_code.ko"
    info "cpa_sample_code installed under ${ROOT_MOUNT_DIR}/usr/local/bin directory"
}

_qat_sample_uninstall() {
    info "Uninstalling samples"
    rm -f "${ROOT_MOUNT_DIR}/lib/firmware/calgary"
    rm -f "${ROOT_MOUNT_DIR}/lib/firmware/calgary32"
    rm -f "${ROOT_MOUNT_DIR}/lib/firmware/canterbury"

    rm -f "${ROOT_MOUNT_DIR}/usr/local/bin/cpa_sample_code"
    rm -f "${ROOT_MOUNT_DIR}/usr/local/bin/cpa_sample_code.ko"
}
