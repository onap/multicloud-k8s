#!/bin/bash

# Precondition:
# Optane PM related utilities download and configure.

# collet and install ndctl and check hardware
echo "[OPTANE] Install ndctl ..."
apt install -y ndctl

echo "[OPTANE] Check the NVDIMM hardware ..."
ndctl_region=`ndctl list -R`
if [[ $ndctl_region == "" ]] ; then
    echo "No NVDIMM hardware, exit ..."
    exit 0
fi

# get current folder path
work_path="$(dirname -- "$(readlink -f -- "$0")")"
node_name="$(kubectl get node -o jsonpath='{.items[0].metadata.name}')"

# collet and install ipmctl
echo "[OPTANE] Install ipmctl ..."
cd $work_path
wget https://launchpad.net/ubuntu/+archive/primary/+sourcefiles/ipmctl/02.00.00.3474+really01.00.00.3469-1/ipmctl_02.00.00.3474+really01.00.00.3469.orig.tar.xz
tar xvf ipmctl_02.00.00.3474+really01.00.00.3469.orig.tar.xz
cd ipmctl-01.00.00.3469/

echo "[OPTANE] Install ipmctl utilities"
mkdir output && cd output
apt install -y cmake build-essential pkg-config asciidoctor asciidoc libndctl-dev git
gem install asciidoctor-pdf --pre

add-apt-repository --yes ppa:jhli/libsafec
apt update
apt-get install -y libsafec-dev

echo "[OPTANE] Build ipmctl ..."
cmake -DRELEASE=ON -DCMAKE_INSTALL_PREFIX=/ ..
make -j all
make install

cd $work_path

echo "[OPTANE] Install cfssl tools ..."
# collect cfssl tools
curl -L https://pkg.cfssl.org/R1.2/cfssl_linux-amd64 -o cfssl
curl -L https://pkg.cfssl.org/R1.2/cfssljson_linux-amd64 -o cfssljson
chmod a+x cfssl cfssljson
cp -rf cfssl cfssljson /usr/bin/

echo "[OPTANE] Create AppDirect Goal ..."
# ipmctl setting
#ipmctl delete -goal
#ipmctl create -f -goal PersistentMemoryType=AppDirectNotInterleaved

# Run certificates set-up script
echo "[OPTANE] Run ca for kubernetes ..."
./setup-ca-kubernetes.sh

# deploy docker hub
echo "[OPTANE] Set label node for storage pmem ..."
kubectl label node $node_name storage=pmem

echo "[OPTANE] kubelet CSIMigration set false ..."
echo -e "featureGates:\n  CSIMigration: false" >> /var/lib/kubelet/config.yaml
# deploy pmem-csi and applications
# select two mode: lvm and direct
#echo "[OPTANE] Create PMEM-CSI plugin service ..."
#kubectl create -f ../images/pmem-csi-lvm.yaml
# kubectl create -f pmem-csi-direct.yaml

