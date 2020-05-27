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
	controller "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
	mtypes "github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/types"
)

const default_host = "localhost"
const default_port = 9031
const default_rsync_name = "rsync"
const ENV_RSYNC_NAME = "RSYNC_NAME"

func GetServerHostPort() (string, int) {

	// expect name of this rsync program to be in env variable "RSYNC_NAME" - e.g. RSYNC_NAME="rsync"
	serviceName := os.Getenv(ENV_RSYNC_NAME)
	if serviceName == "" {
		serviceName = default_rsync_name
		log.Info("Using default name for RSYNC service name", log.Fields{
			"Name": serviceName,
		})
	}

	// expect service name to be in env variable - e.g. RSYNC_SERVICE_HOST
	host := os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_HOST")
	if host == "" {
		host = default_host
		log.Info("Using default host for rsync gRPC controller", log.Fields{
			"Host": host,
		})
	}

	// expect service port to be in env variable - e.g. RSYNC_SERVICE_PORT
	port, err := strconv.Atoi(os.Getenv(strings.ToUpper(serviceName) + "_SERVICE_PORT"))
	if err != nil || port < 0 {
		port = default_port
		log.Info("Using default port for rsync gRPC controller", log.Fields{
			"Port": port,
		})
	}
	return host, port
}

func RegisterGrpcServer(host string, port int) error {
	// expect name of this rsync program to be in env variable "RSYNC_NAME" - e.g. RSYNC_NAME="rsync"
	// This will be the name of the controller that is registered in the orchestrator controller API
	// This same name will be used as the key name for intents in the deployment intent group
	serviceName := os.Getenv(ENV_RSYNC_NAME)
	if serviceName == "" {
		serviceName = default_rsync_name
		log.Info("Using default name for rsync service name", log.Fields{
			"Name": serviceName,
		})
	}

	client := controller.NewControllerClient()

	// Create or update the controller entry
	controller := controller.Controller{
		Metadata: mtypes.Metadata{
			Name: serviceName,
		},
		Spec: controller.ControllerSpec{
			Host:     host,
			Port:     port,
			Type:     controller.CONTROLLER_TYPE_ACTION,
			Priority: controller.MinControllerPriority,
		},
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
