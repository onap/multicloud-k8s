#!/bin/bash

work_path="$(dirname -- "$(readlink -f -- "$0")")"

if [[ `ipmctl show -dimm` =~ "No DIMMs" ]]; then
    echo "No Optane Hardware!"
else
    echo "Optane Plugin start .."
    /usr/local/bin/kubectl apply -f $work_path/pmem-csi-lvm.yaml
fi
