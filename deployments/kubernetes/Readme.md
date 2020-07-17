# Steps for running v2 API microservices

### Steps to install packages
**1. Create namespace for ONAP4K8s v2 Microservices**

`$ kubectl create namespace onap4k8s`

**2. Create Databases used by ONAP4K8s v2 Microservices for Etcd and Mongo**

`$ kubectl apply -f onap4k8sdb.yaml -n onap4k8s`

**3. create ONAP4K8s v2 Microservices**

`$ kubectl apply -f onap4k8s.yaml -n onap4k8s`

### Steps to cleanup  packages
**1. Run "cleanup-emco.sh" to clenaup v2 Microservies**

`$ ./cleanup-emco.sh`

**2. Cleanup v2 Microservices for Etcd and Mongo**

`$ kubectl delete -f onap4k8sdb.yaml -n onap4k8s`
