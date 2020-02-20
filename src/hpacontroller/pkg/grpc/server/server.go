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

package server

import (
	"context"
	"flag"

	pb "github.com/onap/multicloud-k8s/src/hpacontroller/pkg/grpc/controller"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	tls        = flag.Bool("tls", false, "TLS Connection if true, else plain TCP")
	certFile   = flag.String("cert_file", "", "The TLS cert file")
	keyFile    = flag.String("key_file", "", "The TLS key file")
	jsonDBFile = flag.String("json_db_file", "", "A json file containing a list of features")
	port       = flag.Int("port", 10000, "The server port")
)

type controllerServer struct {
	pb.UnimplementedControllerServer
}

func (cs *controllerServer) UpdateAppContext(ctx context.Context, req *pb.ContextUpdateRequest) (*pb.ContextUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateAppContext not implemented")
}
