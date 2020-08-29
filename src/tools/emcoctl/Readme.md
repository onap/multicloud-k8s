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
