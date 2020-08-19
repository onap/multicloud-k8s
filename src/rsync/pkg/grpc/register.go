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
