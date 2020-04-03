#!/bin/bash

# Precondition:
# Optane PM related utilities download and configure.

# get current folder path
work_path=$(dirname $(readlink -f $0))

# collet and install ipmctl and ndctl
apt install ndctl

cd $(work_path)
wget https://launchpad.net/ubuntu/+archive/primary/+sourcefiles/ipmctl/02.00.00.3474+really01.00.00.3469-1/ipmctl_02.00.00.3474+really01.00.00.3469.orig.tar.xz
tar xvf ipmctl_02.00.00.3474+really01.00.00.3469.orig.tar.xz
cd ipmctl-01.00.00.3469/

mkdir output && cd output
apt install cmake build-essential pkg-config asciidoctor asciidoc libndctl-dev git
gem install asciidoctor-pdf --pre

add-apt-repository ppa:jhli/libsafec
apt update
apt-get install libsafec-dev

cmake -DRELEASE=ON -DCMAKE_INSTALL_PREFIX=/ ..
make -j all
make install

cd $(work_path)

# collect cfssl tools
curl -L https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 -o cfssl
curl -L https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 -o cfssljson
chmod a+x cfssl cfssljson
cp -rf cfssl cfssljson /usr/bin/

# ipmctl setting
ipmctl create -goal PersistentMemoryType=AppDirectNotInterleaved

# Run certificates set-up script
./setup-ca-kubernetes.sh

# deploy docker hub

kubectl label node <your node> storage=pmem

# deploy pmem-csi and applications
# select two mode: lvm and direct
kubectl create -f pmem-csi-lvm.yaml
# kubectl create -f pmem-csi-direct.yaml

