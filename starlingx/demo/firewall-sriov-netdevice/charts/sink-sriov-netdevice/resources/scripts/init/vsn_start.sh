#!/bin/bash

apt-get update
apt-get install -y sudo curl net-tools iproute2 inetutils-ping wget darkstat unzip

echo "provision interfaces"

ifconfig veth22 10.10.2.2/24

echo "add route entries"
ip route add 10.10.1.0/24 via 10.10.2.1

echo "update darkstat configuration"
sed -i "s/START_DARKSTAT=.*/START_DARKSTAT=yes/g;s/INTERFACE=.*/INTERFACE=\"-i veth22\"/g" /etc/darkstat/init.cfg

echo "start darkstat"

darkstat -i veth22

echo "done"
sleep infinity