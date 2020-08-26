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

2. Get Emco Resources

Get the resources in the input file. This command will use the metadata name to get the resource.

`$ emcoctl get -f filename.yaml`

3. Delete Emco Resources

Delete resources in the file. The emcoctl will start deleting resources in the reverse order than given in the file to maintain hierarchy. This command will use the metadata name to delete the resource.

`$ emcoctl delete -f filename.yaml`

4. Get all Emco Resources

Get all for the resources in the file.

`$ emcoctl getall -f filename.yaml`
