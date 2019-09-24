# Copyright 2019 Intel Corporation, Inc
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

Installation
============

Installing the Chart
--------------------

Install the ISTIO Gateway chart for multicloud-k8s application using:

```bash
$ helm install gateways
```
Install the ISTIO VirtualService chart for multicloud-k8s application using:

```bash
$ helm install virtualservices
```
Verify the connectivity to the application/service as shown below
This is an example for Multicloud-k8s application

```bash
$ curl -v <IP Address of ISTIO Ingressgateway>/multicloud-k8s/healthcheck
```
The output like below shows that configuration is successful

```bash
Accept: */*

HTTP/1.1 200 OK
```
