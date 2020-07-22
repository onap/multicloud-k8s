
# /*
#  * Copyright 2020 Intel Corporation, Inc
#  *
#  * Licensed under the Apache License, Version 2.0 (the "License");
#  * you may not use this file except in compliance with the License.
#  * You may obtain a copy of the License at
#  *
#  *     http://www.apache.org/licenses/LICENSE-2.0
#  *
#  * Unless required by applicable law or agreed to in writing, software
#  * distributed under the License is distributed on an "AS IS" BASIS,
#  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  * See the License for the specific language governing permissions and
#  * limitations under the License.
#  */

# Script to run all the tests at once
# Aditya Sharoff<aditya.sharoff@intel.com> 07/14/2020

./test_all_intents.sh
./test_composite_app.sh
./test_cost_based_controller.sh
./test_deployment_intent_group.sh
./test_generic_placement_intent_app.sh
./test_generic_placement_intent.sh
./test_HPA.sh
./test_multipart.sh
./test_OVN.sh
./test_profile_apps.sh
./test_profile.sh
./test_project.sh
./test_traffic_controller.sh
