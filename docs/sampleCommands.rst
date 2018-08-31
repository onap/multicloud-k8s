.. Copyright 2018 Intel Corporation.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
        http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

====================
Sample cURL commands
====================

****
POST
****

URL: `localhost:8081/v1/vnf_instances/`

Request Body
------------

.. code-block:: json

    {
        "cloud_region_id": "region1",
        "namespace": "test-namespace",
        "csar_id": "uuid",
        "namespace": "test",
        "oof_parameters": [{
            "key1": "value1",
            "key2": "value2",
            "key3": {}
        }],
        "network_parameters": {
            "oam_ip_address": {
                "connection_point": "string",
                "ip_address": "string",
                "workload_name": "string"
            }
        }
    }

Expected Response
-----------------

.. code-block:: json

    {
        "vnf_id": "52fdfc07",
        "cloud_region_id": "cloudregion1",
        "namespace": "test-namespace",
        "vnf_components": {
            "deployment": [
                "cloudregion1-test-namespace-52fdfc07-kubedeployment"
            ],
            "service": [
                "cloudregion1-test-namespace-52fdfc07-kubeservice"
            ]
        }
    }

***
GET
***

URL: `localhost:8081/v1/vnf_instances`

Expected Response
-----------------

.. code-block:: json

    {
        "vnf_id_list": [
            "52fdfc07"
        ]
    }

***
GET
***

URL: `localhost:8081/v1/vnf_instances/cloudregion1/test-namespace/52fdfc07`

Expected Response
-----------------

.. code-block:: json

    {
        "vnf_id": "52fdfc07",
        "cloud_region_id": "cloudregion1",
        "namespace": "test-namespace",
        "vnf_components": {
            "deployment": [
                "cloudregion1-test-namespace-52fdfc07-kubedeployment"
            ],
            "service": [
                "cloudregion1-test-namespace-52fdfc07-kubeservice"
            ]
        }
    }

***
DELETE
***

URL: `localhost:8081/v1/vnf_instances/cloudregion1/test-namespace/52fdfc07`

Expected Response
-----------------

202 Accepted