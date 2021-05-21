#!/bin/bash

source _common.sh

ENV=$(kubectl get nodes --all-namespaces | wc -l)
if [[ $ENV -gt 2 ]]; then
    COMPUTE_NODE=$(kubectl get nodes --all-namespaces | grep -v master | awk 'NR==2{print $1}')
else
    COMPUTE_NODE=$(kubectl get nodes --all-namespaces | grep master | awk 'NR==1{print $1}')
fi
cases=("exclusive ${COMPUTE_NODE} 1" "shared ${COMPUTE_NODE} -1")
case=(null null 0)
num=${#cases[*]}
POOL=0
NODE=1
CORE=2
DIR=/tmp
pod_name=cmk-test-pod

function wait_for_pod_up {
    status_phase=""
    while [[ $status_phase != "Running" ]]; do
        new_phase=$(kubectl get pods "$@" | awk 'NR==2{print $3}')
        if [[ $new_phase != $status_phase ]]; then
            echo "$(date +%H:%M:%S) - $@ : $new_phase"
            status_phase=$new_phase
        fi
        if [[ $new_phase == "Running" ]]; then
            echo "Pod $@ is up and running.."
        fi
        if [[ $new_phase == "Err"* ]]; then
            exit 1
        fi
    done
}

function start_nginx_pod {
    kubectl delete deployment -n default nginx --ignore-not-found=true
    kubectl create deployment nginx --image=nginx
    kubectl wait deployment nginx --for=condition=available
    nginx_pod=$(kubectl get pods --all-namespaces| grep nginx | awk 'NR==1{print $2}')
    kubectl delete deployment -n default nginx --ignore-not-found=true
    pod_status="Running"
    until [[ $pod_status == "" ]]; do
        pod_status=$(kubectl get pod $nginx_pod --ignore-not-found=true | awk 'NR==2{print $3}')
    done
}

rm -f $DIR/$pod_name.yaml
kubectl delete pod $pod_name --ignore-not-found=true --now --wait
echo
echo "env is $ENV"
echo
for ((i=0;i<$num;i++)); do
    inner_case=(${cases[$i]})
    num_inner=${#inner_case[*]}
    for ((j=0;j<$num_inner;j++)); do
        case[$j]=${inner_case[$j]}
    done
    echo "##################################"
    if [ "${case[$POOL]}" == "exclusive" ]; then
        echo "TC: to allocate ${case[$CORE]} CPU(s) from pool of ${case[$POOL]} on node of ${case[$NODE]}"
        TOTAL=$(kubectl get cmk-nodereport ${case[$NODE]} -o json | jq --arg pool "${case[$POOL]}" '.spec.report.description.pools[$pool].cpuLists|length')
        echo "ready to generate yaml"
cat << EOF > $DIR/$pod_name.yaml
    apiVersion: v1
    kind: Pod
    metadata:
      labels:
        app: cmk-test-pod
      name: cmk-test-pod
    spec:
      nodeName: ${case[$NODE]}
      containers:
      - args:
        - "/opt/bin/cmk isolate --conf-dir=/etc/cmk --pool=exclusive sleep -- 3900"
        command:
        - "sh"
        - "-c"
        env:
        - name: CMK_PROC_FS
          value: "/host/proc"
        - name: CMK_NUM_CORES
          value: "${case[$CORE]}"
        image: ubuntu:18.04
        imagePullPolicy: "IfNotPresent"
        name: cmk-test
        volumeMounts:
        - mountPath: "/host/proc"
          name: host-proc
        - mountPath: "/opt/bin"
          name: cmk-install-dir
        - mountPath: "/etc/cmk"
          name: cmk-conf-dir
      restartPolicy: Never
      volumes:
      - hostPath:
          path: "/opt/bin"
        name: cmk-install-dir
      - hostPath:
          path: "/proc"
        name: host-proc
      - hostPath:
          path: "/etc/cmk"
        name: cmk-conf-dir
EOF

        echo "ready to create pod"
        kubectl create -f $DIR/$pod_name.yaml --validate=false
        sleep 2
        echo "waiting for pod up"
        for pod in $pod_name; do
            wait_for_pod_up $pod
        done
        echo "waiting for CPU allocation finished ..."
        rest=$TOTAL
        timeout=0
        until [[ $TOTAL -gt $rest || $timeout == 180 ]]; do
            sleep 1
            rest=$(kubectl get cmk-nodereport ${case[$NODE]} -o json | jq '.spec.report.description.pools.exclusive.cpuLists[].tasks|select(type=="array" and length==0)' | wc -l) || echo "$TOTAL"
            timeout=$((timeout+1))
        done
        let allocated=`expr $TOTAL - $rest`
        echo "The allocated CPU amount is:" $allocated
        echo "deploy a nginx pod"
        start_nginx_pod
        if [[ $allocated == ${case[$CORE]} ]]; then
            echo "CPU was allocated as expected, TC passed !!"
        else
            echo "failed to allocate CPU ($allocated != ${case[$CORE]}), TC failed !!"
        fi
        rm -f $DIR/$pod_name.yaml
        echo "ready to delete pod"
        kubectl delete pod $pod_name --ignore-not-found=true --now --wait
        echo "Pod was deleted"
        echo "##################################"
        echo
        echo
        [[ $allocated == ${case[$CORE]} ]]
    else
        echo "TC: to allocate CPU(s) from pool of ${case[$POOL]} on node of ${case[$NODE]}"
        echo "ready to generate yaml"
cat << EOF > $DIR/$pod_name.yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: cmk-test-pod
  name: cmk-test-pod
spec:
  nodeName: ${case[$NODE]}
  containers:
  - name: share1
    args:
    - "/opt/bin/cmk isolate --conf-dir=/etc/cmk --pool=shared sleep -- 3900"
    command:
    - "sh"
    - "-c"
    env:
    - name: CMK_PROC_FS
      value: "/host/proc"
    - name: CMK_NUM_CORES
      value: "3"
    image: ubuntu:18.10
    imagePullPolicy: "IfNotPresent"
    volumeMounts:
    - mountPath: "/host/proc"
      name: host-proc
    - mountPath: "/opt/bin"
      name: cmk-install-dir
    - mountPath: "/etc/cmk"
      name: cmk-conf-dir
  - name: share2
    args:
    - "/opt/bin/cmk isolate --conf-dir=/etc/cmk --pool=shared sleep -- 3300"
    command:
    - "sh"
    - "-c"
    env:
    - name: CMK_PROC_FS
      value: "/host/proc"
    - name: CMK_NUM_CORES
      value: "3"
    image: ubuntu:18.10
    imagePullPolicy: "IfNotPresent"
    volumeMounts:
    - mountPath: "/host/proc"
      name: host-proc
    - mountPath: "/opt/bin"
      name: cmk-install-dir
    - mountPath: "/etc/cmk"
      name: cmk-conf-dir
  volumes:
  - hostPath:
      path: "/opt/bin"
    name: cmk-install-dir
  - hostPath:
      path: "/proc"
    name: host-proc
  - hostPath:
      path: "/etc/cmk"
    name: cmk-conf-dir
EOF

        echo "ready to create pod"
        kubectl create -f $DIR/$pod_name.yaml --validate=false
        sleep 2
        echo "waiting for pod up"
        for pod in $pod_name; do
            wait_for_pod_up $pod
        done
        echo "waiting for CPU allocation finished ..."
        rest=0
        timeout=0
        until [[ $rest == 2 || $timeout == 180 ]]; do
            sleep 1
            rest=$(kubectl get cmk-nodereport ${case[$NODE]} -o json | jq '.spec.report.description.pools.shared.cpuLists[].tasks' | jq length) || echo "0"
            timeout=$((timeout+1))
        done
        echo "The CPU allocated in shared pool for 2 tasks"
        echo "deploy a nginx pod"
        start_nginx_pod
        if [[ $rest == 2 ]]; then
            echo "CPU was allocated as expected, TC passed !!"
        else
            echo "failed to allocate CPU ($rest != 2), TC failed !!"
        fi
        rm -f $DIR/$pod_name.yaml
        echo "ready to delete pod"
        kubectl delete pod $pod_name --ignore-not-found=true --now --wait
        echo "Pod was deleted"
        echo "##################################"
        echo
        echo
        [[ $rest == 2 ]]
    fi
done
