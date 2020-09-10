# Steps for running EMCO API microservices

### Steps to install packages
**1. Create namespace for EMCO Microservices**

`$ kubectl create namespace onap4k8s`

**2. Create Databases used by EMCO Microservices for Etcd and Mongo**

`$ kubectl apply -f onap4k8sdb.yaml -n onap4k8s`

**3. create EMCO Microservices**

`$ kubectl apply -f onap4k8s.yaml -n onap4k8s`

### Steps to cleanup  packages
**1. Cleanup EMCO Microservies**

`$ kubectl delete -f onap4k8s.yaml -n onap4k8s`

**2. Cleanup EMCO Microservices for Etcd and Mongo**

`$ kubectl delete -f onap4k8sdb.yaml -n onap4k8s`

# Steps for running the monitor microservice in clusters

The EMCO microservices utilize the monitor microservice to collect
status information from clusters to which EMCO deploys applications.
It must be installed in each cluster to which EMCO deploys applications.

### Steps to install monitor in a cluster

**1. Instantiate the monitor resources

 $ kubectl apply -f monitor-deploy.yaml

### Steps to cleanup monitor in a cluster

**1. Cleanup the monitor resources

 $ kubectl delete -f monitor-deploy.yaml
