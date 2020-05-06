ONAP4K8s Helm charts include charts for ONAP4K8s microservices
along with Promethus, cAdvisor, Flutend and Jaeger

#################################################################
# Installation of v2 ONAP4K8S helm chart
#################################################################

1. Create a local helm repo from Makefile
$ make repo

2. Run make file to package all the required chart in a single tar.gz
$ make clean
$ make all

3. Deploy the generated Chart
$ helm install dist/packages/mco-5.0.0.tgz --name mco --namespace test


To check logs of the different Microservices check fluentd logs
kubectl logs mco-fluentd-0 -n test | grep orchestrator 
Prometheus UI can be used to get statistics for the microservices.
