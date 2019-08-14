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


NOTE - A configMap of available IPs is to be applied in order for services
to get external IP address assigned. Please Update values.yaml with
IP addresses before deploying

Prerequisites
-------------

-  Kubernetes 1.9+

Installing the Chart
--------------------

The chart can be installed as follows:

```console
$ helm install --name metallb metallb
```

The command deploys MetalLB on the Kubernetes cluster. This chart does
not provide a default configuration; MetalLB will not act on your
Kubernetes Services until you provide
one. The [configuration](#configuration) section lists various ways to
provide this configuration.

Uninstalling the Chart
----------------------

To uninstall/delete the `metallb` deployment:

```console
$ helm delete metallb
```

The command removes all the Kubernetes components associated with the
chart, but will not remove the release metadata from `helm` — this will prevent
you, for example, if you later try to create a release also named `metallb`). To
fully delete the release and release history, simply [include the `--purge`
flag][helm-usage]:

```console
$ helm delete --purge metallb
```

Configuration
-------------

See `values.yaml` for configuration notes. Specify each parameter
using the `--set key=value[,key=value]` argument to `helm
install`. For example,

```console
$ helm install --name metallb \
  --set rbac.create=false \
    stable/metallb
```

The above command disables the use of RBAC rules.

Alternatively, a YAML file that specifies the values for the above
parameters can be provided while installing the chart. For example,

```console
$ helm install --name metallb -f values.yaml metallb
```

By default, this chart does not install a configuration for MetalLB, and simply
warns you that you must follow [the configuration instructions on MetalLB's
website][metallb-config] to create an appropriate ConfigMap.

**Please note:** By default, this chart expects a ConfigMap named
'metallb-config' within the same namespace as the chart is
deployed. _This is different than the MetalLB documentation_, which
asks you to create a ConfigMap in the `metallb-system` namespace, with
the name of 'config'.

For simple setups that only use MetalLB's [ARP mode][metallb-arpndp-concepts],
you can specify a single IP range using the `arpAddresses` parameter to have the
chart install a working configuration for you:

```console
$ helm install --name metallb \
  --set arpAddresses=192.168.16.240/30 \
  stable/metallb
```

If you have a more complex configuration and want Helm to manage it for you, you
can provide it in the `config` parameter. The configuration format is
[documented on MetalLB's website][metallb-config].

```console
$ cat values.yaml
configInline:
  peers:
  - peer-address: 10.0.0.1
    peer-asn: 64512
    my-asn: 64512
  address-pools:
  - name: default
    protocol: bgp
    addresses:
    - 198.51.100.0/24

$ helm install --name metallb -f values.yaml metallb
```

[helm-home]: https://helm.sh
[helm-usage]: https://docs.helm.sh/using_helm/
[k8s-home]: https://kubernetes.io
[metallb-arpndp-concepts]: https://metallb.universe.tf/concepts/arp-ndp/
[metallb-config]: https://metallb.universe.tf/configuration/
[metallb-home]: https://metallb.universe.tf
