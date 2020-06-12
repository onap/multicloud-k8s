# Steps for running v2 API microservices
$kubectl create namespace onap4k8s
$kubectl apply -f onap4k8sdb.yaml -n onap4k8s
$kubectl apply -f onap4k8s.yaml -n onap4k8s
