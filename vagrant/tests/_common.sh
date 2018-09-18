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

# popule_CSAR_containers_vFW() - This function creates the content of CSAR file
# required for vFirewal using only containers
function popule_CSAR_containers_vFW {
    local csar_id=$1

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  network:
    - unprotected-private-net-cidr-network.yaml
    - protected-private-net-cidr-network.yaml
    - onap-private-net-cidr-network.yaml
  deployment:
    - $packetgen_deployment_name.yaml
    - $firewall_deployment_name.yaml
    - $sink_deployment_name.yaml
META

    cat << NET > unprotected-private-net-cidr-network.yaml
apiVersion: "kubernetes.cni.cncf.io/v1"
kind: Network
metadata:
  name: unprotected-private-net-cidr
spec:
  config: '{
    "name": "unprotected",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "192.168.10.0/24"
    }
}'
NET

    cat << NET > protected-private-net-cidr-network.yaml
apiVersion: "kubernetes.cni.cncf.io/v1"
kind: Network
metadata:
  name: protected-private-net-cidr
spec:
  config: '{
    "name": "protected",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "192.168.20.0/24"
    }
}'
NET

    cat << NET > onap-private-net-cidr-network.yaml
apiVersion: "kubernetes.cni.cncf.io/v1"
kind: Network
metadata:
  name: onap-private-net-cidr
spec:
  config: '{
    "name": "onap",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "10.10.0.0/16"
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
        kubernetes.v1.cni.cncf.io/networks: '[
            { "name": "unprotected-private-net-cidr", "interfaceRequest": "eth1" },
            { "name": "onap-private-net-cidr", "interfaceRequest": "eth2" }
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
        kubernetes.v1.cni.cncf.io/networks: '[
            { "name": "unprotected-private-net-cidr", "interfaceRequest": "eth1" },
            { "name": "protected-private-net-cidr", "interfaceRequest": "eth2" },
            { "name": "onap-private-net-cidr", "interfaceRequest": "eth3" }
        ]'
    spec:
      containers:
      - name: $firewall_deployment_name
        image: electrocucaracha/firewall
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        resources:
          limits:
            memory: 160Mi
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
        kubernetes.v1.cni.cncf.io/networks: '[
            { "name": "protected-private-net-cidr", "interfaceRequest": "eth1" },
            { "name": "onap-private-net-cidr", "interfaceRequest": "eth2" }
        ]'
    spec:
      containers:
      - name: $sink_deployment_name
        image: electrocucaracha/sink
        imagePullPolicy: IfNotPresent
        tty: true
        stdin: true
        resources:
          limits:
            memory: 160Mi
DEPLOYMENT

    popd
}

# popule_CSAR_vms_vFW() - This function creates the content of CSAR file
# required for vFirewal using only virtual machines
function popule_CSAR_vms_vFW {
    local csar_id=$1
    ssh_key=$(cat $HOME/.ssh/id_rsa.pub)

    _checks_args $csar_id
    pushd ${CSAR_DIR}/${csar_id}

    cat << META > metadata.yaml
resources:
  network:
    - unprotected-private-net-cidr-network.yaml
    - protected-private-net-cidr-network.yaml
    - onap-private-net-cidr-network.yaml
  deployment:
    - $packetgen_deployment_name.yaml
    - $firewall_deployment_name.yaml
    - $sink_deployment_name.yaml
META

    cat << NET > unprotected-private-net-cidr-network.yaml
apiVersion: "kubernetes.cni.cncf.io/v1"
kind: Network
metadata:
  name: unprotected-private-net-cidr
spec:
  config: '{
    "name": "unprotected",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "192.168.10.0/24"
    }
}'
NET

    cat << NET > protected-private-net-cidr-network.yaml
apiVersion: "kubernetes.cni.cncf.io/v1"
kind: Network
metadata:
  name: protected-private-net-cidr
spec:
  config: '{
    "name": "protected",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "192.168.20.0/24"
    }
}'
NET

    cat << NET > onap-private-net-cidr-network.yaml
apiVersion: "kubernetes.cni.cncf.io/v1"
kind: Network
metadata:
  name: onap-private-net-cidr
spec:
  config: '{
    "name": "onap",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "10.10.0.0/16"
    }
}'
NET

    proxy="#!/bin/bash"
    if [[ -n "${http_proxy+x}" ]]; then
        proxy+="
                export http_proxy=$http_proxy
                echo \"Acquire::http::Proxy \\\"$http_proxy\\\";\" | sudo tee --append /etc/apt/apt.conf.d/01proxy"
    fi
    if [[ -n "${https_proxy+x}" ]]; then
        proxy+="
                export https_proxy=$https_proxy
                echo \"Acquire::https::Proxy \\\"$https_proxy\\\";\" | sudo tee --append /etc/apt/apt.conf.d/01proxy"
    fi
    if [[ -n "${no_proxy+x}" ]]; then
        proxy+="
                export no_proxy=$no_proxy"
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
        VirtletCloudInitUserData: |
          users:
          - default
          - name: admin
            sudo: ALL=(ALL) NOPASSWD:ALL
            plain_text_passwd: secret
            groups: sudo
            ssh_authorized_keys:
            - $ssh_key
        VirtletCloudInitUserDataScript: |
            $proxy

            wget -O - https://raw.githubusercontent.com/electrocucaracha/vFW-demo/master/$packetgen_deployment_name | sudo -E bash
        kubernetes.v1.cni.cncf.io/networks: '[
            { "name": "unprotected-private-net-cidr", "interfaceRequest": "eth1" },
            { "name": "onap-private-net-cidr", "interfaceRequest": "eth2" }
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
        VirtletCloudInitUserData: |
          users:
          - default
          - name: admin
            sudo: ALL=(ALL) NOPASSWD:ALL
            plain_text_passwd: secret
            groups: sudo
            ssh_authorized_keys:
            - $ssh_key
        VirtletCloudInitUserDataScript: |
            $proxy

            wget -O - https://raw.githubusercontent.com/electrocucaracha/vFW-demo/master/$firewall_deployment_name | sudo -E bash
        kubernetes.v1.cni.cncf.io/networks: '[
            { "name": "unprotected-private-net-cidr", "interfaceRequest": "eth1" },
            { "name": "protected-private-net-cidr", "interfaceRequest": "eth2" },
            { "name": "onap-private-net-cidr", "interfaceRequest": "eth3" }
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
            memory: 160Mi
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
        VirtletCloudInitUserData: |
          users:
          - default
          - name: admin
            sudo: ALL=(ALL) NOPASSWD:ALL
            plain_text_passwd: secret
            groups: sudo
            ssh_authorized_keys:
            - $ssh_key
        VirtletCloudInitUserDataScript: |
            $proxy

            wget -O - https://raw.githubusercontent.com/electrocucaracha/vFW-demo/master/$sink_deployment_name | sudo -E bash
        kubernetes.v1.cni.cncf.io/networks: '[
            { "name": "protected-private-net-cidr", "interfaceRequest": "eth1" },
            { "name": "onap-private-net-cidr", "interfaceRequest": "eth2" }
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
        resources:
          limits:
            memory: 160Mi
DEPLOYMENT
    popd
}

# popule_CSAR_multus() - This function creates the content of CSAR file
# required for testing Multus feature
function popule_CSAR_multus {
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
apiVersion: "kubernetes.cni.cncf.io/v1"
kind: Network
metadata:
  name: bridge-conf
spec:
  config: '{
    "name": "mynet",
    "type": "bridge",
    "ipam": {
        "type": "host-local",
        "subnet": "10.10.0.0/16"
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
        kubernetes.v1.cni.cncf.io/networks: '[
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

# popule_CSAR_virtlet() - This function creates the content of CSAR file
# required for testing Virtlet feature
function popule_CSAR_virtlet {
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
        # This tells CRI Proxy that this pod belongs to Virtlet runtime
        kubernetes.io/target-runtime: virtlet.cloud
        VirtletCloudInitUserDataScript: |
          #!/bin/sh
          echo hello world
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

