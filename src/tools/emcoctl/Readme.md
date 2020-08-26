#################################################################
# EMCOCTL - CLI for EMCO
#################################################################

Emoctl is command line tool for interacting with EMCO.
All commands take input a file. An input file can contain one or more resources.


### Syntax for describing a resource

```
version: <domain-name>/<api-version>
resourceContext:
  anchor: <URI>
Metadata :
   Name: <name>
   Description: <text>
   userData1: <text>
   userData2: <text>
Spec:
  <key>: <value>
```

### Example resource file

```
version: emco/v2
resourceContext:
  anchor: projects
Metadata :
   Name: proj1
   Description: test
   userData1: test1
   userData2: test2

---
version: emco/v2
resourceContext:
  anchor: projects/proj1/composite-apps
Metadata :
  name: vFw-demo
  description: test
  userData1: test1
  userData2: test2
spec:
  version: v1
```

### EMCO CLI Commands

1. Create Emco Resources

This command will apply the resources in the file. The user is responsible to ensuring the hierarchy of the resources.

`$ emcoctl apply -f filename.yaml`

For applying resources that don't have a json body anchor can be provided as an arguement

`$ emcoctl apply <anchor>`

`$ emcoctl apply projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group/instantiate`


2. Get Emco Resources

Get the resources in the input file. This command will use the metadata name in each of the resources in the file to get information about the resource.

`$ emcoctl get -f filename.yaml`

For getting information for one resource anchor can be provided as an arguement

`$ emcoctl get <anchor>`

`$ emcoctl get projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group`

3. Delete Emco Resources

Delete resources in the file. The emcoctl will start deleting resources in the reverse order than given in the file to maintain hierarchy. This command will use the metadata name in each of the resources in the file to delete the resource..

`$ emcoctl delete -f filename.yaml`

For deleting one resource anchor can be provided as an arguement

`$ emcoctl delete <anchor>`

`$ emcoctl delete projects/testvfw/composite-apps/compositevfw/v1/deployment-intent-groups/vfw_deployment_intent_group`


## Using helm charts through emcoctl

When you need to use emcoctl for deploying helm
charts the following steps are required.

1. Make sure that the composite app which you are planning to deploy, the tree structure is as below

```

$  tree collection/app1/
collection/app1/
├── helm
│   └── collectd
│       ├── Chart.yaml
│       ├── resources
│       │   └── collectd.conf
│       ├── templates
│       │   ├── configmap.yaml
│       │   ├── daemonset.yaml
│       │   ├── _helpers.tpl
│       │   ├── NOTES.txt
│       │   └── service.yaml
│       └── values.yaml
└── profile
    ├── manifest.yaml
    └── override_values.yaml

5 directories, 10 files

$  tree collection/m3db/
collection/m3db/
├── helm
│   └── m3db
│       ├── Chart.yaml
│       ├── del.yaml
│       ├── templates
│       │   └── m3dbcluster.yaml
│       └── values.yaml
└── profile
    ├── manifest.yaml
    └── override_values.yaml

4 directories, 6 files

```

### NOTE
```
* In the above example, we have a composite app : collection
The collection composite-app shown has two apps : app1(collectd)
and m3db
* Each app has two dirs : a. HELM and b. PROFILE.
* Helm dir shall have the real helm charts of the app.
* profile shall have the two files - manifest.yaml and override_values.yaml for creating the customized profile.
```

### Commands for making the tar files from helm.

```
    tar -czf collectd.tar.gz -C $test_folder/vnfs/comp-app/collection/app1/helm .
    tar -czf collectd_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/app1/profile .
    ----------------------------------------
    tar -czf m3db.tar.gz -C $test_folder/vnfs/comp-app/collection/m3db/helm .
    tar -czf m3db_profile.tar.gz -C $test_folder/vnfs/comp-app/collection/m3db/profile .
```

Once you have generated the tar files, you need to give the path in file which you are applying using the emcoctl. For eg:

```
#adding collectd app to the composite app
version: emco/v2
resourceContext:
  anchor: projects/proj1/composite-apps/collection-composite-app/v1/apps
metadata :
  name: collectd
  description: "description for app"
  userData1: test1
  userData2: test2
file:
  /opt/csar/cb009bfe-bbee-11e8-9766-525400435678/collectd.tar.gz

```

```
#adding collectd app profiles to the composite profile
version: emco/v2
resourceContext:
  anchor: projects/proj1/composite-apps/collection-composite-app/v1/composite-profiles/collection-composite-profile/profiles
metadata :
  name: collectd-profile
  description: test
  userData1: test1
  userData2: test2
spec:
  app-name: collectd
file:
  /opt/csar/cb009bfe-bbee-11e8-9766-525400435678/collectd_profile.tar.gz

```

### Running the emcoctl

```
* Make sure that the emcoctl is build.You can build it by issuing the 'make' command.
Dir : $MULTICLOUD-K8s_HOME/src/tools/emcoctl
```
* Then run the emcoctl by command:
```
./emcoctl --config ./examples/emco-cfg.yaml apply -f ./examples/test.yaml

```

Here, emco-cfg.yaml contains the config/port details of each of the microservices you are using.
A sample configuration is :

```
  orchestrator:
    host: localhost
    port: 9015
  clm:
    host: localhost
    port: 9019
  ncm:
    host: localhost
    port: 9016
  ovnaction:
    host: localhost
    port: 9051
```
