#!/bin/bash
cases=("exclusive localhost 3" "exclusive compute01 3" "shared localhost 3" "shared compute01 3")
case=(null null 0)
num=${#cases[*]}
POOL=0
NODE=1
CORE=2
DIR=/tmp
pod_name=cmk-test-pod
rm -f $DIR/$pod_name.yaml
kubectl delete pod $pod_name --ignore-not-found=true --now --wait
ENV=$(kubectl get nodes --all-namespaces | awk 'NR==2{print $1}')
echo "env is $ENV"
echo
for ((i=0;i<$num;i++)); do
    inner_case=(${cases[$i]})
    num_inner=${#inner_case[*]}
    for ((j=0;j<$num_inner;j++)); do
        case[$j]=${inner_case[$j]}
    done
    echo "##################################"
    echo "TC: to allocate ${case[$CORE]} CPU(s) from pool of ${case[$POOL]} on node of ${case[$NODE]}"
    if [ "${case[$NODE]}" == "$ENV" ]; then
        TOTAL=$(kubectl get cmk-nodereport ${case[$NODE]} -o json | jq .spec.report.description.pools.exclusive | jq .cpuLists | awk -F  '{' '{print $(NF)}' | awk -F  '}' '{print $(NF)}' | awk -F  ',' '{print $(NF)}' | grep "\"tasks\": \[" | wc -l)
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
        - "/opt/bin/cmk isolate --conf-dir=/etc/cmk --pool=${case[$POOL]} sleep -- 3900"
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
          status_phase=""
          while [[ $status_phase != "Running" ]]; do
              new_phase=$(kubectl get pods $pod | awk 'NR==2{print $3}')
              if [[ $new_phase != $status_phase ]]; then
                  echo "$(date +%H:%M:%S) - $pod : $new_phase"
                  status_phase=$new_phase
              fi
              if [[ $new_phase == "Running" ]]; then
                  echo "Pod is up and running.."
              fi
              if [[ $new_phase == "Err"* ]]; then
                  exit 1
              fi
           done
        done
        echo "waiting for CPU allocation finished ..."
        rest=$TOTAL
        until [[ $TOTAL -gt $rest ]]; do
               rest=$(kubectl get cmk-nodereport ${case[$NODE]} -o json | jq .spec.report.description.pools.exclusive | jq .cpuLists | awk -F  '{' '{print $(NF)}' | awk -F  '}' '{print $(NF)}' | awk -F  ',' '{print $(NF)}' | grep "\"tasks\": \[\]" | wc -l)
        done
        let allocated=`expr $TOTAL - $rest`
        echo "The allocated CPU amount is:" $allocated
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
done

