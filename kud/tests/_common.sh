#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

packetgen_deployment_name=packetgen
sink_deployment_name=sink
firewall_deployment_name=firewall
image_name=virtlet.cloud/ubuntu/16.04
multus_deployment_name=multus-deployment
virtlet_image=virtlet.cloud/fedora
virtlet_deployment_name=virtlet-deployment
plugin_deployment_name=plugin-deployment
plugin_service_name=plugin-service
ovn4nfv_deployment_name=ovn4nfv-deployment
onap_private_net=onap-private-net
unprotected_private_net=unprotected-private-net
protected_private_net=protected-private-net
ovn_multus_network_name=ovn-networkobj
rbd_metadata=rbd_metatada.json
rbp_metadata=rbp_metatada.json
rbp_instance=rbp_instance.json

# vFirewall vars
demo_artifacts_version=1.5.0
vfw_private_ip_0='192.168.10.3'
vfw_private_ip_1='192.168.20.2'
vfw_private_ip_2='10.10.100.3'
vpg_private_ip_0='192.168.10.2'
vpg_private_ip_1='10.0.100.2'
vsn_private_ip_0='192.168.20.3'
vsn_private_ip_1='10.10.100.4'
dcae_collector_ip='10.0.4.1'
dcae_collector_port='8081'
protected_net_gw='192.168.20.100'
protected_net_cidr='192.168.20.0/24'
protected_private_net_cidr='192.168.10.0/24'
onap_private_net_cidr='10.10.0.0/16'
sink_ipaddr='192.168.20.250'

# populate_CSAR_containers_vFW() - This function creates the content of CSAR file
# required for vFirewal using only containers
function populate_CSAR_containers_vFW {
    local csar_id=$1

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  network:
    - $unprotected_private_net.yaml
    - $protected_private_net.yaml
    - $onap_private_net.yaml
  deployment:
    - $packetgen_deployment_name.yaml
    - $firewall_deployment_name.yaml
    - $sink_deployment_name.yaml
META

    cat << NET > $unprotected_private_net.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $unprotected_private_net
spec:
  config: '{
    "name": "unprotected",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "$protected_private_net_cidr"
    }
}'
NET

    cat << NET > $protected_private_net.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $protected_private_net
spec:
  config: '{
    "name": "protected",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "$protected_net_cidr"
    }
}'
NET

    cat << NET > $onap_private_net.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $onap_private_net
spec:
  config: '{
    "name": "onap",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "$onap_private_net_cidr"
    }
}'
NET

    cat << DEPLOYMENT > $packetgen_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $packetgen_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
  template:
    metadata:
      labels:
        app: vFirewall
      annotations:
        k8s.v1.cni.cncf.io/networks: '[
            { "name": "$unprotected_private_net", "interfaceRequest": "eth1" },
            { "name": "$onap_private_net", "interfaceRequest": "eth2" }
        ]'
    spec:
      containers:
      - name: $packetgen_deployment_name
        image: electrocucaracha/packetgen
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        resources:
          limits:
            memory: 256Mi
DEPLOYMENT

    cat << DEPLOYMENT > $firewall_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $firewall_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
  template:
    metadata:
      labels:
        app: vFirewall
      annotations:
        k8s.v1.cni.cncf.io/networks: '[
            { "name": "$unprotected_private_net", "interfaceRequest": "eth1" },
            { "name": "$protected_private_net", "interfaceRequest": "eth2" },
            { "name": "$onap_private_net", "interfaceRequest": "eth3" }
        ]'
    spec:
      containers:
      - name: $firewall_deployment_name
        image: electrocucaracha/firewall
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
DEPLOYMENT

    cat << DEPLOYMENT > $sink_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $sink_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
      context: darkstat
  template:
    metadata:
      labels:
        app: vFirewall
        context: darkstat
      annotations:
        k8s.v1.cni.cncf.io/networks: '[
            { "name": "$protected_private_net", "interfaceRequest": "eth1" },
            { "name": "$onap_private_net", "interfaceRequest": "eth2" }
        ]'
    spec:
      containers:
      - name: $sink_deployment_name
        image: electrocucaracha/sink
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        securityContext:
          privileged: true
      - name: darkstat
        image: electrocucaracha/darkstat
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        ports:
          - containerPort: 667
DEPLOYMENT
    popd
}

# populate_CSAR_vms_containers_vFW() - This function creates the content of CSAR file
# required for vFirewal using an hybrid combination between virtual machines and
# cotainers
function populate_CSAR_vms_containers_vFW {
    local csar_id=$1
    ssh_key=$(cat $HOME/.ssh/id_rsa.pub)

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  network:
    - onap-ovn4nfvk8s-network.yaml
  onapNetwork:
    - $unprotected_private_net.yaml
    - $protected_private_net.yaml
    - $onap_private_net.yaml
  deployment:
    - $packetgen_deployment_name.yaml
    - $firewall_deployment_name.yaml
    - $sink_deployment_name.yaml
  service:
    - sink-service.yaml
META

    cat << SERVICE > sink-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: sink-service
spec:
  type: NodePort
  ports:
  - port: 667
    nodePort: 30667
  selector:
    app: vFirewall
    context: darkstat
SERVICE

    cat << MULTUS_NET > onap-ovn4nfvk8s-network.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $ovn_multus_network_name
spec:
  config: '{
      "cniVersion": "0.3.1",
      "name": "ovn4nfv-k8s-plugin",
      "type": "ovn4nfvk8s-cni"
    }'
MULTUS_NET

    cat << NET > $unprotected_private_net.yaml
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: Network

metadata:
  name: $unprotected_private_net
spec:
  cniType : ovn4nfv
  ipv4Subnets:
  - subnet: $protected_private_net_cidr
    name: subnet1
    gateway: 192.168.10.1/24
NET

    cat << NET > $protected_private_net.yaml
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: Network
metadata:
  name: $protected_private_net
spec:
  cniType : ovn4nfv
  ipv4Subnets:
  - subnet: $protected_net_cidr
    name: subnet1
    gateway: $protected_net_gw/24
NET

    cat << NET > $onap_private_net.yaml
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: Network
metadata:
  name: $onap_private_net
spec:
  cniType : ovn4nfv
  ipv4Subnets:
  - subnet: $onap_private_net_cidr
    name: subnet1
    gateway: 10.10.0.1/16
NET

    proxy="apt:"
    cloud_init_proxy="
            - export demo_artifacts_version=$demo_artifacts_version
            - export vfw_private_ip_0=$vfw_private_ip_0
            - export vsn_private_ip_0=$vsn_private_ip_0
            - export protected_net_cidr=$protected_net_cidr
            - export dcae_collector_ip=$dcae_collector_ip
            - export dcae_collector_port=$dcae_collector_port
            - export protected_net_gw=$protected_net_gw
            - export protected_private_net_cidr=$protected_private_net_cidr
            - export sink_ipaddr=$sink_ipaddr
"
    if [[ -n "${http_proxy+x}" ]]; then
        proxy+="
            http_proxy: $http_proxy"
        cloud_init_proxy+="
            - export http_proxy=$http_proxy"
    fi
    if [[ -n "${https_proxy+x}" ]]; then
        proxy+="
            https_proxy: $https_proxy"
        cloud_init_proxy+="
            - export https_proxy=$https_proxy"
    fi
    if [[ -n "${no_proxy+x}" ]]; then
        cloud_init_proxy+="
            - export no_proxy=$no_proxy"
    fi

    cat << DEPLOYMENT > $packetgen_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $packetgen_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
  template:
    metadata:
      labels:
        app: vFirewall
      annotations:
        VirtletLibvirtCPUSetting: |
          mode: host-model
        VirtletCloudInitUserData: |
          ssh_pwauth: True
          users:
          - name: admin
            gecos: User
            primary-group: admin
            groups: users
            sudo: ALL=(ALL) NOPASSWD:ALL
            lock_passwd: false
            # the password is "admin"
            passwd: "\$6\$rounds=4096\$QA5OCKHTE41\$jRACivoPMJcOjLRgxl3t.AMfU7LhCFwOWv2z66CQX.TSxBy50JoYtycJXSPr2JceG.8Tq/82QN9QYt3euYEZW/"
            ssh_authorized_keys:
              $ssh_key
          $proxy
          runcmd:
          $cloud_init_proxy
            - wget -O - https://git.onap.org/multicloud/k8s/plain/kud/tests/vFW/$packetgen_deployment_name | sudo -E bash
        VirtletSSHKeys: |
          $ssh_key
        VirtletRootVolumeSize: 5Gi
        k8s.v1.cni.cncf.io/networks: '[{ "name": "$ovn_multus_network_name"}]'
        k8s.plugin.opnfv.org/nfn-network: '{ "type": "ovn4nfv", "interface": [
            { "name": "$unprotected_private_net", "ipAddress": "$vpg_private_ip_0", "interface": "eth1" , "defaultGateway": "false"},
            { "name": "$onap_private_net", "ipAddress": "$vpg_private_ip_1", "interface": "eth2" , "defaultGateway": "false"}
        ]}'
        kubernetes.io/target-runtime: virtlet.cloud
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: extraRuntime
                operator: In
                values:
                - virtlet
      containers:
      - name: $packetgen_deployment_name
        image: $image_name
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        ports:
          - containerPort: 8183
        resources:
          limits:
            memory: 4Gi
DEPLOYMENT

    cat << DEPLOYMENT > $firewall_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $firewall_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
  template:
    metadata:
      labels:
        app: vFirewall
      annotations:
        VirtletLibvirtCPUSetting: |
          mode: host-model
        VirtletCloudInitUserData: |
          ssh_pwauth: True
          users:
          - name: admin
            gecos: User
            primary-group: admin
            groups: users
            sudo: ALL=(ALL) NOPASSWD:ALL
            lock_passwd: false
            # the password is "admin"
            passwd: "\$6\$rounds=4096\$QA5OCKHTE41\$jRACivoPMJcOjLRgxl3t.AMfU7LhCFwOWv2z66CQX.TSxBy50JoYtycJXSPr2JceG.8Tq/82QN9QYt3euYEZW/"
            ssh_authorized_keys:
              $ssh_key
          $proxy
          runcmd:
            $cloud_init_proxy
            - wget -O - https://git.onap.org/multicloud/k8s/plain/kud/tests/vFW/$firewall_deployment_name | sudo -E bash
        VirtletSSHKeys: |
          $ssh_key
        VirtletRootVolumeSize: 5Gi
        k8s.v1.cni.cncf.io/networks: '[{ "name": "$ovn_multus_network_name"}]'
        k8s.plugin.opnfv.org/nfn-network: '{ "type": "ovn4nfv", "interface": [
            { "name": "$unprotected_private_net", "ipAddress": "$vfw_private_ip_0", "interface": "eth1" , "defaultGateway": "false"},
            { "name": "$protected_private_net", "ipAddress": "$vfw_private_ip_1", "interface": "eth2", "defaultGateway": "false" },
            { "name": "$onap_private_net", "ipAddress": "$vfw_private_ip_2", "interface": "eth3" , "defaultGateway": "false"}
        ]}'
        kubernetes.io/target-runtime: virtlet.cloud
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: extraRuntime
                operator: In
                values:
                - virtlet
      containers:
      - name: $firewall_deployment_name
        image: $image_name
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        resources:
          limits:
            memory: 4Gi
DEPLOYMENT

    cat << CONFIGMAP > sink_configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: sink-configmap
data:
  protected_net_gw: $protected_net_gw
  protected_private_net_cidr: $protected_private_net_cidr
CONFIGMAP

    cat << DEPLOYMENT > $sink_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $sink_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
      context: darkstat
  template:
    metadata:
      labels:
        app: vFirewall
        context: darkstat
      annotations:
        k8s.v1.cni.cncf.io/networks: '[{ "name": "$ovn_multus_network_name"}]'
        k8s.plugin.opnfv.org/nfn-network: '{ "type": "ovn4nfv", "interface": [
            { "name": "$protected_private_net", "ipAddress": "$vsn_private_ip_0", "interface": "eth1", "defaultGateway": "false" },
            { "name": "$onap_private_net", "ipAddress": "$vsn_private_ip_1", "interface": "eth2" , "defaultGateway": "false"}
        ]}'
    spec:
      containers:
      - name: $sink_deployment_name
        image: rtsood/onap-vfw-demo-sink:0.2.0
        envFrom:
        - configMapRef:
            name: sink-configmap
        imagePullPolicy: Always
        tty: true
        stdin: true
        securityContext:
          privileged: true

      - name: darkstat
        image: electrocucaracha/darkstat
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        ports:
          - containerPort: 667
DEPLOYMENT
    popd
}

# populate_CSAR_vms_vFW() - This function creates the content of CSAR file
# required for vFirewal using only virtual machines
function populate_CSAR_vms_vFW {
    local csar_id=$1
    ssh_key=$(cat $HOME/.ssh/id_rsa.pub)

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  network:
    - $unprotected_private_net.yaml
    - $protected_private_net.yaml
    - $onap_private_net.yaml
  deployment:
    - $packetgen_deployment_name.yaml
    - $firewall_deployment_name.yaml
    - $sink_deployment_name.yaml
META

    cat << NET > $unprotected_private_net.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $unprotected_private_net
spec:
  config: '{
    "name": "unprotected",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "$protected_private_net_cidr"
    }
}'
NET

    cat << NET > $protected_private_net.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $protected_private_net
spec:
  config: '{
    "name": "protected",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "$protected_net_cidr"
    }
}'
NET

    cat << NET > $onap_private_net.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $onap_private_net
spec:
  config: '{
    "name": "onap",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "$onap_private_net_cidr"
    }
}'
NET

    proxy="apt:"
    cloud_init_proxy="
            - export demo_artifacts_version=$demo_artifacts_version
            - export vfw_private_ip_0=$vfw_private_ip_0
            - export vsn_private_ip_0=$vsn_private_ip_0
            - export protected_net_cidr=$protected_net_cidr
            - export dcae_collector_ip=$dcae_collector_ip
            - export dcae_collector_port=$dcae_collector_port
            - export protected_net_gw=$protected_net_gw
            - export protected_private_net_cidr=$protected_private_net_cidr
"
    if [[ -n "${http_proxy+x}" ]]; then
        proxy+="
            http_proxy: $http_proxy"
        cloud_init_proxy+="
            - export http_proxy=$http_proxy"
    fi
    if [[ -n "${https_proxy+x}" ]]; then
        proxy+="
            https_proxy: $https_proxy"
        cloud_init_proxy+="
            - export https_proxy=$https_proxy"
    fi
    if [[ -n "${no_proxy+x}" ]]; then
        cloud_init_proxy+="
            - export no_proxy=$no_proxy"
    fi

    cat << DEPLOYMENT > $packetgen_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $packetgen_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
  template:
    metadata:
      labels:
        app: vFirewall
      annotations:
        VirtletLibvirtCPUSetting: |
          mode: host-model
        VirtletCloudInitUserData: |
          ssh_pwauth: True
          users:
          - name: admin
            gecos: User
            primary-group: admin
            groups: users
            sudo: ALL=(ALL) NOPASSWD:ALL
            lock_passwd: false
            # the password is "admin"
            passwd: "\$6\$rounds=4096\$QA5OCKHTE41\$jRACivoPMJcOjLRgxl3t.AMfU7LhCFwOWv2z66CQX.TSxBy50JoYtycJXSPr2JceG.8Tq/82QN9QYt3euYEZW/"
            ssh_authorized_keys:
              $ssh_key
          $proxy
          runcmd:
          $cloud_init_proxy
            - wget -O - https://git.onap.org/multicloud/k8s/plain/kud/tests/vFW/$packetgen_deployment_name | sudo -E bash
        VirtletSSHKeys: |
          $ssh_key
        VirtletRootVolumeSize: 5Gi
        k8s.v1.cni.cncf.io/networks: '[
            { "name": "$unprotected_private_net", "interfaceRequest": "eth1" },
            { "name": "$onap_private_net", "interfaceRequest": "eth2" }
        ]'
        kubernetes.io/target-runtime: virtlet.cloud
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: extraRuntime
                operator: In
                values:
                - virtlet
      containers:
      - name: $packetgen_deployment_name
        image: $image_name
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        ports:
          - containerPort: 8183
        resources:
          limits:
            memory: 4Gi
DEPLOYMENT

    cat << DEPLOYMENT > $firewall_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $firewall_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
  template:
    metadata:
      labels:
        app: vFirewall
      annotations:
        VirtletLibvirtCPUSetting: |
          mode: host-model
        VirtletCloudInitUserData: |
          ssh_pwauth: True
          users:
          - name: admin
            gecos: User
            primary-group: admin
            groups: users
            sudo: ALL=(ALL) NOPASSWD:ALL
            lock_passwd: false
            # the password is "admin"
            passwd: "\$6\$rounds=4096\$QA5OCKHTE41\$jRACivoPMJcOjLRgxl3t.AMfU7LhCFwOWv2z66CQX.TSxBy50JoYtycJXSPr2JceG.8Tq/82QN9QYt3euYEZW/"
            ssh_authorized_keys:
              $ssh_key
          $proxy
          runcmd:
            $cloud_init_proxy
            - wget -O - https://git.onap.org/multicloud/k8s/plain/kud/tests/vFW/$firewall_deployment_name | sudo -E bash
        VirtletSSHKeys: |
          $ssh_key
        VirtletRootVolumeSize: 5Gi
        k8s.v1.cni.cncf.io/networks: '[
            { "name": "$unprotected_private_net", "interfaceRequest": "eth1" },
            { "name": "$protected_private_net", "interfaceRequest": "eth2" },
            { "name": "$onap_private_net", "interfaceRequest": "eth3" }
        ]'
        kubernetes.io/target-runtime: virtlet.cloud
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: extraRuntime
                operator: In
                values:
                - virtlet
      containers:
      - name: $firewall_deployment_name
        image: $image_name
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        resources:
          limits:
            memory: 4Gi
DEPLOYMENT

    cat << DEPLOYMENT > $sink_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $sink_deployment_name
  labels:
    app: vFirewall
spec:
  replicas: 1
  selector:
    matchLabels:
      app: vFirewall
  template:
    metadata:
      labels:
        app: vFirewall
      annotations:
        VirtletLibvirtCPUSetting: |
          mode: host-model
        VirtletCloudInitUserData: |
          ssh_pwauth: True
          users:
          - name: admin
            gecos: User
            primary-group: admin
            groups: users
            sudo: ALL=(ALL) NOPASSWD:ALL
            lock_passwd: false
            # the password is "admin"
            passwd: "\$6\$rounds=4096\$QA5OCKHTE41\$jRACivoPMJcOjLRgxl3t.AMfU7LhCFwOWv2z66CQX.TSxBy50JoYtycJXSPr2JceG.8Tq/82QN9QYt3euYEZW/"
            ssh_authorized_keys:
              $ssh_key
          $proxy
          runcmd:
            $cloud_init_proxy
            - wget -O - https://git.onap.org/multicloud/k8s/plain/kud/tests/vFW/$sink_deployment_name | sudo -E bash
        VirtletSSHKeys: |
          $ssh_key
        VirtletRootVolumeSize: 5Gi
        k8s.v1.cni.cncf.io/networks: '[
            { "name": "$protected_private_net", "interfaceRequest": "eth1" },
            { "name": "$onap_private_net", "interfaceRequest": "eth2" }
        ]'
        kubernetes.io/target-runtime: virtlet.cloud
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: extraRuntime
                operator: In
                values:
                - virtlet
      containers:
      - name: $sink_deployment_name
        image: $image_name
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        ports:
          - containerPort: 667
        resources:
          limits:
            memory: 4Gi
DEPLOYMENT
    popd
}

# populate_CSAR_multus() - This function creates the content of CSAR file
# required for testing Multus feature
function populate_CSAR_multus {
    local csar_id=$1

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  network:
    - bridge-network.yaml
  deployment:
    - $multus_deployment_name.yaml
META

    cat << NET > bridge-network.yaml
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
        "subnet": "$onap_private_net_cidr"
    }
}'
NET

    cat << DEPLOYMENT > $multus_deployment_name.yaml
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
        k8s.v1.cni.cncf.io/networks: '[
          { "name": "bridge-conf", "interfaceRequest": "eth1" },
          { "name": "bridge-conf", "interfaceRequest": "eth2" }
        ]'
    spec:
      containers:
      - name: $multus_deployment_name
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT
    popd
}

# populate_CSAR_virtlet() - This function creates the content of CSAR file
# required for testing Virtlet feature
function populate_CSAR_virtlet {
    local csar_id=$1

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  deployment:
    - $virtlet_deployment_name.yaml
META

    cat << DEPLOYMENT > $virtlet_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $virtlet_deployment_name
  labels:
    app: virtlet
spec:
  replicas: 1
  selector:
    matchLabels:
      app: virtlet
  template:
    metadata:
      labels:
        app: virtlet
      annotations:
        VirtletLibvirtCPUSetting: |
          mode: host-passthrough
        # This tells CRI Proxy that this pod belongs to Virtlet runtime
        kubernetes.io/target-runtime: virtlet.cloud
        VirtletCloudInitUserData: |
          ssh_pwauth: True
          users:
          - name: testuser
            gecos: User
            primary-group: testuser
            groups: users
            lock_passwd: false
            shell: /bin/bash
            # the password is "testuser"
            passwd: "\$6\$rounds=4096\$wPs4Hz4tfs\$a8ssMnlvH.3GX88yxXKF2cKMlVULsnydoOKgkuStTErTq2dzKZiIx9R/pPWWh5JLxzoZEx7lsSX5T2jW5WISi1"
            sudo: ALL=(ALL) NOPASSWD:ALL
          runcmd:
            - echo hello world
    spec:
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: extraRuntime
                operator: In
                values:
                - virtlet
      containers:
      - name: $virtlet_deployment_name
        # This specifies the image to use.
        # virtlet.cloud/ prefix is used by CRI proxy, the remaining part
        # of the image name is prepended with https:// and used to download the image
        image: $virtlet_image
        imagePullPolicy: IfNotPresent
        # tty and stdin required for "kubectl attach -t" to work
        tty: true
        stdin: true
        resources:
          limits:
            # This memory limit is applied to the libvirt domain definition
            memory: 160Mi
DEPLOYMENT
    popd
}

# populate_CSAR_plugin()- Creates content used for Plugin functional tests
function populate_CSAR_plugin {
    local csar_id=$1

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  deployment:
    - $plugin_deployment_name.yaml
  service:
    - service.yaml
META

    cat << DEPLOYMENT > $plugin_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $plugin_deployment_name
spec:
  replicas: 1
  selector:
    matchLabels:
      app: plugin
  template:
    metadata:
      labels:
        app: plugin
    spec:
      containers:
      - name: $plugin_deployment_name
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT

    cat << SERVICE > service.yaml
apiVersion: v1
kind: Service
metadata:
  name: $plugin_service_name
spec:
  ports:
  - port: 80
    protocol: TCP
  selector:
    app: sise
SERVICE
    popd
}

# populate_CSAR_ovn4nfv() - Create content used for OVN4NFV functional test
function populate_CSAR_ovn4nfv {
    local csar_id=$1

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  onap_network:
    - ovn-port-net.yaml
    - ovn-priv-net.yaml
  network:
    - onap-ovn4nfvk8s-network.yaml
  deployment:
    - $ovn4nfv_deployment_name.yaml
META

    cat << MULTUS_NET > onap-ovn4nfvk8s-network.yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: $ovn_multus_network_name
spec:
  config: '{
      "cniVersion": "0.3.1",
      "name": "ovn4nfv-k8s-plugin",
      "type": "ovn4nfvk8s-cni"
    }'
MULTUS_NET

    cat << NETWORK > ovn-port-net.yaml
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: Network
metadata:
  name: ovn-port-net
spec:
  cniType : ovn4nfv
  ipv4Subnets:
  - subnet: 172.16.33.0/24
    name: subnet1
    gateway: 172.16.33.1/24
NETWORK

    cat << NETWORK > ovn-priv-net.yaml
apiVersion: k8s.plugin.opnfv.org/v1alpha1
kind: Network
metadata:
  name: ovn-priv-net
spec:
  cniType : ovn4nfv
  ipv4Subnets:
  - subnet: 172.16.44.0/24
    name: subnet1
    gateway: 172.16.44.1/24
NETWORK

    cat << DEPLOYMENT > $ovn4nfv_deployment_name.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: $ovn4nfv_deployment_name
  labels:
    app: ovn4nfv
spec:
  replicas: 1
  selector:
    matchLabels:
      app: ovn4nfv
  template:
    metadata:
      labels:
        app: ovn4nfv
      annotations:
        k8s.v1.cni.cncf.io/networks: '[{ "name": "$ovn_multus_network_name"}]'
        k8s.plugin.opnfv.org/nfn-network: '{ "type": "ovn4nfv", "interface": [{ "name": "ovn-port-net", "interface": "net0" , "defaultGateway": "false"},
                      { "name": "ovn-priv-net", "interface": "net1" , "defaultGateway": "false"}]}'
    spec:
      containers:
      - name: $ovn4nfv_deployment_name
        image: "busybox"
        command: ["top"]
        stdin: true
        tty: true
DEPLOYMENT
    popd
}

# populate_CSAR_rbdefinition() - Function that populates CSAR folder
# for testing resource bundle definition
function populate_CSAR_rbdefinition {
    _checks_args "$1"
    pushd "${CSAR_DIR}/$1"
    print_msg "Create Helm Chart Archives"
    rm -f *.tar.gz
    tar -czf rb_profile.tar.gz -C $test_folder/vnfs/testrb/helm/profile .
    #Creates vault-consul-dev-0.0.0.tgz
    helm package $test_folder/vnfs/testrb/helm/vault-consul-dev --version 0.0.0
    popd
}

# populate_CSAR_edgex_rbdefinition() - Function that populates CSAR folder
# for testing resource bundle definition of edgex scenario
function populate_CSAR_edgex_rbdefinition {
    _checks_args "$1"
    pushd "${CSAR_DIR}/$1"
    print_msg "Create Helm Chart Archives"
    rm -f *.tar.gz
    tar -czf rb_profile.tar.gz -C $test_folder/vnfs/edgex/profile .
    tar -czf rb_definition.tar.gz -C $test_folder/vnfs/edgex/helm edgex
    popd
}

# populate_CSAR_fw_rbdefinition() - Function that populates CSAR folder
# for testing resource bundle definition of firewall scenario
function populate_CSAR_fw_rbdefinition {
    _checks_args "$1"
    pushd "${CSAR_DIR}/$1"
    print_msg "Create Helm Chart Archives for vFirewall"
    rm -f *.tar.gz
    # Reuse profile from the edgeX case as it is an empty profile
    tar -czf rb_profile.tar.gz -C $test_folder/vnfs/edgex/profile .
    tar -czf rb_definition.tar.gz -C $test_folder/../demo firewall
    popd
}

