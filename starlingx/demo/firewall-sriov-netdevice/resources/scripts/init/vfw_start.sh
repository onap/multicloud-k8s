#!/bin/bash

apt-get update
apt-get install -y sudo curl net-tools iproute2
curl -s https://packagecloud.io/install/repositories/fdio/release/script.deb.sh | sudo bash
export VPP_VER=19.01.2-release
apt-get install -y vpp=$VPP_VER vpp-lib=$VPP_VER
apt-get install -y vpp-plugins=$VPP_VER

if [ -e /run/vpp/cli-vpp2.sock ]; then
    rm /run/vpp/cli-vpp2.sock
fi

# extract core list
#	root@vpktgen:/# taskset -p --cpu-list 1
#	pid 1's current affinity list: 1,2,29

corelist=`taskset -p -c 1 |cut -d : -f 2 | sed 's/^ *//' | sed 's/ *$//'`
#extract master core
mastercoreidx=`echo $corelist | cut -d , -f 1`
#extract worker cores
workercorelist=`echo $corelist | sed -E 's/^[0-9]*,//'`

echo 'start... vpp'
vpp unix {cli-listen /run/vpp/cli-vpp2.sock} api-segment { prefix vpp2 } \
    cpu { main-core $mastercoreidx  corelist-workers $workercorelist }

echo 'wait vpp be up ...'
while [ ! -e /run/vpp/cli-vpp2.sock ]; do
    sleep 1;
done

echo 'configure vpp ...'

vppctl -s /run/vpp/cli-vpp2.sock show ver
vppctl -s /run/vpp/cli-vpp2.sock show threads

vppctl -s /run/vpp/cli-vpp2.sock create host-interface name veth12

vppctl -s /run/vpp/cli-vpp2.sock set int state host-veth12 up

vppctl -s /run/vpp/cli-vpp2.sock set int ip address host-veth12 10.10.1.1/24

vppctl -s /run/vpp/cli-vpp2.sock create host-interface name veth21

vppctl -s /run/vpp/cli-vpp2.sock set int state host-veth21 up

vppctl -s /run/vpp/cli-vpp2.sock set int ip address host-veth21 10.10.2.1/24

vppctl -s /run/vpp/cli-vpp2.sock show hardware
vppctl -s /run/vpp/cli-vpp2.sock show int
vppctl -s /run/vpp/cli-vpp2.sock show int addr

vppctl -s /run/vpp/cli-vpp2.sock show ip fib

#vppctl -s /run/vpp/cli-vpp2.sock trace add af-packet-input 10

#vppctl -s /run/vpp/cli-vpp2.sock ping 10.10.1.2

#vppctl -s /run/vpp/cli-vpp2.sock show trace
#vppctl -s /run/vpp/cli-vpp2.sock show ip arp

echo "done"
sleep infinity
