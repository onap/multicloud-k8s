#!/bin/bash
# SPDX-license-identifier: Apache-2.0
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

wget -O - https://raw.githubusercontent.com/akraino-edge-stack/icn/master/cmd/bpa-operator/e2etest/bpa_virtletvm_verifier.sh | sudo -E bash
