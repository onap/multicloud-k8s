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

package status

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	yaml "github.com/ghodss/yaml"
	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/onap/multicloud-k8s/src/clm/pkg/cluster"
	v1alpha1 "github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"
	clientset "github.com/onap/multicloud-k8s/src/monitor/pkg/generated/clientset/versioned"
	informers "github.com/onap/multicloud-k8s/src/monitor/pkg/generated/informers/externalversions"
	appcontext "github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type channelManager struct {
	channels map[string]chan struct{}
	sync.Mutex
}

var channelData channelManager

const monitorLabel = "emco/deployment-id"

// HandleStatusUpdate for an application in a cluster
func HandleStatusUpdate(clusterId string, id string, v *v1alpha1.ResourceBundleState) {
	// Get the contextId from the label (id)
	result := strings.SplitN(id, "-", 2)
	if result[0] == "" {
		logrus.Info(clusterId, "::label is missing an appcontext identifier::", id)
		return
	}

	if len(result) != 2 {
		logrus.Info(clusterId, "::invalid label format::", id)
		return
	}

	// Get the app from the label (id)
	if result[1] == "" {
		logrus.Info(clusterId, "::label is missing an app identifier::", id)
		return
	}

	// Look up the contextId
	var ac appcontext.AppContext
	_, err := ac.LoadAppContext(result[0])
	if err != nil {
		logrus.Info(clusterId, "::App context not found::", result[0], "::Error::", err)
		return
	}

	// produce yaml representation of the status
	vjson, err := json.Marshal(v.Status)
	if err != nil {
		logrus.Info(clusterId, "::Error marshalling status information::", err)
		return
	}

	chandle, err := ac.GetClusterHandle(result[1], clusterId)
	if err != nil {
		logrus.Info(clusterId, "::Error getting cluster handle::", err)
		return
	}
	// Get the handle for the context/app/cluster status object
	handle, _ := ac.GetLevelHandle(chandle, "status")

	// If status handle was not found, then create the status object in the appcontext
	if handle == nil {
		ac.AddLevelValue(chandle, "status", string(vjson))
	} else {
		ac.UpdateStatusValue(handle, string(vjson))
	}

	return
}

// StartClusterWatcher watches for CR
// configBytes - Kubectl file data
func StartClusterWatcher(clusterId string) error {

	configBytes, err := getKubeConfig(clusterId)
	if err != nil {
		return err
	}

	//key := provider + "+" + name
	// Get the lock
	channelData.Lock()
	defer channelData.Unlock()
	// For first time
	if channelData.channels == nil {
		channelData.channels = make(map[string]chan struct{})
	}
	_, ok := channelData.channels[clusterId]
	if !ok {
		// Create Channel
		channelData.channels[clusterId] = make(chan struct{})
		// Create config
		config, err := clientcmd.RESTConfigFromKubeConfig(configBytes)
		if err != nil {
			logrus.Info(fmt.Sprintf("RESTConfigFromKubeConfig error: %s", err.Error()))
			return pkgerrors.Wrap(err, "RESTConfigFromKubeConfig error")
		}
		k8sClient, err := clientset.NewForConfig(config)
		if err != nil {
			return pkgerrors.Wrap(err, "Clientset NewForConfig error")
		}
		// Create Informer
		mInformerFactory := informers.NewSharedInformerFactory(k8sClient, 0)
		mInformer := mInformerFactory.K8splugin().V1alpha1().ResourceBundleStates().Informer()
		go scheduleStatus(clusterId, channelData.channels[clusterId], mInformer)
	}
	return nil
}

// StopClusterWatcher stop watching a cluster
func StopClusterWatcher(clusterId string) {
	//key := provider + "+" + name
	if channelData.channels != nil {
		c, ok := channelData.channels[clusterId]
		if ok {
			close(c)
		}
	}
}

// CloseAllClusterWatchers close all channels
func CloseAllClusterWatchers() {
	if channelData.channels == nil {
		return
	}
	// Close all Channels to stop all watchers
	for _, e := range channelData.channels {
		close(e)
	}
}

// Per Cluster Go routine to watch CR
func scheduleStatus(clusterId string, c <-chan struct{}, s cache.SharedIndexInformer) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			v, ok := obj.(*v1alpha1.ResourceBundleState)
			if ok {
				labels := v.GetLabels()
				l, ok := labels[monitorLabel]
				if ok {
					HandleStatusUpdate(clusterId, l, v)
				}
			}
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			v, ok := obj.(*v1alpha1.ResourceBundleState)
			if ok {
				labels := v.GetLabels()
				l, ok := labels[monitorLabel]
				if ok {
					HandleStatusUpdate(clusterId, l, v)
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			// Ignore it
		},
	}
	s.AddEventHandler(handlers)
	s.Run(c)
}

// getKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the clustername. This is written out to a file.
// TODO - consolidate with other rsync methods to get kubeconfig files
func getKubeConfig(clustername string) ([]byte, error) {

	if !strings.Contains(clustername, "+") {
		return nil, pkgerrors.New("Not a valid cluster name")
	}
	strs := strings.Split(clustername, "+")
	if len(strs) != 2 {
		return nil, pkgerrors.New("Not a valid cluster name")
	}
	kubeConfig, err := cluster.NewClusterClient().GetClusterContent(strs[0], strs[1])
	if err != nil {
		return nil, pkgerrors.New("Get kubeconfig failed")
	}

	dec, err := base64.StdEncoding.DecodeString(kubeConfig.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return dec, nil
}

// GetStatusCR returns a status monitoring customer resource
func GetStatusCR(label string) ([]byte, error) {

	var statusCr v1alpha1.ResourceBundleState

	statusCr.TypeMeta.APIVersion = "k8splugin.io/v1alpha1"
	statusCr.TypeMeta.Kind = "ResourceBundleState"
	statusCr.SetName(label)

	labels := make(map[string]string)
	labels["emco/deployment-id"] = label
	statusCr.SetLabels(labels)

	labelSelector, err := metav1.ParseToLabelSelector("emco/deployment-id = " + label)
	if err != nil {
		return nil, err
	}
	statusCr.Spec.Selector = labelSelector

	// Marshaling to json then convert to yaml works better than marshaling to yaml
	// The 'apiVersion' attribute was marshaling to 'apiversion'
	j, err := json.Marshal(&statusCr)
	if err != nil {
		return nil, err
	}
	y, err := yaml.JSONToYAML(j)
	if err != nil {
		return nil, err
	}

	return y, nil
}
