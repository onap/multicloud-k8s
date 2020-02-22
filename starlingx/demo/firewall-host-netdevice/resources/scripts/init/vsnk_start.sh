#!/bin/bash

apt-get update
apt-get install -y sudo curl net-tools iproute2 inetutils-ping wget darkstat unzip

ifconfig veth22 10.10.2.2/24

ip route add 10.10.1.0/24 via 10.10.2.1

sed -i "s/START_DARKSTAT=.*/START_DARKSTAT=yes/g;s/INTERFACE=.*/INTERFACE=\"-i veth22\"/g" /etc/darkstat/init.cfg

darkstat -i veth22
