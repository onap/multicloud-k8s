#!/bin/bash


if [ "$#" -ne 1 ] ; then
	kubectl -n emco exec `kubectl get pods -lapp=etcd -n emco --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -it -- etcdctl get /context/ --prefix=true --keys-only=true
else
	kubectl -n emco exec `kubectl get pods -lapp=etcd -n emco --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}'` -it -- etcdctl get $1 --prefix=true
fi
