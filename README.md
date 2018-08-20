# MultiCloud-k8-plugin

MultiCloud Kubernetes plugin for ONAP multicloud.

# Installation

Requirements:
* Go 1.10
* Dep

Steps:

* Clone repo in GOPATH src:
    * `cd $GOPATH/src && git clone https://git.onap.org/multicloud/k8s`

* Run unit tests:
    *  `make build`

* Compile to build Binary:
    * `make deploy`

# Archietecture

Create Virtual Network Function

![Create VNF](https://raw.githubusercontent.com/shank7485/k8-plugin-multicloud/master/docs/create_vnf.png)

Create Virtual Link

![Create VL](https://raw.githubusercontent.com/shank7485/k8-plugin-multicloud/master/docs/create_vl.png)
