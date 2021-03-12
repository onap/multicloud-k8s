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
	"google.golang.org/grpc"
	"k8s.io/helm/pkg/proto/hapi/release"
	"k8s.io/helm/pkg/proto/hapi/services"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/db"
	log "github.com/onap/multicloud-k8s/src/k8splugin/internal/logutils"

	pkgerrors "github.com/pkg/errors"
)

//implements basic stream implementation that logs progress
//and updates state in DB
type StreamImpl struct {
	Key               HealthcheckKey
	StoreName         string
	Tag               string
	grpc.ServerStream //only to satisfy interface needs, it's not used
}

var _ services.ReleaseService_RunReleaseTestServer = StreamImpl{}

func NewStream(key HealthcheckKey, store, tag string) *StreamImpl {
	s := StreamImpl{
		Key:       key,
		StoreName: store,
		Tag:       tag,
	}
	return &s
}

func (si StreamImpl) Send(m *services.TestReleaseResponse) error {
	log.Info("Stream message", log.Fields{"msg": m})

	DBResp, err := db.DBconn.Read(si.StoreName, si.Key, si.Tag)
	if err != nil || DBResp == nil {
		return pkgerrors.Wrap(err, "Error retrieving Healthcheck data")
	}

	resp := InstanceHCStatus{}
	err = db.DBconn.Unmarshal(DBResp, &resp)
	if err != nil {
		return pkgerrors.Wrap(err, "Unmarshaling Healthcheck Value")
	}

	resp.Status = release.TestRun_Status_name[int32(m.Status)]
	if m.Msg != "" {
		if resp.Info == "" {
			resp.Info = m.Msg
		} else {
			resp.Info = resp.Info + "\n" + m.Msg
		}
	}

	err = db.DBconn.Update(si.StoreName, si.Key, si.Tag, resp)
	if err != nil {
		return pkgerrors.Wrap(err, "Updating Healthcheck")
	}
	return nil
}
