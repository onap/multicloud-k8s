#!/bin/bash

# Simple script to view appcontext
# with no argumnet, it will list all keys under /context/
# with 1 argument, it will show the value of the key provided
# note: assumes emoco services are running in namespace emco
if [ "$#" -ne 1 ] ; then
    kubectl -n emco exec `kubectl get pods -lapp=etcd -n emco --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -it -- etcdctl get /context/ --prefix=true --keys-only=true
else
if [ "$1" == "del" ] ; then
    kubectl -n emco exec `kubectl get pods -lapp=etcd -n emco --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -it -- etcdctl del /context/ --prefix=true
else
    kubectl -n emco exec `kubectl get pods -lapp=etcd -n emco --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -it -- etcdctl get $1 --prefix=true
fi
fi
