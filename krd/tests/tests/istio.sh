#!/bin/bash
# SPDX-license-identifier: Apache-2.0
##############################################################################
# Copyright (c) 2018
# All rights reserved. This program and the accompanying materials
# are made available under the terms of the Apache License, Version 2.0
# which accompanies this distribution, and is available at
# http://www.apache.org/licenses/LICENSE-2.0
##############################################################################

set -o errexit
set -o nounset
set -o pipefail

source _functions.sh

csar_id=ac39959e-e82c-11e8-9133-525400912638

base_dest=$(grep "base_dest:" $test_folder/../playbooks/krd-vars.yml | awk -F ': ' '{print $2}')
istio_dest=$(grep "istio_dest:" $test_folder/../playbooks/krd-vars.yml | awk -F ': ' '{print $2}' | sed "s|{{ base_dest }}|$base_dest|g;s|\"||g")
istio_version=$(grep "istio_version:" $test_folder/../playbooks/krd-vars.yml | awk -F ': ' '{print $2}')

if ! $(istioctl version &>/dev/null); then
    echo "This funtional test requires istioctl client"
    exit 1
fi

_checks_args $csar_id
pushd ${CSAR_DIR}/${csar_id}
istioctl kube-inject -f $istio_dest/istio-$istio_version/samples/bookinfo/platform/kube/bookinfo.yaml > bookinfo-inject.yml
kubectl apply -f bookinfo-inject.yml
kubectl apply -f $istio_dest/istio-$istio_version/samples/bookinfo/networking/bookinfo-gateway.yaml

for deployment in details-v1 productpage-v1 ratings-v1 reviews-v1 reviews-v2 reviews-v3; do
    wait_deployment $deployment
done
INGRESS_PORT=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.spec.ports[?(@.name=="http2")].nodePort}')
INGRESS_HOST=$(kubectl get po -l istio=ingressgateway -n istio-system -o 'jsonpath={.items[0].status.hostIP}')
curl -o /dev/null -s -w "%{http_code}\n" http://$INGRESS_HOST:$INGRESS_PORT/productpage
popd
