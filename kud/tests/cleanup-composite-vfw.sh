# To clean up composite vfw demo resources in a cluster
kubectl -n onap4k8s delete deployment clm
kubectl -n onap4k8s delete deployment orchestrator
kubectl -n onap4k8s delete deployment ncm
kubectl -n onap4k8s delete deployment ovnaction
kubectl -n onap4k8s delete deployment rsync
kubectl -n onap4k8s delete service clm
kubectl -n onap4k8s delete service orchestrator
kubectl -n onap4k8s delete service ncm
kubectl -n onap4k8s delete service ovnaction
kubectl -n onap4k8s delete service rsync
kubectl -n onap4k8s delete configmap clm
kubectl -n onap4k8s delete configmap orchestrator
kubectl -n onap4k8s delete configmap ncm
kubectl -n onap4k8s delete configmap ovnaction
kubectl -n onap4k8s delete configmap rsync

# delete the networks
kubectl delete network protected-private-net
kubectl delete providernetwork emco-private-net
kubectl delete providernetwork unprotected-private-net
