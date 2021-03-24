#!/bin/bash

#set -x
source _common.sh
source _qat-driver-installer.sh

# IMPORTANT: If the driver version is changed, review the QAT Makefile
# against _qat.sh.  The steps in _qat.sh are from the Makefile and
# have been modified to run inside a container.
QAT_DRIVER_VERSION="${QAT_DRIVER_VERSION:-1.7.l.4.12.0-00011}"
QAT_DRIVER_DOWNLOAD_URL_DEFAULT="https://01.org/sites/default/files/downloads/qat${QAT_DRIVER_VERSION}.tar.gz"
QAT_DRIVER_DOWNLOAD_URL="${QAT_DRIVER_DOWNLOAD_URL:-$QAT_DRIVER_DOWNLOAD_URL_DEFAULT}"
QAT_DRIVER_ARCHIVE="$(basename "${QAT_DRIVER_DOWNLOAD_URL}")"
QAT_INSTALL_DIR_HOST="${QAT_INSTALL_DIR_HOST:-/opt/qat}"
QAT_INSTALL_DIR_CONTAINER="${QAT_INSTALL_DIR_CONTAINER:-/usr/local/qat}"
QAT_ENABLE_SRIOV="${QAT_ENABLE_SRIOV:-host}"
CACHE_FILE="${QAT_INSTALL_DIR_CONTAINER}/.cache"

check_kernel_boot_parameter() {
    if [[ $(grep -c intel_iommu=on /proc/cmdline) != "0" ]]; then
        info "Found intel_iommu=on kernel boot parameter"
    else
        error "Missing intel_iommu=on kernel boot parameter"
        exit "${RETCODE_ERROR}"
    fi
}

check_sriov_hardware_capabilities() {
    if [[ $(lspci -vn -d 8086:0435 | grep -c SR-IOV) != "0" ]]; then
        info "Found dh895xcc SR-IOV hardware capabilities"
    elif [[ $(lspci -vn -d 8086:37c8 | grep -c SR-IOV) != "0" ]]; then
        info "Found c6xx SR-IOV hardware capabilities"
    elif [[ $(lspci -vn -d 8086:6f54 | grep -c SR-IOV) != "0" ]]; then
        info "Found d15xx SR-IOV hardware capabilities"
    elif [[ $(lspci -vn -d 8086:19e2 | grep -c SR-IOV) != "0" ]]; then
        info "Found c3xxx SR-IOV hardware capabilities"
    else
        error "Missing SR-IOV hardware capabilities"
        exit "${RETCODE_ERROR}"
    fi
}

download_qat_src() {
    info "Downloading QAT source ... "
    mkdir -p "${QAT_INSTALL_DIR_CONTAINER}"
    pushd "${QAT_INSTALL_DIR_CONTAINER}" > /dev/null
    curl -L -sS "${QAT_DRIVER_DOWNLOAD_URL}" -o "${QAT_DRIVER_ARCHIVE}"
    tar xf "${QAT_DRIVER_ARCHIVE}"
    popd > /dev/null
}

build_qat_src() {
    info "Building QAT source ... "
    pushd "${QAT_INSTALL_DIR_CONTAINER}" > /dev/null
    KERNEL_SOURCE_ROOT="${KERNEL_SRC_DIR}" ./configure --enable-icp-sriov="${QAT_ENABLE_SRIOV}"
    make
    popd > /dev/null
}

install_qat() {
    check_kernel_boot_parameter
    check_sriov_hardware_capabilities
    download_qat_src
    build_qat_src
    _qat_driver_install
    _adf_ctl_install
    _qat_service_install
}

uninstall_qat() {
    _adf_ctl_uninstall
    _qat_service_shutdown
    _qat_service_uninstall
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
        if [[ "${QAT_DRIVER_VERSION}" == "${CACHE_QAT_DRIVER_VERSION}" ]]; then
            info "Found existing driver installation for kernel version $(uname -r) and driver version ${QAT_DRIVER_VERSION}"
            return "${RETCODE_SUCCESS}"
        fi
    fi
    return "${RETCODE_ERROR}"
}

update_cached_version() {
    cat >"${CACHE_FILE}"<<__EOF__
CACHE_KERNEL_VERSION=$(uname -r)
CACHE_QAT_DRIVER_VERSION=${QAT_DRIVER_VERSION}
__EOF__

    info "Updated cached version as:"
    cat "${CACHE_FILE}"
}

upgrade_driver() {
    uninstall_qat
    install_qat
}

check_driver_started() {
    _qat_check_started
}

start_driver() {
    _qat_service_start
    _qat_check_started
}

uninstall_driver() {
    uninstall_qat
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
        install-sample)
            _qat_sample_install
            ;;
        uninstall-sample)
            _qat_sample_uninstall
            ;;
    esac
}

main "$@"
