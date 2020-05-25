#!/bin/bash

if [[ `ipmctl show -dimm` =~ "No DIMMs" ]]; then
    echo "No Optane Hardware!"
else
    echo "Optane Plugin start .."
    /usr/local/bin/kubectl apply -f optane/pmem-csi-lvm.yaml
fi
