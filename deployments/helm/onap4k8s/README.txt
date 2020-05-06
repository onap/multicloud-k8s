# Copyright ? 2019 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#################################################################
# Installation of ONAP4K8S helm chart
#################################################################

1. Create a helm repo (onap4k8s) from Makefile
$ make repo

1. Run "Makefile" in ONAP4K8S repo
$ make all

2. Deploy the generated Chart
$ helm install dist/packages/multicloud-k8s-5.0.0.tgz
