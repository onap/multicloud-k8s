# Kud test case
## Summary
This folder contains all the test scripts. Most of those test scripts can be executed simply by `bash <script-name>`.

## Table of contents
* [Multus CNI](#Multus-CNI)
    1. [define the additional network custom resource definition(CRD)](#Define-custom-resource-definition)
        * [bridge CNI](#bridge-CNI)
        * [macvlan CNI](#macvlan-CNI)
        * [ipvlan CNI](#ipvlan-CNI)
        * [ptp CNI](#ptp-CNI)
    2. [Create a pod with the previously CRD annotation](#Create-a-pod-with-the-previously-CRD-annotation)
    3. [Verify the additional interface was configured](#Verify-the-additional-interface-was-configured)
* [SRIOV plugin](#SRIOV-plugin)
    1. [define SRIOV network CRD](#define-SRIOV-network-CRD)
    2. [Create a pod with single/multiple VF interface](#Create-a-pod-with-single/multiple-VF-interface)
        * [single VF allocated](#single-VF-allocated)
        * [multiple VF allocated](#multiple-VF-allocated)
    3. [Verify the VF interface was allocated](#Verify-the-VF-interface-was-allocated)
* [NFD](#NFD)
    1. [Create a pod to run on particular node](#Create-a-pod-to-run-on-particular-node)
        * [nodeSelector](#nodeSelector)
        * [node affinity](#node-affinity)
    2. [Verify pod created status](#Verify-pod-created-status)
* [CMK](#)

## Multus CNI
[Multus CNI](https://github.com/intel/multus-cni) is a container network interface (CNI) plugin for Kubernetes that enables attaching multiple network interfaces to pods. Typically, in Kubernetes each pod only has one network interface (apart from a loopback) -- with Multus you can create a multi-homed pod that has multiple interfaces. This is accomplished by Multus acting as a "meta-plugin", a CNI plugin that can call multiple other CNI plugins.

### Define custom resource definition
#### bridge CNI

---
##### Overview

With bridge plugin, all containers (on the same host) are plugged into a bridge (virtual switch) that resides in the host network namespace. Please refer to [the bridge cni](https://github.com/containernetworking/plugins/tree/master/plugins/main/bridge) for details.

##### Example configuration

```
    cat << NET | kubectl apply -f -
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: bridge-conf
spec:
  config: '{
    "cniVersion": "0.3.0",
    "name": "mynet",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "$multus_private_net_cidr"
    }
}'
NET
```
##### Network configuration reference

* `name` (string, required): the name of the network.
* `type` (string, required): "bridge".
* `bridge` (string, optional): name of the bridge to use/create. Defaults to "cni0".
* `isGateway` (boolean, optional): assign an IP address to the bridge. Defaults to false.
* `isDefaultGateway` (boolean, optional): Sets isGateway to true and makes the assigned IP the default route. Defaults to false.
* `forceAddress` (boolean, optional): Indicates if a new IP address should be set if the previous value has been changed. Defaults to false.
* `ipMasq` (boolean, optional): set up IP Masquerade on the host for traffic originating from this network and destined outside of it. Defaults to false.
* `mtu` (integer, optional): explicitly set MTU to the specified value. Defaults to the value chosen by the kernel.
* `hairpinMode` (boolean, optional): set hairpin mode for interfaces on the bridge. Defaults to false.
* `ipam` (dictionary, required): IPAM configuration to be used for this network. For L2-only network, create empty dictionary.
* `promiscMode` (boolean, optional): set promiscuous mode on the bridge. Defaults to false.
* `vlan` (int, optional): assign VLAN tag. Defaults to none.

#### macvlan CNI

---

##### Overview

macvlan functions like a switch that is already connected to the host interface.
A host interface gets "enslaved" with the virtual interfaces sharing the physical device but having distinct MAC addresses.
Please refer to [the macvlan cni](https://github.com/containernetworking/plugins/tree/master/plugins/main/macvlan) for details.

##### Example configuration

```
    cat << NET  | kubectl apply -f -
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: macvlan-conf
spec:
  config: '{
    "name": "mynet",
    "type": "macvlan",
    "master": "$master_name",
    "ipam": {
        "type": "host-local",
        "subnet": "$multus_private_net_cidr"
    }
}'
NET
```

##### Network configuration reference

* `name` (string, required): the name of the network
* `type` (string, required): "macvlan"
* `master` (string, optional): name of the host interface to enslave. Defaults to default route interface.
* `mode` (string, optional): one of "bridge", "private", "vepa", "passthru". Defaults to "bridge".
* `mtu` (integer, optional): explicitly set MTU to the specified value. Defaults to the value chosen by the kernel. The value must be \[0, master's MTU\].
* `ipam` (dictionary, required): IPAM configuration to be used for this network. For interface only without ip address, create empty dictionary.

#### ipvlan CNI

---

##### Overview

ipvlan is a new addition to the Linux kernel. It virtualizes the host interface. Please refer to [the ipvlan cni](https://github.com/containernetworking/plugins/tree/master/plugins/main/ipvlan) for details.

##### Example configuration

```
    cat << NET  | kubectl apply -f -
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: ipvlan-conf
spec:
  config: '{
    "name": "mynet",
    "type": "ipvlan",
    "master": "$master_name",
    "ipam": {
        "type": "host-local",
        "subnet": "$multus_private_net_cidr"
    }
}'
NET
```

##### Network configuration reference

* `name` (string, required): the name of the network.
* `type` (string, required): "ipvlan".
* `master` (string, required unless chained): name of the host interface to enslave.
* `mode` (string, optional): one of "l2", "l3", "l3s". Defaults to "l2".
* `mtu` (integer, optional): explicitly set MTU to the specified value. Defaults to the value chosen by the kernel.
* `ipam` (dictionary, required unless chained): IPAM configuration to be used for this network.

#### ptp CNI

---
##### Overview
The ptp plugin creates a point-to-point link between a container and the host by using a veth device.
Please refer to [the ptp cni](https://github.com/containernetworking/plugins/tree/master/plugins/main/ptp) for details.

##### Example network configuration

```
    cat << NET  | kubectl apply -f -
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: ptp-conf
spec:
  config: '{
    "name": "mynet",
    "type": "ptp",
    "ipam": {
        "type": "host-local",
        "subnet": "$multus_private_net_cidr"
    }
}'
NET
```

##### Network configuration reference

* `name` (string, required): the name of the network
* `type` (string, required): "ptp"
* `ipMasq` (boolean, optional): set up IP Masquerade on the host for traffic originating from ip of this network and destined outside of this network. Defaults to false.
* `mtu` (integer, optional): explicitly set MTU to the specified value. Defaults to value chosen by the kernel.
* `ipam` (dictionary, required): IPAM configuration to be used for this network.
* `dns` (dictionary, optional): DNS information to return as described in the result.

#### Create a pod with the previously CRD annotation
```
    cat << DEPLOYMENT | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $multus_deployment_name
  labels:
    app: multus
spec:
  replicas: 1
  selector:
    matchLabels:
      app: multus
  template:
    metadata:
      labels:
        app: multus
      annotations:
        k8s.v1.cni.cncf.io/networks: bridge-conf
    spec:
      containers:
      - name: $multus_deployment_name
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT
```
> You can add more interfaces to a pod by creating more custom resources and then referring to them in pod's annotation. You can also reuse configurations, so for example, to attach a bridge interface and a macvlan interface to a pod, you could create a pod like so:
```
    cat << DEPLOYMENT | kubectl create -f -
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $multus_deployment_name
  labels:
    app: multus
spec:
  replicas: 1
  selector:
    matchLabels:
      app: multus
  template:
    metadata:
      labels:
        app: multus
      annotations:
        k8s.v1.cni.cncf.io/networks: bridge-conf, macvlan-conf
    spec:
      containers:
      - name: $multus_deployment_name
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT
```
> Note that the annotation now reads k8s.v1.cni.cncf.io/networks: bridge-conf, macvlan-conf. Where we have the same configuration used twice, separated by a comma.
 If you were to create another custom resource with the name foo you could use that such as: k8s.v1.cni.cncf.io/networks: foo,macvlan-conf, and use any number of attachments.

#### Verify the additional interface was configured
We can Verify the additional interface by running command as shown below.
```
kubectl exec -it $deployment_pod -- ip a
```
The output should looks like the following.
```
===== multus-deployment-688659b564-79dth details =====
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if543: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1450 qdisc noqueue
    link/ether 0a:58:0a:f4:44:1e brd ff:ff:ff:ff:ff:ff
    inet 10.244.68.30/24 scope global eth0
       valid_lft forever preferred_lft forever
5: net1@if544: <BROADCAST,MULTICAST,UP,LOWER_UP,M-DOWN> mtu 1500 qdisc noqueue
    link/ether 46:9d:68:90:f1:eb brd ff:ff:ff:ff:ff:ff
    inet 10.20.0.12/16 scope global net1
       valid_lft forever preferred_lft forever
```
You should note that a new interface named `net1` is attached.
>For further test information please refer to the file [`./multus.sh`](./multus.sh).

## SRIOV plugin


### define SRIOV network CRD

##### Example CRD configuration

```
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: sriov-conf
  annotations:
    k8s.v1.cni.cncf.io/resourceName: intel.com/intel_sriov_700
spec:
  config: '{
    "type": "sriov",
    "cniVersion": "0.3.1",
    "ipam": {
            "type": "host-local",
            "subnet": "10.56.206.0/24",
            "routes": [
                    { "dst": "0.0.0.0/0" }
            ],
            "gateway": "10.56.206.1"
    }
  }'
```
### Create a pod with single/multiple VF interface

#### single VF allocated
```
cat << POD | kubectl create -f - --validate=false
apiVersion: v1
kind: Pod
metadata:
  name: $deployment_pod
  annotations:
    k8s.v1.cni.cncf.io/networks: sriov-conf
spec:
  containers:
  - name: test-pod
    image: docker.io/centos/tools:latest
    command:
    - /sbin/init
    resources:
      requests:
        intel.com/intel_sriov_700: '1'
      limits:
        intel.com/intel_sriov_700: '1'
POD
```
#### multiple VF allocated
```
cat << POD | kubectl create -f - --validate=false
apiVersion: v1
kind: Pod
metadata:
  name: $deployment_pod
  annotations:
    k8s.v1.cni.cncf.io/networks: sriov-conf, sriov-conf
spec:
  containers:
  - name: test-pod
    image: docker.io/centos/tools:latest
    command:
    - /sbin/init
    resources:
      requests:
        intel.com/intel_sriov_700: '2'
      limits:
        intel.com/intel_sriov_700: '2'
POD
```

### Verify the VF interface was allocated

We can Verify the additional VF interface by running command as shown below.
```
kubectl exec -it $deployment_pod -- ip a
```
The output should looks like the following.
```
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
    inet 127.0.0.1/8 scope host lo
       valid_lft forever preferred_lft forever
3: eth0@if429: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1450 qdisc noqueue state UP group default
    link/ether 0a:58:0a:f4:40:09 brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.244.64.9/24 scope global eth0
       valid_lft forever preferred_lft forever
413: net2: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc mq state DOWN group default qlen 1000
    link/ether 36:57:78:8b:e1:3b brd ff:ff:ff:ff:ff:ff
    inet 10.56.206.4/24 brd 10.56.206.255 scope global net2
       valid_lft forever preferred_lft forever
414: net1: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc mq state DOWN group default qlen 1000
    link/ether de:d0:73:53:08:66 brd ff:ff:ff:ff:ff:ff
    inet 10.56.206.3/24 brd 10.56.206.255 scope global net1
       valid_lft forever preferred_lft forever
```
>For further test information please refer to the file [`./sriov.sh`](./sriov.sh).

## NFD

node feature discovery([NFD](https://github.com/kubernetes-sigs/node-feature-discovery)) detects hardware features available on each node in a Kubernetes cluster, and advertises those features using node labels.
### Create a pod to run on particular node
#### nodeSelector
`nodeSelector` is a field of PodSpec. It specifies a map of key-value pairs. For the pod to be eligible to run on a node, the node must have each of the indicated key-value pairs as labels (it can have additional labels as well). The most common usage is one key-value pair. 
##### To create pod configuration with `nodeSelector`
```
cat << POD > $HOME/$pod_name.yaml | kubectl create -f $HOME/$pod_name.yaml
apiVersion: v1
kind: Pod
metadata:
  name: $pod_name
spec:
  nodeSelector:
    feature.node.kubernetes.io/kernel-version.major: '4'
  containers:
  - name: with-node-affinity
    image: gcr.io/google_containers/pause:2.0
POD
```
#### node affinity
`nodeAffinity` is conceptually similar to `nodeSelector` – it allows you to constrain which nodes your pod is eligible to be scheduled on, based on labels on the node.

##### To create pod configuration with `nodeAffinity`
```
cat << POD > $HOME/$pod_name.yaml | kubectl create -f $HOME/$pod_name.yaml
apiVersion: v1
kind: Pod
metadata:
  name: $pod_name
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchExpressions:
          - key: "feature.node.kubernetes.io/kernel-version.major"
            operator: Gt
            values:
            - '3'
  containers:
  - name: with-node-affinity
    image: gcr.io/google_containers/pause:2.0
POD
```
>For further information on how to configure nodeAffinity `operator` field please refer to the file [`./nfd.sh`](./nfd.sh).

### Verify pod created status
To Verify the nfd pod by running command as shown below.
```
kubectl get pods -A | grep $pod_name
```
If the output shows pod `STATUS` field is `running`, the pod have been scheduled successfully.
