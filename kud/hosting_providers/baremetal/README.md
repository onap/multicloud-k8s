# Kubernetes Deployment

## Summary

This project offers a means for deploying a Kubernetes cluster
that satisfies the requirements of [ONAP multicloud/k8s plugin][1]. Its
ansible playbooks allow to provision a deployment on Baremetal. 

![Diagram](../../../docs/img/installer_workflow.png)


These bash scripts contains the minimal Ubuntu instructions required for running this project.

Note that these scripts must be run as root user.

`sudo -s`

## Configuring KUD project

### Setup KUD

Run [All-in-one KUD installer](aio.sh) to setup KUD.

Note: Should the following error occur: `ImportError: No module named _internal.cli.main`

Run `apt remove python-pip`

Retry [All-in-one KUD installer](aio.sh)

### Generate ssh-keys

`echo -e "\n\n\n" | ssh-keygen -t rsa -N ""`

`cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys`

`chmod og-wx ~/.ssh/authorized_keys`

### Enable nested virtualization

Run [Enable nested virtualization](../vagrant/node.sh)


## Deploying KUD Services

Run [installer](../vagrant/installer.sh)

NOTE: for cmk bare metal deployment, preset 1/2 CPUs for
      shared/exlusive pools respectively to fit CI server machines
      users can adjust the parameters to meet their own requirements.

## License

Apache-2.0

[1]: https://git.onap.org/multicloud/k8s
