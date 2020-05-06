#################################################################
# ONAP4K8S v2 helm charts
#################################################################

ONAP4K8s Helm charts include charts for ONAP4K8s microservices
along with MongoDb, etcd, Promethus, cAdvisor, Flutend and Jaeger

1. Create a local helm repo from Makefile
$ make repo

2. Run make file to package all the required chart. Output is in tar.gz format.
$ make clean
$ make all

The output from this is in dist/packages directory and the package of intrest are:
    - mco-db-5.0.0.tgz -> Contains database packages for mongo & etcd
    - mco-services-5.0.0.tgz -> Contains packages for all ONAP4K8s services like orchestrator, ncm, rsync etcd
    - mco-tools-5.0.0.tgz -> Tools like Prometheus, Collectd, Fluentd to be used with ONAP4K8s
    - mco-5.0.0.tgz  -> Contains all charts including database mongo & etcd, all services and tools

3. Deploy the generated Chart mco-5.0.0.tgz can be deployed to deploy all packages or individual packages can be deployed like:
    $ helm install dist/packages/mco-db-5.0.0.tgz --name onap4k8s-db --namespace onap
    $ helm install dist/packages/mco-services-5.0.0.tgz --name onap4k8s --namespace onap

4. Deploy tools (Optional)
    $ helm install dist/packages/mco-tools-5.0.0.tgz --name onap4k8s-tools --namespace onap
        To check logs of the different Microservices check fluentd logs
        kubectl logs mco-fluentd-0 -n test | grep orchestrator
        Prometheus UI can be used to get statistics for the microservices.

5. Delete all packages
    $helm delete onap4k8s --purge
    $helm delete onap4k8s-db --purge
    $helm delete onap4k8s-tools --purge

