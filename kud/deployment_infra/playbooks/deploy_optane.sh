#!/bin/bash

if [[ `ipmctl show -dimm` =~ "No DIMMs" ]]; then
    echo "No Optane Hardware!"
else
    /usr/local/bin/kubectl apply -f $1/../images/pmem-csi-lvm.yaml
fi
