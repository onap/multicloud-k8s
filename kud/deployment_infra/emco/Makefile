# Copyright © 2017 Amdocs, Bell Canada
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

ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
OUTPUT_DIR := $(ROOT_DIR)/output
PACKAGE_DIR := $(OUTPUT_DIR)/packages

ADDONS := multus-cni ovn4nfv node-feature-discovery sriov-network-operator sriov-network qat-device-plugin cpu-manager

.PHONY: $(ADDONS)

all: $(ADDONS)
	@cp composite-app.yaml $(OUTPUT_DIR)
	@cp values.yaml.example $(OUTPUT_DIR)
	@cp values-resources.yaml.example $(OUTPUT_DIR)

$(ADDONS):
	@echo "\n[$@]"
	@make chart-$@
	@make profile-$@

dep-%:
	@if grep "^dependencies:" ../helm/$*/Chart.yaml; then helm dep up ../helm/$*; fi

lint-%: dep-%
	@helm lint ../helm/$*

chart-%: lint-%
	@mkdir -p $(PACKAGE_DIR)
	@tar -czf $(PACKAGE_DIR)/$*.tar.gz -C ../helm $*

profile-%:
	@mkdir -p $(PACKAGE_DIR)
	@tar -czf $(PACKAGE_DIR)/$*_profile.tar.gz -C ../profiles/$* .

clean:
	@rm -rf $(OUTPUT_DIR)

%:
	@:
