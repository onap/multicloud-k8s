#!/bin/bash
#
# This is simple wrapper script for executing all-in-one kud deployments
# from community CI jenkins jobs. Main purpose of it is to keep control
# over future changes within this single repo
#

# setting-up bash flags
set -x -e -o pipefail

# boilerplate for geting correct relative path
SCRIPT_DIR=$(dirname "${0}")
LOCAL_PATH=$(readlink -f "$SCRIPT_DIR")

cd "${LOCAL_PATH}"/"${RELATIVE_PATH}"/../hosting_providers/vagrant/

# trigger all-in-one deployment from this CI wrapper script
./aio.sh
