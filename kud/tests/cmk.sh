#!/bin/bash
cases=("exclusive localhost 3" "exclusive compute01 3" "shared localhost -1" "shared compute01 -1")
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
    nginx_pod=$(kubectl get pods --all-namespaces| grep nginx | awk 'NR==1{print $2}')
    wait_for_pod_up $nginx_pod
    kubectl delete deployment -n default nginx --ignore-not-found=true
    pod_status="Running"
    until [[ $pod_status == "" ]]; do
        pod_status=$(kubectl get pod $nginx_pod --ignore-not-found=true | awk 'NR==2{print $3}')
    done
}

rm -f $DIR/$pod_name.yaml
kubectl delete pod $pod_name --ignore-not-found=true --now --wait
ENV=$(kubectl get nodes --all-namespaces | awk 'NR==2{print $1}')
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
        if [ "${case[$NODE]}" == "$ENV" ]; then
            TOTAL=$(kubectl get cmk-nodereport ${case[$NODE]} -o json | jq .spec.report.description.pools.${case[$POOL]} | jq .cpuLists | awk -F  '{' '{print $(NF)}' | awk -F  '}' '{print $(NF)}' | awk -F  ',' '{print $(NF)}' | grep "\"tasks\": \[" | wc -l)
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
          value: "/proc"
        - name: CMK_NUM_CORES
          value: "${case[$CORE]}"
        image: ubuntu:18.04
        imagePullPolicy: "IfNotPresent"
        name: cmk-test
        volumeMounts:
        - mountPath: "/proc"
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
            echo "waiting for pod up"
            for pod in $pod_name; do
                wait_for_pod_up $pod
            done
            echo "waiting for CPU allocation finished ..."
            rest=$TOTAL
            until [[ $TOTAL -gt $rest ]]; do
                rest=$(kubectl get cmk-nodereport ${case[$NODE]} -o json | jq .spec.report.description.pools.exclusive | jq .cpuLists | awk -F  '{' '{print $(NF)}' | awk -F  '}' '{print $(NF)}' | awk -F  ',' '{print $(NF)}' | grep "\"tasks\": \[\]" | wc -l)
            done
            let allocated=`expr $TOTAL - $rest`
            echo "The allocated CPU amount is:" $allocated
            echo "deploy a nginx pod"
            start_nginx_pod
            if [[ $allocated == ${case[$CORE]} ]]; then
                echo "CPU was allocated as expected, TC passed !!"
            else
                echo "failed to allocate CPU, TC failed !!"
            fi
            rm -f $DIR/$pod_name.yaml
            echo "ready to delete pod"
            kubectl delete pod $pod_name --ignore-not-found=true --now --wait
            echo "Pod was deleted"
            echo "##################################"
            echo
            echo
        else
            echo skip this TC, due to env mismatching
            echo "##################################"
            echo
            echo
        fi
    else
        echo "TC: to allocate CPU(s) from pool of ${case[$POOL]} on node of ${case[$NODE]}"
        if [ "${case[$NODE]}" == "$ENV" ]; then
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
      value: "/proc"
    - name: CMK_NUM_CORES
      value: "3"
    image: ubuntu:18.10
    imagePullPolicy: "IfNotPresent"
    volumeMounts:
    - mountPath: "/proc"
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
      value: "/proc"
    - name: CMK_NUM_CORES
      value: "3"
    image: ubuntu:18.10
    imagePullPolicy: "IfNotPresent"
    volumeMounts:
    - mountPath: "/proc"
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
            echo "waiting for pod up"
            for pod in $pod_name; do
                wait_for_pod_up $pod
            done
            echo "waiting for CPU allocation finished ..."
            rest=0
            timeout=0
            until [ $rest == 2 -o $timeout == 180 ]; do
                rest=$(kubectl get cmk-nodereport ${case[$NODE]} -o json | jq .spec.report.description.pools.shared | jq .cpuLists | awk -F  '{' '{print $(NF)}' | awk -F  '}' '{print $(NF)}' | grep -v "cpus" | grep "  "| grep -v "tasks"| grep -v "\]" | wc -l)
                sleep -- 1
                let timeout++
            done
            echo "The CPU allocated in shared pool for 2 tasks"
            echo "deploy a nginx pod"
            start_nginx_pod
            if [[ $rest == 2 ]]; then
                echo "CPU was allocated as expected, TC passed !!"
            else
                echo "failed to allocate CPU, TC failed !!"
            fi
            rm -f $DIR/$pod_name.yaml
            echo "ready to delete pod"
            kubectl delete pod $pod_name --ignore-not-found=true --now --wait
            echo "Pod was deleted"
            echo "##################################"
            echo
            echo
        else
            echo skip this TC, due to env mismatching
            echo "##################################"
            echo
            echo
        fi
    fi
done
