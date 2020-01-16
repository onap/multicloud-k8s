# vIPSec use case in ONAP
This use case is composed of four virtual functions (VFs) including two
IPSec gateways, a packet generator and a traffic sink, each running in
separate Ubuntu Virtual Machines:

  * [Packet generator][1]: Sends packets to the packet sink through the
tunnel constructed thru IPSec. This includes a script that installs the
packet generator based on packetgen[4].
  * [IPsec gateways][2]: Two IPSec gateways constructed the secure tunnel
for traffic transportation. This includes a script to install and configure
the IPSec gateways thru VPP.
  * [Traffic sink][3]: Displays the traffic volume that lands at the sink
VM using the link http://192.168.80.250:667 through your browser
and enable automatic page refresh by clicking the "Off" button. You
can see the traffic volume in the charts.

This set of scripts aims to construct the vIPSec use case in order to set
up a secure tunnel between peers and improve its performance along with
hardware acceleration technologies such as SRIOV and QAT.

User can apply the helm chart named 'vipsec' inside the k8s/kud/demo folder
to set up the whole use case. A fully-functional Kubernetes cluster, Virtlet
as well as ovn4nfv-k8s[5] plugin need to be pre-installed for the usage.
*[Place needs improvements] After having the virtual machines ready, please
manually change the MAC address inside the ipsec.conf to enable the routing.
And also start up the packetgen to send packet with src and dst defined in
the templates/values.yaml inside the helm chart. Detail instructions will be
put inside the helm chart.

If you'd like to test the performance with QAT/SRIOV involved, first get
these hardwares pre-configured. Then change the value of 'qat_enabled' and
'sriov_enabled' inside templates/values.yaml of the helm chart accordingly.
User could observe variance in throughput inside the traffic sink.

[4] https://pktgen-dpdk.readthedocs.io/en/latest/
[5] https://github.com/opnfv/ovn4nfv-k8s-plugin
