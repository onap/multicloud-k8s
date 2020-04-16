/*
Copyright 2020 Intel Corporation.
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

package grpc

import (
	"os"
	"strconv"
	"strings"

	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	module "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module"
)

const default_host = "localhost"
const default_port = 9030
const default_ncm_name = "ncm"
const ENV_NCM_NAME = "NCM_NAME"

func GetServerHostPort() (string, int) {

	// expect name of this ncm program to be in env variable "NCM_NAME" - e.g. NCM_NAME="ncm"
	serviceName := os.Getenv(ENV_NCM_NAME)
	if serviceName == "" {
		serviceName = default_ncm_name
		log.Info("Using default name for NCM service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. NCM_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for ncm gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. NCM_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		log.Info("Using default port for ncm gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}

func RegisterGrpcServer(host string, port int) error {
	// expect name of this ncm program to be in env variable "NCM_NAME" - e.g. NCM_NAME="ncm"
	// This will be the name of the controller that is registered in the orchestrator controller API
	// This same name will be used as the key name for intents in the deployment intent group
	serviceName := os.Getenv(ENV_NCM_NAME)
	if serviceName == "" {
		serviceName = default_ncm_name
		log.Info("Using default name for NCM service name", log.Fields{
			"Name": serviceName,
		})
	}

	client := module.NewControllerClient()

	// Create or update the controller entry
	controller := module.Controller{
		Name: serviceName,
		Host: host,
		Port: strconv.Itoa(port),
	}
	_, err := client.CreateController(controller, true)
	if err != nil {
		log.Error("Failed to create/update a gRPC controller", log.Fields{
			"Error":      err,
			"Controller": serviceName,
		})
		return err
	}

	return nil
}
