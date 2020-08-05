#################################################################
# EMCO v2 helm charts
#################################################################

EMCO Helm charts include charts for EMCO microservices along with MongoDb, etcd, Flutend


### Steps to generate and install packages
**1. Create a local helm repo from Makefile**

`$ make repo`

**2. Run make file to package all the required chart**

`$ make clean`

`$ make all`

Pacakges helm charts in tar.gz format. All packages are in **dist/packages** directory and the package of intrest are:

   File      | Description |
  | ----------- | ----------- |
  | **emco-db-0.1.0.tgz**      | Includes database packages for mongo & etcd       |
  | **emco-services-0.1.0.tgz**   | Includes packages for all EMCO services like orchestrator, ncm, rsync etc        |
  | **emco-tools-0.1.0.tgz**   | Tools like Fluentd to be used with EMCO        |
  | **emco-0.1.0.tgz**   | Includes all charts including database, all services and tools        |


**3. Deploy EMCO Packages for Databases and Services**

`$ helm install dist/packages/emco-db-0.1.0.tgz --name emco-db --namespace emco`

`$ helm install dist/packages/emco-services-0.1.0.tgz --name emco-services --namespace emco`

**4. Deploy tools (Optional)**

`$ helm install dist/packages/emco-tools-0.1.0.tgz --name rel-tools --namespace emco`

    NOTE: Deploy the Chart emco-0.1.0.tgz to deploy all packages including database, services and tools.

    `$ helm install dist/packages/emco-0.1.0.tgz --name rel --namespace emco`


**5. To check logs of the different Microservices check fluentd logs**

`kubectl logs emco-fluentd-0 -n test | grep orchestrator`


**6. Delete all packages**

`$helm delete emco-services --purge`

`$helm delete emco-db --purge`

Optional if tools were installed

`$helm delete emco-tools --purge`


**7. Delete local helm repo**

`make repo-stop`

### Known Issues

After deleting the db package and before installing the package again following error happens:

        `Error: release emco-db failed: object is being deleted: persistentvolumes "emco-db-emco-etcd-data-0" already exists`

Workaround :

    `kubectl edit persistentvolumes emco-db-emco-etcd-data-0 -n emco`

    and remover finalizers section

