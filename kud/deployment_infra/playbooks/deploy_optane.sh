#!/bin/bash

work_path="$(dirname -- "$(readlink -f -- "$0")")"
ndctl_region=`ndctl list -R`
if [[ $ndctl_region == "" ]] ; then
    echo "No Optane Hardware!"
else
    echo "Optane Plugin start .."
    /usr/local/bin/kubectl apply -f $work_path/pmem-csi-lvm.yaml
fi
