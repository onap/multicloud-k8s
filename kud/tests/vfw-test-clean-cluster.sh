#!/bin/bash

# script to delete vfw resources (until terminate is completed)
kubectl delete deploy fw0-packetgen
kubectl delete deploy fw0-firewall
kubectl delete deploy fw0-sink
kubectl delete service packetgen-service
kubectl delete service sink-service
kubectl delete configmap sink-configmap

kubectl delete network protected-private-net
kubectl delete providernetwork emco-private-net
kubectl delete providernetwork unprotected-private-net

for i in `kubectl get resourcebundlestate --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'`; do
    kubectl delete resourcebundlestate $i
done
