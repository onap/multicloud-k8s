#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2020
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

# Script name: ./test_project.sh
# Purpose: Verify if POST/DELETE/GET API calls succeed with invalid/null name
# Expected Results: POST api should fail and return code as documented (example:400)

source _test_functions.sh

# TEST-1 Registering null project name
print_msg "Registering project with null project_name"
project_name=""
payload="$(cat <<EOF
{
  "metadata": {
    "name": "${project_name}",
    "description": "${project_description}",
    "userData1": "${userData1}",
    "userData2": "${userData2}"
   }
}
EOF
)"
call_api_negative -d "${payload}" "${base_url}/projects"
if [ $return_status == 400 ] ;then
    print_msg "Test:project-post with null project name. Expected = 400, Actual = $return_status PASSED"
else
    print_msg "Test:project-post with null project name. Expected = 400, Actual = $return_status FAILED"
fi

# TEST-2 Delete a null project 
project_name=""
print_msg "Deleting ${project_name}"
delete_resource_negative "${base_url}/projects/${project_name}"
if [ $return_status == 404 ] ;then
    print_msg "Test:project-delete-1 with null project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:project-delete-1 with null project name. Expected = 404, Actual = $return_status FAILED"
fi

# TEST-3 Delete a non existing project
project_name="foo"
print_msg "Deleting ${project_name}"
delete_resource_negative "${base_url}/projects/${project_name}"
if [ $return_status == 404 ] ;then
    print_msg "Test:project-delete-2 with invalid project name. Expected = 404, Actual = $return_status PASSED"
else
    print_msg "Test:project-delete-2 with invalid project name. Expected = 404, Actual = $return_status FAILED"
fi

# TEST-4 Get an invalid project 
project_name="foo"
get_resource_negative "${base_url}/projects/${project_name}"
if [ $return_status == 404 ] ;then
    print_msg "Test:project-get with null project name. Expected = 404, \
    Actual = $return_status PASSED"
else
    print_msg "Test:project-get with null project name. Expected = 404, \
    Actual = $return_status FAILED"
fi

# END
