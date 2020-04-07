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

package module

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	controllerpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/controller"
	rpc "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	pkgerrors "github.com/pkg/errors"
)

// Controller contains the parameters needed for Controllers
// It implements the interface for managing the Controllers
type Controller struct {
	Name string `json:"name"`

	Host string `json:"host"`

	Port string `json:"port"`
}

// ControllerKey is the key structure that is used in the database
type ControllerKey struct {
	ControllerName string `json:"controller-name"`
}

// We will use json marshalling to convert to string to
// preserve the underlying structure.
func (mk ControllerKey) String() string {
	out, err := json.Marshal(mk)
	if err != nil {
		return ""
	}

	return string(out)
}

// ControllerManager is an interface exposes the Controller functionality
type ControllerManager interface {
	CreateController(ms Controller) (Controller, error)
	GetController(name string) (Controller, error)
	DeleteController(name string) error
	HealthCheck(controllerName string) error
}

// ControllerClient implements the Manager
// It will also be used to maintain some localized state
type ControllerClient struct {
	collectionName string
	tagMeta        string
}

// NewControllerClient returns an instance of the ControllerClient
// which implements the Manager
func NewControllerClient() *ControllerClient {
	return &ControllerClient{
		collectionName: "controller",
		tagMeta:        "controllermetadata",
	}
}

// CreateController a new collection based on the Controller
func (mc *ControllerClient) CreateController(m Controller) (Controller, error) {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: m.Name,
	}

	//Check if this Controller already exists
	_, err := mc.GetController(m.Name)
	if err == nil {
		return Controller{}, pkgerrors.New("Controller already exists")
	}

	err = db.DBconn.Create(mc.collectionName, key, mc.tagMeta, m)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	err = rpc.InitializeRPC(m.Host, m.Port, m.Name)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "Initilize RPC Failed")
	}

	err = mc.HealthCheck(m.Name)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "HealthCheck Failed")
	}

	return m, nil
}

// HealthCheck performs a gRPC healthcheck of the provided controller
func (mc *ControllerClient) HealthCheck(controllerName string)(error) {
	var err error
	var rpcClient controllerpb.ControllerClient 
	var healthRes *controllerpb.HealthCheckResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if rpc.RPC[controllerName] != nil {
		rpcClient = rpc.RPC[controllerName]
		healthReq := new(controllerpb.HealthCheckRequest)
		healthRes, err = rpcClient.HealthCheck(ctx, healthReq)
	} else {
		return pkgerrors.Wrap(err, "HealthCheck Failed - Could not get ControllerClient")
	}
	
	if err != nil {
		return pkgerrors.Wrap(err, "HealthCheck Failed")
	}
	log.Println("HealthCheck Passed: ")
	log.Printf("%+v\n", healthRes)
	return err
}

// GetController returns the Controller for corresponding name
func (mc *ControllerClient) GetController(name string) (Controller, error) {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: name,
	}
	value, err := db.DBconn.Read(mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "Get Controller")
	}

	//value is a byte array
	if value != nil {
		microserv := Controller{}
		err = db.DBconn.Unmarshal(value, &microserv)
		if err != nil {
			return Controller{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return microserv, nil
	}

	return Controller{}, pkgerrors.New("Error getting Controller")
}

// DeleteController the  Controller from database
func (mc *ControllerClient) DeleteController(name string) error {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: name,
	}
	err := db.DBconn.Delete(name, key, mc.tagMeta)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Controller Entry;")
	}
	return nil
}
