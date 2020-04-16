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
	"fmt"
	"strconv"
	"strings"
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

var RpcCtl = make(chan RpcConnReq)

type RpcConnReq struct {
	Name     string
	Host     string
	Port     int
	RespChan chan *grpc.ClientConn
}

func GetRpcConnReq(name, host string, port int) RpcConnReq {
	return RpcConnReq{
		Name:     name,
		Host:     host,
		Port:     port,
		RespChan: make(chan *grpc.ClientConn),
	}
}

func HandleRpcConnections() {
	type rpcInfo struct {
		conn *grpc.ClientConn
		host string
		port int
	}
	rpcConnections := make(map[string]rpcInfo)

	for {
		select {
		case r := <-RpcCtl:
			// Handle shutdown request
			if r.Name == "" {
				fmt.Println("Shutting done RPC connection manager")
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
				r.RespChan <- nil
				break
			}

			// If Host string is null, then connection for given controller is being removed.
			// Close and remove connection if present in the rpc connection list.
			// Send nil connection on the response channel
			if r.Host == "" {
				fmt.Printf("Closing RPC connection %v", r.Name)
				if val, ok := rpcConnections[r.Name]; ok {
					err := val.conn.Close()
					if err != nil {
						log.Warn("Error closing RPC connection", log.Fields{
							"Server": r.Name,
							"Host":   val.host,
							"Port":   val.port,
							"Error":  err,
						})
					}
					delete(rpcConnections, r.Name)
				}
				r.RespChan <- nil
				continue
			}

			if val, ok := rpcConnections[r.Name]; ok {
				// close connection if mismatch in db config vs last connection
				if val.host != r.Host || val.port != r.Port {
					log.Info("Closing RPC connection due to mismatch", log.Fields{
						"Server":   r.Name,
						"Old Host": val.host,
						"Old Port": val.port,
						"New Host": r.Host,
						"New Port": r.Port,
					})
					err := val.conn.Close()
					if err != nil {
						log.Warn("Error closing RPC connection", log.Fields{
							"Server": r.Name,
							"Host":   val.host,
							"Port":   val.port,
							"Error":  err,
						})
					}
				} else {
					if val.conn.GetState() == connectivity.TransientFailure {
						val.conn.ResetConnectBackoff()
						ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
						if !val.conn.WaitForStateChange(ctx, connectivity.TransientFailure) {
							log.Warn("Error re-establishing RPC connection", log.Fields{
								"Server": r.Name,
								"Host":   val.host,
								"Port":   val.port,
							})
						}
						cancel()
					}
					r.RespChan <- val.conn
					continue
				}
			}
			// connect and update rpcConnection list
			conn, err := createClientConn(r.Host, r.Port)
			if err != nil {
				log.Error("Failed to create RPC Client connection", log.Fields{
					"Error": err,
				})
				delete(rpcConnections, r.Name)
				r.RespChan <- nil
			} else {
				rpcConnections[r.Name] = rpcInfo{
					conn: conn,
					host: r.Host,
					port: r.Port,
				}
				r.RespChan <- conn
			}
		}
	}
}

// createConn creates the Rpc Client Connection
func createClientConn(Host string, Port int) (*grpc.ClientConn, error) {
	var err error
	var tls bool
	var opts []grpc.DialOption

	serverAddr := Host + ":" + strconv.Itoa(Port)
	serverHostVerify := config.GetConfiguration().GrpcServerHostVerify

	if strings.Contains(config.GetConfiguration().GrpcTLS, "enable") {
		tls = true
	} else {
		tls = false
	}

	caFile := config.GetConfiguration().GrpcCAFile

	if tls {
		if caFile == "" {
			caFile = testdata.Path("ca.pem")
		}
		creds, err := credentials.NewClientTLSFromFile(caFile, serverHostVerify)
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
