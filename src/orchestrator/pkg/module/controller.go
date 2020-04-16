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
	"time"

	healthpb "github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/healthcheck"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/db"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	rpc "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/rpc"
	pkgerrors "github.com/pkg/errors"
)

// Controller contains the parameters needed for Controllers
// It implements the interface for managing the Controllers
type Controller struct {
	Metadata Metadata       `json:"metadata"`
	Spec     ControllerSpec `json:"spec"`
}

type ControllerSpec struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Type     string `json:"type"`
	Priority int    `json:"priority"`
}

const MinControllerPriority = 1
const MaxControllerPriority = 1000000
const CONTROLLER_TYPE_ACTION string = "action"
const CONTROLLER_TYPE_PLACEMENT string = "placement"

var CONTROLLER_TYPES = [...]string{CONTROLLER_TYPE_ACTION, CONTROLLER_TYPE_PLACEMENT}

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
	CreateController(ms Controller, mayExist bool) (Controller, error)
	GetController(name string) (Controller, error)
	GetControllers() ([]Controller, error)
	InitControllers()
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
func (mc *ControllerClient) CreateController(m Controller, mayExist bool) (Controller, error) {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: m.Metadata.Name,
	}

	//Check if this Controller already exists
	_, err := mc.GetController(m.Metadata.Name)
	if err == nil && !mayExist {
		return Controller{}, pkgerrors.New("Controller already exists")
	}

	err = db.DBconn.Insert(mc.collectionName, key, nil, mc.tagMeta, m)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	err = mc.HealthCheck(m.Metadata.Name)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "HealthCheck Failed")
	}

	return m, nil
}

// HealthCheck performs a gRPC healthcheck of the provided controller
func (mc *ControllerClient) HealthCheck(controllerName string) error {
	var err error
	var rpcClient healthpb.HealthcheckClient
	var healthRes *healthpb.HealthCheckResponse
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := mc.GetController(controllerName)
	if err != nil {
		return pkgerrors.Wrapf(err, "HealthCheck - error find controller in db: %v", controllerName)
	}
	getConn := rpc.GetRpcConnReq(controllerName, c.Spec.Host, c.Spec.Port)
	rpc.RpcCtl <- getConn
	conn := <-getConn.RespChan

	if conn != nil {
		rpcClient = healthpb.NewHealthcheckClient(conn)
		healthReq := new(healthpb.HealthCheckRequest)
		healthRes, err = rpcClient.HealthCheck(ctx, healthReq)
	} else {
		return pkgerrors.Errorf("HealthCheck Failed - Could not get HealthcheckClient: %v", controllerName)
	}

	if err != nil {
		return pkgerrors.Wrapf(err, "HealthCheck Failed: %v", controllerName)
	}
	log.Info("HealthCheck Passed", log.Fields{
		"Controller": controllerName,
		"Result":     healthRes,
	})
	return err
}

// GetController returns the Controller for corresponding name
func (mc *ControllerClient) GetController(name string) (Controller, error) {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: name,
	}
	value, err := db.DBconn.Find(mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return Controller{}, pkgerrors.Wrap(err, "Get Controller")
	}

	if value != nil {
		microserv := Controller{}
		err = db.DBconn.Unmarshal(value[0], &microserv)
		if err != nil {
			return Controller{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}
		return microserv, nil
	}

	return Controller{}, pkgerrors.New("Error getting Controller")
}

// GetControllers returns all the  Controllers that are registered
func (mc *ControllerClient) GetControllers() ([]Controller, error) {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: "",
	}

	var resp []Controller
	values, err := db.DBconn.Find(mc.collectionName, key, mc.tagMeta)
	if err != nil {
		return []Controller{}, pkgerrors.Wrap(err, "Get Controller")
	}

	for _, value := range values {
		microserv := Controller{}
		err = db.DBconn.Unmarshal(value, &microserv)
		if err != nil {
			return []Controller{}, pkgerrors.Wrap(err, "Unmarshaling Value")
		}

		// run healthcheck
		/*
			err = mc.HealthCheck(microserv.Name)
			if err != nil {
				log.Warn("HealthCheck Failed", log.Fields{
					"Controller": microserv.Name,
				})
			}
		*/
		resp = append(resp, microserv)
	}

	return resp, nil
}

// InitControllers initializes connctions for controllers in the DB
func (mc *ControllerClient) InitControllers() {
	vals, _ := mc.GetControllers()
	for _, v := range vals {
		log.Info("Initializing RPC connection for controller", log.Fields{
			"Controller": v.Metadata.Name,
		})
		getConn := rpc.GetRpcConnReq(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
		rpc.RpcCtl <- getConn
		<-getConn.RespChan
	}
}

// DeleteController the  Controller from database
func (mc *ControllerClient) DeleteController(name string) error {

	//Construct the composite key to select the entry
	key := ControllerKey{
		ControllerName: name,
	}
	err := db.DBconn.Remove(mc.collectionName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Controller Entry;")
	}

	// send message to close rpc connection
	rmConn := rpc.GetRpcConnReq(name, "", 0)
	rpc.RpcCtl <- rmConn
	<-rmConn.RespChan

	return nil
}
