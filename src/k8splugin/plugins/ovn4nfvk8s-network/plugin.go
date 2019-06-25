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

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"
	v1 "github.com/onap/multicloud-k8s/src/k8splugin/plugins/network/v1"

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
func CreateNetwork(network *v1.OnapNetwork) (string, error) {

	name := network.Spec.Name
	if name == "" {
		return "", pkgerrors.New("Invalid Network Name")
	}

	subnet := network.Spec.Subnet
	if subnet == "" {
		return "", pkgerrors.New("Invalid Subnet Address")
	}

	gatewayIPMask := network.Spec.Gateway
	if gatewayIPMask == "" {
		return "", pkgerrors.New("Invalid Gateway Address")
	}

	routerMac, stderr, err := ovnCmd.Run(getAuthStr(), "--if-exist", "-v", "get", "logical_router_port", "rtos-"+name, "mac")
	if err != nil {
		return "", pkgerrors.Wrapf(err, "Failed to get logical router port,stderr: %q, error: %v", stderr, err)
	}

	if routerMac == "" {
		log.Print("Generate MAC address")
		prefix := "00:00:00"
		newRand := rand.New(rand.NewSource(time.Now().UnixNano()))
		routerMac = fmt.Sprintf("%s:%02x:%02x:%02x", prefix, newRand.Intn(255), newRand.Intn(255), newRand.Intn(255))
	}

	_, stderr, err = ovnCmd.Run(getAuthStr(), "--may-exist", "lrp-add", ovn4nfvRouter, "rtos-"+name, routerMac, gatewayIPMask)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "Failed to add logical port to router, stderr: %q, error: %v", stderr, err)
	}

	// Create a logical switch and set its subnet.
	stdout, stderr, err := ovnCmd.Run(getAuthStr(), "--", "--may-exist", "ls-add", name, "--", "set", "logical_switch", name, "other-config:subnet="+subnet, "external-ids:gateway_ip="+gatewayIPMask)
	if err != nil {
		return "", pkgerrors.Wrapf(err, "Failed to create a logical switch %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
	}

	// Connect the switch to the router.
	stdout, stderr, err = ovnCmd.Run(getAuthStr(), "--", "--may-exist", "lsp-add", name, "stor-"+name, "--", "set", "logical_switch_port", "stor-"+name, "type=router", "options:router-port=rtos-"+name, "addresses="+"\""+routerMac+"\"")
	if err != nil {
		return "", pkgerrors.Wrapf(err, "Failed to add logical port to switch, stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
	}

	return name, nil
}

// DeleteNetwork in OVN controller
func DeleteNetwork(name string) error {
	log.Printf("Deleting Network: Ovn4nfvk8s %s", name)

	stdout, stderr, err := ovnCmd.Run(getAuthStr(), "--if-exist", "ls-del", name)
	if err != nil {
		return pkgerrors.Wrapf(err, "Failed to delete switch %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
	}

	stdout, stderr, err = ovnCmd.Run(getAuthStr(), "--if-exist", "lrp-del", "rtos-"+name)
	if err != nil {
		return pkgerrors.Wrapf(err, "Failed to delete router port %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
	}

	stdout, stderr, err = ovnCmd.Run(getAuthStr(), "--if-exist", "lsp-del", "stor-"+name)
	if err != nil {
		return pkgerrors.Wrapf(err, "Failed to delete switch port %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
	}

	return nil
}

func getAuthStr() string {
	//TODO: Remove hardcoding: Use ESR data passed to Initialize
	ovnCentralAddress := config.GetConfiguration().OVNCentralAddress
	return "--db=tcp:" + ovnCentralAddress
}
