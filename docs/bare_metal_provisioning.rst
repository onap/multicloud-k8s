.. Copyright 2018 Intel Corporation.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at
        http://www.apache.org/licenses/LICENSE-2.0
   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

***********************
Bare-Metal Provisioning
***********************

The Kubernetes Deployment, aka KUD, has been designed to be consumed
by Virtual Machines as well as Bare-Metal servers. The *baremetal/aio.sh*
script contains the bash instructions for provisioning an All-in-One Kubernetes
deployment in a Bare-Metal server. This document lists the Hardware & Software
requirements and walkthrough the instructions that *baremetal/aio.sh* contains.

Hardware Requirements
#####################

+-----------+--------+
| Concept   | Amount |
+===========+========+
| CPUs      | 8      |
+-----------+--------+
| Memory    | 32GB   |
+-----------+--------+
| Hard Disk | 150GB  |
+-----------+--------+

Software Requirements
#####################

- Ubuntu Server 16.04 LTS

baremetal/aio.sh
################

This bash script provides an automated process for deploying an All-in-One
Kubernetes cluster.

The following two instructions start the provisioning process.

.. code-block:: bash

    $ sudo su
    # git clone https://git.onap.org/multicloud/k8s/
    # cd k8s/kud/hosting_providers/baremetal/
    # ./aio.sh

In overall, this script can be summarized in three general phases:

1. Generating Inventory.
2. Enabiling Nested-Virtualization.
3. Deploying KUD services.

**Inventory**

Ansible works agains multiple systems, the way for selecting them is through the
usage of the inventory. The inventory file is a static source for determining the
target servers used for the execution of ansible tasks. The *aio.sh* script creates
an inventory file for addressing those tasks to localhost.

.. code-block:: bash

    cat <<EOL > inventory/hosts.ini
    [all]
    localhost

    [kube-master]
    localhost

    [kube-node]
    localhost

    [etcd]
    localhost

    [ovn-central]
    localhost

    [ovn-controller]
    localhost

    [virtlet]
    localhost

    [k8s-cluster:children]
    kube-node
    kube-master
    EOL

KUD consumes kubespray_ for provisioning a Kubernetes base deployment.

.. _kubespray: https://github.com/kubernetes-incubator/kubespray

Ansible uses SSH protocol for executing remote instructions. The following
instructions create and register ssh keys which avoid the usage of passwords.

.. code-block:: bash

    # echo -e "\n\n\n" | ssh-keygen -t rsa -N ""
    # cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
    # chmod og-wx ~/.ssh/authorized_keys

**Enabling Nested-Virtualization**

KUD installs Virtlet_ Kubernetes CRI for running Virtual Machine workloads.
Nested-virtualization gives the ability of running a Virtual Machine within
another. The *node.sh* bash script contains the instructions for enabling
Nested-Virtualization.

.. _Virtlet : https://github.com/Mirantis/virtlet

.. code-block:: bash

    # ./node.sh

**Deploying KUD services**

Finally, the KUD provisioning process can be started through the use of
*installer.sh* bash script. The output of this script is collected in the
*kud_installer.log* file for future reference.

.. code-block:: bash

    # ./installer.sh | tee kud_installer.log

.. image:: ./img/installer_workflow.png
