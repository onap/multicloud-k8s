# Notes on running the composite vFW test case

# Infrastructure
As written, the vfw-test.sh script assumes 3 clusters
1 - the cluster in which the EMCO microservices are running
2 - two edge clusters in which the vFW will be instantiated

The edge cluster in which vFW will be instantiated should be KUD clusters.

# Preparations

## Prepare the Composite vFW Application Charts and Profiles

1. In the multicloud-k8s/kud/demo/composite-vfw directory, prepare the 3 helm
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

    export packetgen_helm_chart=../demo/composite-firewall/packetgen.tar.gz
    export firewall_helm_chart=../demo/composite-firewall/firewall.tar.gz
    export sink_helm_chart=../demo/composite-firewall/sink.tar.gz

3.  Composite profile application profiles (as prepared above)

    export packetgen_profile_file=../demo/composite-firewall/profile.tar.gz
    export firewall_profile_file=../demo/composite-firewall/profile.tar.gz
    export sink_profile_file=../demo/composite-firewall/profile.tar.gz

4.  Modify the script to address the EMCO cluster

    Modifiy the the urls at the top part of the script to point to the
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


# Removing resources (until termination sequences are completed)

1. Run the cleanup script (or equivalent) in the edge clusters.
   (once the terminate flow via EMCO is complete, this step will not be necessary)

   bash cleanup-composite-vfw.sh

2. Terminate the network intents

   vfw-test.sh terminate

3. Delete everything from the Mongo DB

   vfw-test.sh delete
