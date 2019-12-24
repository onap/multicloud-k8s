# vIPSec use case in ONAP
This use case is composed of four virtual functions (VFs) including two 
IPSec gateways, a packet generator and a traffic sink, each running in
separate Ubuntu Virtual Machines:

  * [Packet generator][1]: Sends packets to the packet sink through the
tunnel constructed thru IPSec. This includes a script that installs the
packet generator based on packetgen.
  * [IPsec gateways][2]: Two IPSec gateways constructed the secure tunnel
for traffic transportation. This includes a script to install and configure
the IPSec gateways thru VPP. 
  * [Traffic sink][3]: Displays the traffic volume that lands at the sink
VM using the link http://192.168.20.250:667 through your browser
and enable automatic page refresh by clicking the "Off" button. You
can see the traffic volume in the charts.

