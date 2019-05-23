#!/bin/bash
#
# This is simple wrapper script for executing all-in-one kud deployments
# from community CI jenkins jobs. Main purpose of it is to keep control
# over future changes within this single repo
#

# setting-up bash flags
set -x -e -o pipefail

# boilerplate for getting correct relative path
curr_dir=$(cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd)

cd ${curr_dir}/../hosting_providers/vagrant

# trigger all-in-one deployment from this CI wrapper script
./aio.sh
