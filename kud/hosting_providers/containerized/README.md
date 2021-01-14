# Multi cluster installation

## Introduction

Multi Cluster installation is an important features for production deployments.

Most of the project are using the Kubernetes as undercloud orchestration. So deploying multi cluster for the multi cloud region should be maintained by Kubernetes

This section explains how to deploy the Multi cluster of Kubernetes from a containerized KUD running as a Kubernetes Job.

## How it works

KUD installation installer is divided into two regions with args - `--install-pkg` and `--cluster <cluster-name>`

### Args
**--install-pkg** - Installs packages required to run installer script itself inside a container and kubespray packages

**--cluster < cluster-name >** - Installs k8s cluster, addons and plugins and store the artifacts in the host machine

### Internal Mechanism

* Container image is build using the `installer --install-pkg` arg and Kubernetes job is used to install the cluster using `installer --cluster <cluster-name>`. Installer will invoke the kubespray cluster.yml, kud-addsons and plugins ansible cluster.

Installer script finds the `hosts.ini` for each cluster in `/opt/multi-cluster/<cluster-name>`

Kubernetes jobs (a cluster per job) are used to install multiple clusters and logs of each cluster deployments are stored in the `/opt/kud/multi-cluster/<cluster-name>/logs` and artifacts are stored as follows `/opt/kud/multi-cluster/<cluster-name>/artifacts`

## Creating TestBed for Testing and Development

This section creates VM for testing and development work in containerization setup. User can skip this steps, if they have baremetal avaiable for testing or development.

```
$ git clone https://github.com/onap/multicloud-k8s.git
$ pushd multicloud-k8s/kud/hosting_providers/vagrant
$ sudo ./setup.sh -p libvirt
$ popd
$ pushd multicloud-k8s/kud/hosting_providers/containerized/testing
$ vagrant up
$ popd
```
Do the following steps to keep note of
1. Get the IP address for the Vagrant machine - <VAGRANT_IP_ADDRESS>
2. Copy the host /root/.ssh/id_rsa.pub into the vagrant /root/.ssh/authorized_keys
3. From host make sure to ssh into vagrant without password ssh root@<VAGRANT_IP_ADDRESS>

## Quickstart Installation Guide

Build the kud docker images as follows. Add `KUD_ENABLE_TESTS` & `KUD_PLUGIN_ENABLED`
for the testing only. Currently only docker and containerd are supported CRI
runtimes and can be configured using the `CONTAINER_RUNTIME` environment variable.
To be able to run secure containers using Kata Containers, it is required to
change the CRI runtime to containerd.

```
$ git clone https://github.com/onap/multicloud-k8s.git && cd multicloud-k8s
$ docker build  --rm \
	--build-arg http_proxy=${http_proxy} \
	--build-arg HTTP_PROXY=${HTTP_PROXY} \
	--build-arg https_proxy=${https_proxy} \
	--build-arg HTTPS_PROXY=${HTTPS_PROXY} \
	--build-arg no_proxy=${no_proxy} \
	--build-arg NO_PROXY=${NO_PROXY} \
	--build-arg KUD_ENABLE_TESTS=true \
	--build-arg KUD_PLUGIN_ENABLED=true \
	--build-arg CONTAINER_RUNTIME=docker \
	-t github.com/onap/multicloud-k8s:latest . -f kud/build/Dockerfile
```
Let's create a cluster-101 and cluster-102 hosts.ini as follows

```
$ mkdir -p /opt/kud/multi-cluster/{cluster-101,cluster-102}
```

Create the hosts.ini as follows in the directory cluster-101(c01 IP address 10.10.10.3) and cluster-102(c02 IP address 10.10.10.5). If the user used a Vagrant setup as mentioned in the above steps, replace the IP address with the vagrant IP address.

```
$ cat /opt/kud/multi-cluster/cluster-101/hosts.ini
[all]
c01 ansible_ssh_host=<VAGRANT_IP_ADDRESS> ansible_ssh_port=22

[kube-master]
c01

[kube-node]
c01

[etcd]
c01

[ovn-central]
c01

[ovn-controller]
c01

[virtlet]
c01

[k8s-cluster:children]
kube-node
kube-master
```
Do the same for the cluster-102 with c01 and IP address 10.10.10.5.

Create the ssh secret for Baremetal or VM based on your deployment. Launch the kubernetes job as follows.
```
$ kubectl create secret generic ssh-key-secret --from-file=id_rsa=/root/.ssh/id_rsa --from-file=id_rsa.pub=/root/.ssh/id_rsa.pub
$ CLUSTER_NAME=cluster-101
$ cat <<EOF | kubectl create -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: kud-$CLUSTER_NAME
spec:
  template:
    spec:
      hostNetwork: true
      containers:
        - name: kud
          image: github.com/onap/multicloud-k8s:latest
          imagePullPolicy: IfNotPresent
          volumeMounts:
          - name: multi-cluster
            mountPath: /opt/kud/multi-cluster
          - name: secret-volume
            mountPath: "/.ssh"
          command: ["/bin/sh","-c"]
          args: ["cp -r /.ssh /root/; chmod -R 600 /root/.ssh; ./installer --cluster $CLUSTER_NAME --plugins onap4k8s"]
          securityContext:
            privileged: true
      volumes:
      - name: multi-cluster
        hostPath:
          path: /opt/kud/multi-cluster
      - name: secret-volume
        secret:
          secretName: ssh-key-secret
      restartPolicy: Never
  backoffLimit: 0

EOF
```
See the logs as follows to see the progress of KUD installation
```
$ kubectl logs -f kud-$CLUSTER_NAME-<AAAA>
```

Multi - cluster information from the host machine;

```
$ kubectl --kubeconfig=/opt/kud/multi-cluster/cluster-101/artifacts/admin.conf cluster-info
Kubernetes master is running at https://192.168.121.2:6443
coredns is running at https://192.168.121.2:6443/api/v1/namespaces/kube-system/services/coredns:dns/proxy
kubernetes-dashboard is running at https://192.168.121.2:6443/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
$ kubectl --kubeconfig=/opt/kud/multi-cluster/cluster-102/artifacts/admin.conf cluster-info
Kubernetes master is running at https://192.168.121.6:6443
coredns is running at https://192.168.121.6:6443/api/v1/namespaces/kube-system/services/coredns:dns/proxy
kubernetes-dashboard is running at https://192.168.121.6:6443/api/v1/namespaces/kube-system/services/https:kubernetes-dashboard:/proxy

To further debug and diagnose cluster problems, use 'kubectl cluster-info dump'.
```


## License

Apache-2.0
