/*
 * Copyright 2020 Intel Corporation, Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongoclient

import (
	"log"

	"github.com/onap/multicloud-k8s/src/orchestrator/internal/config"
	"github.com/onap/multicloud-k8s/src/orchestrator/internal/db"
)

// MongoClient implements the db.Store interface
// It will also be used to maintain some localized state
type MongoClient struct {
	mongodb *db.MongoStore
	conf    *conf.Configuration
}

// Config is the structure that stores the configuration
var mongoClient *MongoClient

func createMongoClient() error {
	conf, err := config.readConfigFile("config.json")
	if err != nil {
		log.Println("Error loading config file: ", err)
		log.Println("Using defaults...")
	}
	mongoClient.conf = conf
	err := db.InitializeDatabaseConnection()
	if err != nil {
		log.Println("Unable to initialize database connection...")
		log.Println(err)
		log.Fatalln("Exiting...")
		return
	}
	mongoClient.mongodb = db.DBconn
}
