# Copyright 2018 Intel Corporation.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#     http://www.apache.org/licenses/LICENSE-2.0
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Sample Commands:

* POST
    URL:`localhost:8081/v1/vnf_instances/cloudregion1/namespacetest`
    Request Body:

.. code-block:: json

    {
        "cloud_region_id": "region1",
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

Expected Response:

.. code-block:: json

    {
        "response": "Created Deployment:nginx-deployment"
    }

The above POST request will download the following YAML file and run it on the Kubernetes cluster.

.. code-block:: yaml

    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: nginx-deployment
    labels:
        app: nginx
    spec:
    replicas: 3
    selector:
        matchLabels:
        app: nginx
    template:
        metadata:
        labels:
            app: nginx
        spec:
        containers:
        - name: nginx
            image: nginx:1.7.9
            ports:
            - containerPort: 80

* GET
    URL: `localhost:8081/v1/vnf_instances`
