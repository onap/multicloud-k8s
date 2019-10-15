#!/bin/bash

for file in $(find /etc/*.conf -type f -name "c6xxvf_dev*.conf"); do
       device_id=$( echo $file | cut -d '_' -f 2 | tr -cd '[[:digit:]]')
       echo $device_id
       cat /etc/c6xxvf_dev${device_id}.conf
       sudo sed -i "s/\[SSL\]/\[SSL${device_id}\]/g" /etc/c6xxvf_dev${device_id}.conf
done
