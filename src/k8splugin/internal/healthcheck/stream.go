/*
Copyright © 2021 Samsung Electronics
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

package healthcheck

import (
	"os"

	"google.golang.org/grpc"
	"k8s.io/helm/pkg/proto/hapi/services"

	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"
	//pkgerrors "github.com/pkg/errors"
)

//implements basic stream implementation that logs progress
//and updates state in DB
type StreamImpl struct {
	*os.File          //destination of messages
	grpc.ServerStream //only to implement necessary methods
}

var _ services.ReleaseService_RunReleaseTestServer = StreamImpl{}

func NewStream() *StreamImpl {
	s := new(StreamImpl)
	s.File = os.Stdout
	return s
}

func (si StreamImpl) Send(m *services.TestReleaseResponse) error {
	log.Info("Stream message", log.Fields{"msg": m})
	//TODO dump to DB
	return nil
}

// Unfinished FIXME TODO
