#!/bin/bash
# PDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

echo "[OPTANE-TEST] Check the NVDIMM hardware ..."
ndctl_region=`ndctl list -R`
if [[ $ndctl_region == "" ]] ; then
    echo "No NVDIMM hardware, exit ..."
    exit 0
fi

pod_sc_01=pod-sc-case-01
pod_pvc_01=pod-pvc-case-01
pod_app_01=pod-app-case-01

cat << POD > $HOME/$pod_sc_01.yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: pmem-csi-sc-ext4
parameters:
  csi.storage.k8s.io/fstype: ext4
  eraseafter: "true"
provisioner: pmem-csi.intel.com
reclaimPolicy: Delete
volumeBindingMode: Immediate
POD

cat << POD > $HOME/$pod_pvc_01.yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pmem-csi-pvc-ext4
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi
  storageClassName: pmem-csi-sc-ext4
POD

cat << POD > $HOME/$pod_app_01.yaml
kind: Pod
apiVersion: v1
metadata:
  name: my-csi-app-1
spec:
  containers:
    - name: my-frontend
      image: busybox
      command: [ "sleep", "100000" ]
      volumeMounts:
      - mountPath: "/data"
        name: my-csi-volume
  volumes:
  - name: my-csi-volume
    persistentVolumeClaim:
      claimName: pmem-csi-pvc-ext4
POD

kubectl apply -f $HOME/$pod_sc_01.yaml
kubectl apply -f $HOME/$pod_pvc_01.yaml
kubectl apply -f $HOME/$pod_app_01.yaml

echo "Sleep for several minutes ..."
sleep 300

pvc_meta="$(kubectl get pvc -o jsonpath='{.items[0].metadata.name}')"
pvc_status="$(kubectl get pvc -o jsonpath='{.items[0].status.phase}')"
if [[ $pvc_meta == "pmem-csi-pvc-ext4" ]] && [[ $pvc_status == "Bound" ]] ; then
    echo "[OPTANE] SUCCESS: created PMEM-CSI volume!"
else
    echo "[OPTANE] FAILED: cannot create PMEM-CSI volume!"
fi

echo "Wait and remove the test resource ..."
sleep 60

kubectl delete -f $HOME/$pod_sc_01.yaml
kubectl delete -f $HOME/$pod_pvc_01.yaml
kubectl delete -f $HOME/$pod_app_01.yaml

