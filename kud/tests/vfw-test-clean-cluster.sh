kubectl delete deploy fw0-packetgen
kubectl delete deploy fw0-firewall
kubectl delete deploy fw0-sink
kubectl delete service packetgen-service
kubectl delete service sink-service
kubectl delete configmap sink-configmap
