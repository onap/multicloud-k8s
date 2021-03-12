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
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/tiller/environment"

	"github.com/onap/multicloud-k8s/src/k8splugin/internal/app"
	"github.com/onap/multicloud-k8s/src/k8splugin/internal/config"

	pkgerrors "github.com/pkg/errors"
)

//implements environment.KubeClient but overrides it so that
//custom labels can be injected into created resources
//using internal k8sClient
type KubeClientImpl struct {
	environment.KubeClient
	labels map[string]string
	k      app.KubernetesClient
}

var _ environment.KubeClient = KubeClientImpl{}

func NewKubeClient(instanceId, cloudRegion string) (*KubeClientImpl, error) {
	k8sClient := app.KubernetesClient{}
	err := k8sClient.Init(cloudRegion, instanceId)
	if err != nil {
		return nil, pkgerrors.Wrap(err, "Initializing k8sClient")
	}
	return &KubeClientImpl{
		labels: map[string]string{
			config.GetConfiguration().KubernetesLabelName: instanceId,
		},
		KubeClient: kube.New(&k8sClient),
		k:          k8sClient,
	}, nil
}

/* FIXME
// Need to correct this later and provide override of Create method to use our k8sClient
// So that healthcheck hook resources would be labeled with vf-module data just like currently
// every k8splugin-managed resource is

//Create function is overrided to label test resources with custom labels
func (kci *KubeClientImpl) Create(namespace string, reader io.Reader, timeout int64, shouldWait bool) error {
	return nil
}
*/
