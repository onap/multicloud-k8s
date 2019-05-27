/*
Copyright 2018 Intel Corporation.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"k8splugin/internal/connection"
	v1 "k8splugin/plugins/network/v1"

	pkgerrors "github.com/pkg/errors"
	kexec "k8s.io/utils/exec"
)

const (
	ovn4nfvRouter   = "ovn4nfv-master"
	ovnNbctlCommand = "ovn-nbctl"
)

type OVNNbctler interface {
	Run(args ...string) (string, string, error)
}

type OVNNbctl struct {
	run  func(args ...string) (string, string, error)
	exec kexec.Interface
	path string
}

// Run a command via ovn-nbctl
func (ctl *OVNNbctl) Run(args ...string) (string, string, error) {
	if ctl.exec == nil {
		ctl.exec = kexec.New()
	}
	if ctl.path == "" {
		nbctlPath, err := ctl.exec.LookPath(ovnNbctlCommand)
		if err != nil {
			return "", "", pkgerrors.Wrap(err, "Look nbctl path error")
		}
		ctl.path = nbctlPath
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := ctl.exec.Command(ctl.path, args...)
	cmd.SetStdout(stdout)
	cmd.SetStderr(stderr)
	err := cmd.Run()

	return strings.Trim(strings.TrimFunc(stdout.String(), unicode.IsSpace), "\""),
		stderr.String(), err
}

var ovnCmd OVNNbctler

func init() {
	ovnCmd = &OVNNbctl{}
}

// CreateNetwork in OVN controller
func CreateNetwork(network *v1.OnapNetwork, cloudRegion string) (string, error) {

	ovnCentralAddress := getAuthStr(cloudRegion)

	name := network.Spec.Name
	if name == "" {
		return "", pkgerrors.New("Invalid Network Name")
	}
	log.Printf("Creating Network: Ovn4nfvk8s %s", name)

	subnet := network.Spec.Subnet
	if subnet == "" {
		return "", pkgerrors.New("Invalid Subnet Address")
	}

	gatewayIPMask := network.Spec.Gateway
	if gatewayIPMask == "" {
		return "", pkgerrors.New("Invalid Gateway Address")
	}

	routerMac, stderr, err := ovnCmd.Run(ovnCentralAddress, "--if-exist", "-v", "get", "logical_router_port", "rtos-"+name, "mac")
	if err != nil {
		return "", pkgerrors.Wrapf(err, "Failed to get logical router port,stderr: %q, error: %v", stderr, err)
	}

	if routerMac == "" {
		log.Print("Generate MAC address")
		prefix := "00:00:00"
		newRand := rand.New(rand.NewSource(time.Now().UnixNano()))
		routerMac = fmt.Sprintf("%s:%02x:%02x:%02x", prefix, newRand.Intn(255), newRand.Intn(255), newRand.Intn(255))
	}

	_, stderr, err = ovnCmd.Run(ovnCentralAddress, "--may-exist", "lrp-add", ovn4nfvRouter, "rtos-"+name, routerMac, gatewayIPMask)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "Failed to add logical port to router, stderr: %q, error: %v", stderr, err)
	}

	// Create a logical switch and set its subnet.
	stdout, stderr, err := ovnCmd.Run(ovnCentralAddress, "--", "--may-exist", "ls-add", name, "--", "set", "logical_switch", name, "other-config:subnet="+subnet, "external-ids:gateway_ip="+gatewayIPMask)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "Failed to create a logical switch %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
	}

	// Connect the switch to the router.
	stdout, stderr, err = ovnCmd.Run(ovnCentralAddress, "--", "--may-exist", "lsp-add", name, "stor-"+name, "--", "set", "logical_switch_port", "stor-"+name, "type=router", "options:router-port=rtos-"+name, "addresses="+"\""+routerMac+"\"")
	if err != nil {
		return "", pkgerrors.Wrapf(err, "Failed to add logical port to switch, stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
	}

	return name, nil
}

// DeleteNetwork in OVN controller
func DeleteNetwork(name, cloudRegion string) error {
	log.Printf("Deleting Network: Ovn4nfvk8s %s", name)
	ovnCentralAddress := getAuthStr(cloudRegion)

	stdout, stderr, err := ovnCmd.Run(ovnCentralAddress, "--if-exist", "ls-del", name)
	if err != nil {
		return pkgerrors.Wrapf(err, "Failed to delete switch %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
	}

	stdout, stderr, err = ovnCmd.Run(ovnCentralAddress, "--if-exist", "lrp-del", "rtos-"+name)
	if err != nil {
		return pkgerrors.Wrapf(err, "Failed to delete router port %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
	}

	stdout, stderr, err = ovnCmd.Run(ovnCentralAddress, "--if-exist", "lsp-del", "stor-"+name)
	if err != nil {
		return pkgerrors.Wrapf(err, "Failed to delete switch port %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
	}

	return nil
}

func getConnInfo(conn map[string]interface{}) (string, string) {
	var ipAddress, port string
	for key, value := range conn {
		if key == "ovn-ip-address" {
			if str, ok := value.(string); ok {
				ipAddress = str
			} else {
				return "", ""
			}
		}
		if key == "ovn-port" {
			if str, ok := value.(string); ok {
				port = str
			} else {
				return "", ""
			}
		}
	}
	return ipAddress, port
}

func getAuthStr(cloudRegion string) string  {
	conn := connection.NewConnectionClient()
	connInfo, err := conn.Get(cloudRegion)
	if err != nil {
		return ""
	}
	ipAddress, port := getConnInfo(connInfo.OtherConnectivityList)
	if ipAddress == "" || port == "" {
		return ""
	} else {
		ovnCentralAddress := ipAddress + ":" + port
                log.Printf("ovnCentralAddress: %v", ovnCentralAddress)
		return "--db=tcp:" + ovnCentralAddress
	}
}
