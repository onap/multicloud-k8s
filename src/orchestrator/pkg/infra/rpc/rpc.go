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

package rpc

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/config"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/testdata"
)

type ContextUpdateRequest interface {
}

type ContextUpdateResponse interface {
}

type InstallAppRequest interface {
}

type InstallAppResponse interface {
}

type rpcInfo struct {
	conn *grpc.ClientConn
	host string
	port int
}

var mutex = &sync.Mutex{}
var rpcConnections = make(map[string]rpcInfo)

// GetRpcConn is used by RPC client code which needs the connection for a
// given controller for doing RPC calls with that controller.
func GetRpcConn(name string) *grpc.ClientConn {
	mutex.Lock()
	defer mutex.Unlock()
	if val, ok := rpcConnections[name]; ok {
		if val.conn.GetState() == connectivity.TransientFailure {
			val.conn.ResetConnectBackoff()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			if !val.conn.WaitForStateChange(ctx, connectivity.TransientFailure) {
				log.Warn("Error re-establishing RPC connection", log.Fields{
					"Server": name,
					"Host":   val.host,
					"Port":   val.port,
				})
			}
			cancel()
		}
		return val.conn
	}
	return nil
}

func UpdateRpcConn(name, host string, port int) {
	mutex.Lock()
	defer mutex.Unlock()
	if val, ok := rpcConnections[name]; ok {
		// close connection if mismatch in update vs cached connect info
		if val.host != host || val.port != port {
			log.Info("Closing RPC connection due to mismatch", log.Fields{
				"Server":   name,
				"Old Host": val.host,
				"Old Port": val.port,
				"New Host": host,
				"New Port": port,
			})
			err := val.conn.Close()
			if err != nil {
				log.Warn("Error closing RPC connection", log.Fields{
					"Server": name,
					"Host":   val.host,
					"Port":   val.port,
					"Error":  err,
				})
			}
		} else {
			if val.conn.GetState() == connectivity.TransientFailure {
				val.conn.ResetConnectBackoff()
			}
			return
		}
	}
	// connect and update rpcConnection list - for new or modified connection
	conn, err := createClientConn(host, port)
	if err != nil {
		log.Warn("Failed to create RPC Client connection", log.Fields{
			"Error": err,
		})
		delete(rpcConnections, name)
	} else {
		log.Info("Added RPC Client connection", log.Fields{
			"Controller": name,
		})
		rpcConnections[name] = rpcInfo{
			conn: conn,
			host: host,
			port: port,
		}
	}
}

// CloseAllRpcConn closes all connections
func CloseAllRpcConn() {
	mutex.Lock()
	for k, v := range rpcConnections {
		err := v.conn.Close()
		if err != nil {
			log.Warn("Error closing RPC connection", log.Fields{
				"Server": k,
				"Host":   v.host,
				"Port":   v.port,
				"Error":  err,
			})
		}
	}
	mutex.Unlock()
}

// RemoveRpcConn closes the connection and removes from map
func RemoveRpcConn(name string) {
	mutex.Lock()
	if val, ok := rpcConnections[name]; ok {
		err := val.conn.Close()
		if err != nil {
			log.Warn("Error closing RPC connection", log.Fields{
				"Server": name,
				"Host":   val.host,
				"Port":   val.port,
				"Error":  err,
			})
		}
		delete(rpcConnections, name)
	}
	mutex.Unlock()
}

// createConn creates the Rpc Client Connection
func createClientConn(Host string, Port int) (*grpc.ClientConn, error) {
	var err error
	var tls bool
	var opts []grpc.DialOption

	serverAddr := Host + ":" + strconv.Itoa(Port)
	serverNameOverride := config.GetConfiguration().GrpcServerNameOverride

	if strings.Contains(config.GetConfiguration().GrpcEnableTLS, "enable") {
		tls = true
	} else {
		tls = false
	}

	caFile := config.GetConfiguration().GrpcCAFile

	if tls {
		if caFile == "" {
			caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(caFile, serverNameOverride)
		if err != nil {
			log.Error("Failed to create TLS credentials", log.Fields{
				"Error": err,
				"Host":  Host,
				"Port":  Port,
			})
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		pkgerrors.Wrap(err, "Grpc Client Initialization failed with error")
	}

	return conn, err
}
