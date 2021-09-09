#### SPDX-License-Identifier: Apache-2.0
#### Copyright (c) 2021 Intel Corporation

# Installing KUD addons with emcoctl

This folder contains KUD addons to deploy with EMCO. The example
configuration assumes one edge cluster to deploy to. EMCO needs to be
installed on the cluster before deploying these addons and emcoctl
needs to be installed and configured for the edge cluster.

1. Multus CNI
2. OVN4NFV K8s Plugin
3. Node Feature Discovery
4. SR-IOV Network Operator
5. SR-IOV Network
6. QuickAssist Technology (QAT) Device Plugin
7. CPU Manager for Kubernetes
8. KubeVirt and CDI Operators
9. KubeVirt and CDI Instances

## Setup environment to deploy addons

1. Export environment variables
   - KUBE_PATH: where the kubeconfig for edge cluster is located, and
   - HOST_IP: IP address of the cluster where EMCO is installed.

#### NOTE: For HOST_IP, assuming here that nodeports are used to access all EMCO services both from outside and between the EMCO services.

2. Customize values.yaml.

    `$ envsubst < values.yaml.example > values.yaml`

## Create prerequisites to deploy addons

Apply the prerequisites. This creates the controllers, one or more
clusters, one project, and one default logical cloud. This step is
required to be done only once.

    `$ emcoctl apply -f 00-controllers.yaml -v values.yaml`
    `$ emcoctl apply -f 01-cluster.yaml -v values.yaml`
    `$ emcoctl apply -f 02-project.yaml -v values.yaml`

## Create addons project

This creates the project with the addons listed `CompositeApps` value.

    `$ emcoctl apply -f 03-addons-app.yaml -v values.yaml`

## Instantiate the addons

This instantiates each composite app listed in the `CompositeApps`
value.

NOTE: The ordering is important when both the sriov-network and
kubevirt addons are enabled.  The sriov-network addon will trigger a
drain of the nodes and kubevirt will prevent the drain from
completing, so kubevirt must be instantiated after sriov-network has
completed the drain.

	`$ emcoctl apply projects/kud/composite-apps/addons/v1/deployment-intent-groups/deployment/instantiate`
	`$ emcoctl apply projects/kud/composite-apps/networks/v1/deployment-intent-groups/deployment/instantiate`
	`$ emcoctl apply projects/kud/composite-apps/kubevirt/v1/deployment-intent-groups/deployment/instantiate`

## Cleanup

1. Terminate addons.

	`$ emcoctl apply projects/kud/composite-apps/kubevirt/v1/deployment-intent-groups/deployment/terminate`
	`$ emcoctl apply projects/kud/composite-apps/networks/v1/deployment-intent-groups/deployment/terminate`
	`$ emcoctl apply projects/kud/composite-apps/addons/v1/deployment-intent-groups/deployment/terminate`

2. Delete addons.

    `$ emcoctl delete -f 03-addons-app.yaml -v values.yaml`

3. Cleanup prerequisites.

    `$ emcoctl delete -f 02-project.yaml -v values.yaml`
    `$ emcoctl delete -f 01-cluster.yaml -v values.yaml`
    `$ emcoctl delete -f 00-controllers.yaml -v values.yaml`

#### NOTE: Known issue: Deletion of the resources fails sometimes as some resources can't be deleted before others are deleted. This can happen due to timing issue. In that case try deleting again and the deletion should succeed.
