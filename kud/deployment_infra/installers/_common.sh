#!/bin/bash

set -o errexit
set -o pipefail
set -u

ROOT_MOUNT_DIR="${ROOT_MOUNT_DIR:-/root}"
ROOT_OS_RELEASE="${ROOT_OS_RELEASE:-$ROOT_MOUNT_DIR/etc/os-release}"
KERNEL_SRC_DIR=$(readlink -f "${ROOT_MOUNT_DIR}/lib/modules/$(uname -r)/build")
[[ "${KERNEL_SRC_DIR}" == "${ROOT_MOUNT_DIR}/*" ]] || KERNEL_SRC_DIR="${ROOT_MOUNT_DIR}${KERNEL_SRC_DIR}"
KERNEL_MOD_SIGN_CMD="${KERNEL_MOD_SIGN_CMD:-}"

RETCODE_SUCCESS=0
RETCODE_ERROR=1

_log() {
    local -r prefix="$1"
    shift
    echo "[${prefix}$(date -u "+%Y-%m-%d %H:%M:%S %Z")] ""$*" >&2
}

info() {
    _log "INFO    " "$*"
}

warn() {
    _log "WARNING " "$*"
}

error() {
    _log "ERROR   " "$*"
}

load_etc_os_release() {
    if [[ ! -f "${ROOT_OS_RELEASE}" ]]; then
        error "File ${ROOT_OS_RELEASE} not found, /etc/os-release from host must be mounted"
        exit ${RETCODE_ERROR}
    fi
    . "${ROOT_OS_RELEASE}"
    info "Running on ${NAME} kernel version $(uname -r)"
}
