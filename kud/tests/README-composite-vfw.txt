# Notes on running the composite vFW test case

# Infrastructure
As written, the vfw-test.sh script assumes 3 clusters
1 - the cluster in which the EMCO microservices are running
2 - two edge clusters in which the vFW will be instantiated

The edge cluster in which vFW will be instantiated should be KUD clusters.

# Edge cluster preparation

For status monitoring support, the 'monitor' docker image must be built and
deployed.

In multicloud-k8s repo:
	cd multicloud-k8s/src/monitor
 	docker build -f build/Dockerfile . -t monitor
	<tag and push docker image to dockerhub ...>

Deploy monitor program in each cluster (assumes multicloud-k8s repo is present in cloud)
	# one time setup per cluster - install the CRD
	cd multicloud-k8s/src/monitor/deploy
	kubectl apply -f crds/k8splugin_v1alpha1_resourcebundlestate_crd.yaml
	
	# one time setup per cluster
	# update yaml files with correct image
	# (cleanup first, if monitor was already installed - see monitor-cleanup.sh)
	cd multicloud-k8s/src/monitor/deploy
	monitor-deploy.sh


# Preparation of the vFW Composit Application

## Prepare the Composite vFW Application Charts and Profiles

1. In the multicloud-k8s/kud/demo/composite-firewall directory, prepare the 3 helm
   charts for the vfw.

   tar cvf packetgen.tar packetgen
   tar cvf firewall.tar firewall
   tar cvf sink.tar sink
   gzip *.tar

2. Prepare the profile file (same one will be used for all profiles in this demo)

   tar cvf profile.tar manifest.yaml override_values.yaml
   gzip profile.tar

## Set up environment variables for the vfw-test.sh script

The vfw-test.sh script expects a number of files to be provided via environment
variables.

Change directory to multicloud-k8s/kud/tests

1.  Edge cluster kubeconfig files - the script expects 2 of these

    export kubeconfigfile=<path to first cluster kube config file>
    export kubeconfigfile2=<path to second cluster kube config file>

    for example:  export kubeconfigfile=/home/vagrant/multicloud-k8s/cluster-configs/config-edge01


2.  Composite app helm chart files (as prepared above)

    export packetgen_helm_path=../demo/composite-firewall/packetgen.tar.gz
    export firewall_helm_path=../demo/composite-firewall/firewall.tar.gz
    export sink_helm_path=../demo/composite-firewall/sink.tar.gz

3.  Composite profile application profiles (as prepared above)

    export packetgen_profile_targz=../demo/composite-firewall/profile.tar.gz
    export firewall_profile_targz=../demo/composite-firewall/profile.tar.gz
    export sink_profile_targz=../demo/composite-firewall/profile.tar.gz

4.  Modify the script to address the EMCO cluster

    Modifiy the urls at the top part of the script to point to the
    cluster IP address of the EMCO cluster.

    That is, modify the IP address 10.10.10.6 to the correct value for
    your environment.

    Note also that the node ports used in the following are based on the values
    defined in multicloud-k8s/deployments/kubernetes/onap4k8s.yaml

        base_url_clm=${base_url_clm:-"http://10.10.10.6:31856/v2"}
        base_url_ncm=${base_url_ncm:-"http://10.10.10.6:32737/v2"}
        base_url_orchestrator=${base_url_orchestrator:-"http://10.10.10.6:31298/v2"}
        base_url_ovnaction=${base_url_ovnaction:-"http://10.10.10.6:31181/v2"}


# Run the vfw-test.sh

The rest of the data needed for the test is present in the script.

1.  Invoke API calls to create the data
    
    vfw-test.sh create

    This does all of the data setup
    - registers clusters
    - registers controllers
    - sets up the composite app and profile
    - sets up all of the intents

2.  Query results (optional)

    vfw-test.sh get

3.  Apply the network intents

    For the vFW test, the 3 networks used by the vFW are created by using network intents.
    Both virtual and provider networks are used.

    vfw-test.sh apply

    On the edge clusters, check to see the networks were created:

    kubectl get network
    kubectl get providernetwork

4.  Instantiate the vFW

    vfw-test.sh instantiate

    This will instantiate the vFW on the two edge clusters (as defined by the generic
    placement intent).

5. Status query

   vfw-test.sh status

6. Terminate
   Terminate will remove the resources from the clusters and delete the internal
   composite application information in the etcd base AppContext.
   The script will do it for both the deployment intent group (i.e. the vfW composite
   app) and the network intents.

   In principle, after runnin terminate, the 'apply' and 'instantiate' commands could
   be invoked again to re-insantiate the networks and the vFW composite app.

   vfw-test.sh terminate

7. Delete the data
   After running 'terminate', the 'delete' command can be invoked to remove all
   the data created.  This should leave the system back in the starting state -
   begin with point #1 above to start again.

   vfw-test.sh delete
