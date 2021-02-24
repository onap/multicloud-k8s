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
	"io"

	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/tiller/environment"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"

	pkgerrors "github.com/pkg/errors"
)

//implements environment.KubeClient but overrides it so that
//custom labels can be injected into created resources
// using internal k8sClient
type KubeClientImpl struct {
	environment.KubeClient
	Labels []string
	k      app.KubernetesClient
	//FIXME Add internal implementation for client
}

func newKubeClient(instanceId string, labels []string) (*KubeClientImpl, error) {
	//FIXME Implement init by your own
	//k8sClient := app.KubernetesClient{}
	//err := k8sClient.init(i.CloudRegion, id)
	var err error
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Initializing k8sClient")
	}
	return &KubeClientImpl{
		Labels:     labels,
		KubeClient: kube.New(nil),
	}, nil
}

//Create function is overrided to label test resources with custom labels
func (kci *KubeClientImpl) Create(namespace string, reader io.Reader, timeout int64, shouldWait bool) error {
	//FIXME
	return nil
}
