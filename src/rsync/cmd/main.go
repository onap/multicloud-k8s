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

package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"
	//"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/auth"
	//"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
	contextDb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
/*
	gplugin "rsync/pkg/resource"
	plugin "rsync/pkg/plugin"
	"rsync/pkg/app"
*/
	"rsync/pkg/context"
)

func main() {

	rand.Seed(time.Now().UnixNano())

	// Initialize the mongodb
	err := db.InitializeDatabaseConnection("mco")
	if err != nil {
		fmt.Println(" Exiting mongod ")
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	// Initialize contextdb
	err = contextDb.InitializeContextDatabase()
	if err != nil {
		fmt.Println(" Exiting etcd")
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
	}

	// Initialize grpc

        instca := context.CompositeAppContext{}

        err = instca.InstantiateComApp("7871147865598089755")
        if err != nil {
                fmt.Printf("\n instantiation failed \n")
        }
/*
	k8sClient := app.KubernetesClient{}
	err = k8sClient.Init("testcluster", "12345")
	if err != nil {
		fmt.Printf(" Init failed = %s", err.Error())
		//return InstanceResponse{}, pkgerrors.Wrap(err, "Getting CloudRegion Information")
	}
*/
/*
        _, err = k8sClient.CreateResources(nil, "default")
        if err != nil {
		fmt.Printf(" Create resource failed = %s", err.Error())
                //return InstanceResponse{}, pkgerrors.Wrap(err, "Create Kubernetes Resources")
        }
*/
/*

	var c plugin.KubernetesConnector
	c = &k8sClient
	var gp gplugin.GenericPlugin
	_, err = gp.Create("/vagrant/yaml/example.yml","default", c)
	if err != nil {
		fmt.Printf(" Create failed = %s", err.Error())
	}
	fmt.Println("Initialization done")
*/

}
