#!/bin/bash

#set -x
source _common.sh
SCRIPT_DIR=$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")

IAVF_DRIVER_VERSION="${IAVF_DRIVER_VERSION:-4.0.2}"
IAVF_DRIVER_DOWNLOAD_URL_DEFAULT="https://downloadmirror.intel.com/30305/eng/iavf-${IAVF_DRIVER_VERSION}.tar.gz"
IAVF_DRIVER_DOWNLOAD_URL="${IAVF_DRIVER_DOWNLOAD_URL:-$IAVF_DRIVER_DOWNLOAD_URL_DEFAULT}"
IAVF_DRIVER_ARCHIVE="$(basename "${IAVF_DRIVER_DOWNLOAD_URL}")"
IAVF_INSTALL_DIR_HOST="${IAVF_INSTALL_DIR_HOST:-/opt/iavf}"
IAVF_INSTALL_DIR_CONTAINER="${IAVF_INSTALL_DIR_CONTAINER:-/usr/local/iavf}"
CACHE_FILE="${IAVF_INSTALL_DIR_CONTAINER}/.cache"

check_adapter() {
    local -r nic_models="X710 XL710 X722"
    if [[ $(lspci | grep -c "Ethernet .* \(${nic_models// /\\|}\)") != "0" ]]; then
        info "Found adapter"
    else
        error "Missing adapter"
        exit "${RETCODE_ERROR}"
    fi
}

download_iavf_src() {
    info "Downloading IAVF source ... "
    mkdir -p "${IAVF_INSTALL_DIR_CONTAINER}"
    pushd "${IAVF_INSTALL_DIR_CONTAINER}" > /dev/null
    curl -L -sS "${IAVF_DRIVER_DOWNLOAD_URL}" -o "${IAVF_DRIVER_ARCHIVE}"
    tar xf "${IAVF_DRIVER_ARCHIVE}" --strip-components=1
    info "Patching IAVF source ... "
    # Ubuntu 18.04 added the skb_frag_off definitions to the kernel
    # headers beginning with 4.15.0-159
    patch -p1 < "${SCRIPT_DIR}/skb-frag-off.patch"
    popd > /dev/null
}

build_iavf_src() {

    info "Building IAVF source ... "
    pushd "${IAVF_INSTALL_DIR_CONTAINER}/src" > /dev/null
    KSRC=${KERNEL_SRC_DIR} SYSTEM_MAP_FILE="${ROOT_MOUNT_DIR}/boot/System.map-$(uname -r)" INSTALL_MOD_PATH="${ROOT_MOUNT_DIR}" make install
    # TODO Unable to update initramfs. You may need to do this manaully.
    popd > /dev/null
}

install_iavf() {
    check_adapter
    download_iavf_src
    build_iavf_src
}

uninstall_iavf() {
    if [[ $(lsmod | grep -c "iavf") != "0" ]]; then
        rmmod iavf
    fi
    if [[ $(lsmod | grep -c "i40evf") != "0" ]]; then
        rmmod i40evf
    fi
    if [[ -d "${IAVF_INSTALL_DIR_CONTAINER}/src" ]]; then
        pushd "${IAVF_INSTALL_DIR_CONTAINER}/src" > /dev/null
        KSRC=${KERNEL_SRC_DIR} SYSTEM_MAP_FILE="${ROOT_MOUNT_DIR}/boot/System.map-$(uname -r)" INSTALL_MOD_PATH="${ROOT_MOUNT_DIR}" make uninstall
        popd > /dev/null
    fi
    # This is a workaround for missing INSTALL_MOD_PATH prefix in the Makefile:
    rm -f "${ROOT_MOUNT_DIR}/etc/modprobe.d/iavf.conf"
}

check_cached_version() {
    info "Checking cached version"
    if [[ ! -f "${CACHE_FILE}" ]]; then
        info "Cache file ${CACHE_FILE} not found"
        return "${RETCODE_ERROR}"
    fi
    # Source the cache file and check if the cached driver matches
    # currently running kernel and driver versions.
    . "${CACHE_FILE}"
    if [[ "$(uname -r)" == "${CACHE_KERNEL_VERSION}" ]]; then
        if [[ "${IAVF_DRIVER_VERSION}" == "${CACHE_IAVF_DRIVER_VERSION}" ]]; then
            info "Found existing driver installation for kernel version $(uname -r) and driver version ${IAVF_DRIVER_VERSION}"
            return "${RETCODE_SUCCESS}"
        fi
    fi
    return "${RETCODE_ERROR}"
}

update_cached_version() {
    cat >"${CACHE_FILE}"<<__EOF__
CACHE_KERNEL_VERSION=$(uname -r)
CACHE_IAVF_DRIVER_VERSION=${IAVF_DRIVER_VERSION}
__EOF__

    info "Updated cached version as:"
    cat "${CACHE_FILE}"
}

upgrade_driver() {
    uninstall_iavf
    install_iavf
}

check_driver_started() {
    if [[ $(lsmod | grep -c "iavf") == "0" ]]; then
        return "${RETCODE_ERROR}"
    fi
    return 0
}

start_driver() {
    modprobe -d "${ROOT_MOUNT_DIR}" -C "${ROOT_MOUNT_DIR}/etc/modprobe.d" iavf
    if ! check_driver_started; then
        error "Driver not started"
    fi
}

uninstall_driver() {
    uninstall_iavf
    rm -f "${CACHE_FILE}"
}

main() {
    load_etc_os_release
    local -r cmd="${1:-install}"
    case $cmd in
        install)
            if ! check_cached_version; then
                upgrade_driver
                update_cached_version
            fi
            if ! check_driver_started; then
                start_driver
            fi
            ;;
        uninstall)
            uninstall_driver
            ;;
    esac
}

main "$@"
