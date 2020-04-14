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
	contextDb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/contextdb"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
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

}
