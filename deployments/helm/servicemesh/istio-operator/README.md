/*
 * Copyright 2019 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

## Prerequisites

- Kubernetes 1.10.0+

## Installing the chart

To install the chart from local directory:

```
helm install --name=istio-operator --namespace=istio-system istio-operator
```

## Uninstalling the Chart

To uninstall/delete the `istio-operator` release:

```
$ helm del --purge istio-operator
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Banzaicloud Istio Operator chart and their default values.

Parameter | Description | Default
--------- | ----------- | -------
`operator.image.repository` | Operator container image repository | `banzaicloud/istio-operator`
`operator.image.tag` | Operator container image tag | `0.2.1`
`operator.image.pullPolicy` | Operator container image pull policy | `IfNotPresent`
`operator.resources` | CPU/Memory resource requests/limits (YAML) | Memory: `128Mi/256Mi`, CPU: `100m/200m`
`istioVersion` | Supported Istio version | `1.2`
`prometheusMetrics.enabled` | If true, use direct access for Prometheus metrics | `false`
`prometheusMetrics.authProxy.enabled` | If true, use auth proxy for Prometheus metrics | `true`
`prometheusMetrics.authProxy.image.repository` | Auth proxy container image repository | `gcr.io/kubebuilder/kube-rbac-proxy`
`prometheusMetrics.authProxy.image.tag` | Auth proxy container image tag | `v0.4.0`
`prometheusMetrics.authProxy.image.pullPolicy` | Auth proxy container image pull policy | `IfNotPresent`
`rbac.enabled` | Create rbac service account and roles | `true`
