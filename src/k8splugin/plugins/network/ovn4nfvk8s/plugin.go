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
	"log"
	"k8splugin/krd"
	 kexec "k8s.io/utils/exec"
	"k8s.io/client-go/kubernetes"
	pkgerrors "github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"os"
	"io/ioutil"
	"time"
	"math/rand"
)

type Ovn4NfvK8sFile struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
			   Name    string `json:"name"`
			   Cnitype string `json:"cnitype"`
		   } `json:"metadata"`
	Spec struct {
			   Name    string `json:"name"`
			   Subnet  string `json:"subnet"`
			   Gateway string `json:"gateway"`
			   Routes  []struct {
				   Dst string `json:" "dst"`
				   Gw  string `json:"gw"`
			   } `json:"routes"`
		   } `json:"spec"`
}

const  (
	ovn4nfvRouter = "ovn4nfv-master"
	ovnNbctlCommand   = "ovn-nbctl"
	ipCommand         = "ip"
)
var initData *krd.InitData

// Exec runs various OVN and OVS utilities
type execHelper struct {
	exec           kexec.Interface
	nbctlPath      string
	ipPath         string
}

var runner *execHelper



func Initialize(data *krd.InitData) (string, error) {

	log.Printf("Intiialize Ovn4nfvk8s Network Plugin")

	initData = data

	exec := kexec.New()
	if err := SetExec(exec); err != nil {
		return "", err
	}
	return "ovn4nfvk8s", nil
}

// Reads network yaml
var ReadNetworkFile = func(path string) (Ovn4NfvK8sFile, error) {
	var ovnFile Ovn4NfvK8sFile

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ovnFile, pkgerrors.Wrap(err, "Network YAML file does not exist")
	}

	log.Println("Reading Network YAML: " + path)
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return ovnFile, pkgerrors.Wrap(err, "Network YAML file read error")
	}
	err = yaml.Unmarshal(yamlFile, &ovnFile)
	if err != nil {
		return ovnFile, pkgerrors.Wrap(err, "Network YAML file unmarshal error")
	}
	log.Printf("network:\n%v", ovnFile)

	return ovnFile, nil
}

// Create a OVN Network
func CreateNetwork(data *krd.ResourceData, client kubernetes.Interface) (string, error) {

	log.Println("CreateNetwork: Ovn4nfvk8s")

	ovnFile, err := ReadNetworkFile(data.YamlFilePath )
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error while reading File: "+data.YamlFilePath)
	}

	err = createOvnNetwork(ovnFile.Spec.Name, ovnFile.Spec.Subnet, ovnFile.Spec.Gateway)
	if err != nil {
		return "", err
	}
	return ovnFile.Spec.Name, nil
}

// Delete a OVN Network
func DeleteNetwork(name string, client kubernetes.Interface) (error) {

	log.Println("Deleting Network: Ovn4nfvk8s %s", name)

	err := deleteOvnNetwork(name)
	if err != nil {
		return err
	}
	return nil
}

// Get a OVN Network
func GetNetwork(name string, client kubernetes.Interface) (string, error) {

	log.Println("GetNetwork: Ovn4nfvk8s %s", name)
	return "", nil
}

// SetExec validates executable paths and saves the given exec interface
// to be used for running various OVS and OVN utilites
func SetExec(exec kexec.Interface) error {
	var err error

	runner = &execHelper{exec: exec}
	runner.nbctlPath, err = exec.LookPath(ovnNbctlCommand)
	if err != nil {
		return err
	}
	runner.ipPath, err = exec.LookPath(ipCommand)
	if err != nil {
		return err
	}
	return nil
}

func run(cmdPath string, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := runner.exec.Command(cmdPath, args...)
	cmd.SetStdout(stdout)
	cmd.SetStderr(stderr)
	//log.Println("exec: %s %s", cmdPath, strings.Join(args, " "))
	err := cmd.Run()
	if err != nil {
		log.Println("exec: %s %s => %v", cmdPath, strings.Join(args, " "), err)
	}
	return stdout, stderr, err
}

// RunOVNNbctl runs command via ovn-nbctl
func RunOVNNbctl(args ...string) (string, string, error) {

	stdout, stderr, err := run(runner.nbctlPath, args...)
	return strings.Trim(strings.TrimFunc(stdout.String(), unicode.IsSpace), "\""),
		stderr.String(), err
}

func RunIP(args ...string) (string, string, error) {
	stdout, stderr, err := run(runner.ipPath, args...)
	return strings.TrimSpace(stdout.String()), stderr.String(), err
}

// GenerateMac generates mac address.
func GenerateMac() string {
	prefix := "00:00:00"
	newRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	mac := fmt.Sprintf("%s:%02x:%02x:%02x", prefix, newRand.Intn(255), newRand.Intn(255), newRand.Intn(255))
	return mac
}

func getAuthStr()(string) {

	//TODO: Remove hardcoding: Use ESR data passed to Initialize
	return "--db=tcp:10.10.10.3:6641"
}

func createOvnNetwork(name, subnet, gatewayIPMask string) (error) {

	routerMac, stderr, err := RunOVNNbctl(getAuthStr(), "--if-exist", "-v", "get", "logical_router_port", "rtos-"+name, "mac")
	//routerMac, stderr, err := RunOVNNbctl( "--if-exist",  "get", "logical_router_port", "rtos-"+name, "mac")
	if err != nil {
		log.Println("Failed to get logical router port,stderr: %q, error: %v", stderr, err)
		return err
	}

	if routerMac == "" {
		routerMac = GenerateMac()
	}

	_, stderr, err = RunOVNNbctl(getAuthStr(), "--may-exist", "lrp-add", ovn4nfvRouter, "rtos-"+name, routerMac, gatewayIPMask)
	if err != nil {
		log.Println("Failed to add logical port to router, stderr: %q, error: %v", stderr, err)
		return err
	}

	// Create a logical switch and set its subnet.
	stdout, stderr, err := RunOVNNbctl(getAuthStr(), "--", "--may-exist", "ls-add", name, "--", "set", "logical_switch", name, "other-config:subnet="+subnet, "external-ids:gateway_ip="+gatewayIPMask)
	if err != nil {
		log.Println("Failed to create a logical switch %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
		return err
	}

	// Connect the switch to the router.
	stdout, stderr, err = RunOVNNbctl(getAuthStr(), "--", "--may-exist", "lsp-add", name, "stor-"+name, "--", "set", "logical_switch_port", "stor-"+name, "type=router", "options:router-port=rtos-"+name, "addresses="+"\""+routerMac+"\"")
	if err != nil {
		log.Println("Failed to add logical port to switch, stdout: %q, stderr: %q, error: %v", stdout, stderr, err)
		return err
	}
	return nil

}

func deleteOvnNetwork(name string) (error) {

	stdout, stderr, err := RunOVNNbctl(getAuthStr(), "--if-exist", "ls-del", name)
	if err != nil {
		log.Println("Failed to delete switch %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
		return err
	}

	stdout, stderr, err = RunOVNNbctl(getAuthStr(),  "--if-exist", "lrp-del", "rtos-"+name)
	if err != nil {
		log.Println("Failed to delete router port %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
		return err
	}

	stdout, stderr, err = RunOVNNbctl(getAuthStr(), "--if-exist", "lsp-del", "stor-"+name)
	if err != nil {
		log.Println("Failed to delete switch port %v, stdout: %q, stderr: %q, error: %v", name, stdout, stderr, err)
		return err
	}
	return nil

}
