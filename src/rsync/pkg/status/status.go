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
	"encoding/json"
	"fmt"
	"sync"

	pkgerrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	v1alpha1 "github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"
	clientset "github.com/onap/multicloud-k8s/src/monitor/pkg/generated/clientset/versioned"
	informers "github.com/onap/multicloud-k8s/src/monitor/pkg/generated/informers/externalversions"
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
// TODO: Add code for specific handling
func HandleStatusUpdate(provider, name string, id string, v *v1alpha1.ResourceBundleState) error {
	logrus.Info("label::", id)
	//status := v.Status.ServiceStatuses
	//podStatus := v.Status.PodStatuses
	// Store Pod Status in app context
	out, _ := json.Marshal(v.Status)
	logrus.Info("Status::", string(out))
	return nil
}

// StartClusterWatcher watches for CR
// configBytes - Kubectl file data
func StartClusterWatcher(provider, name string, configBytes []byte) error {
	key := provider + "+" + name
	// Get the lock
	channelData.Lock()
	defer channelData.Unlock()
	// For first time
	if channelData.channels == nil {
		channelData.channels = make(map[string]chan struct{})
	}
	_, ok := channelData.channels[key]
	if !ok {
		// Create Channel
		channelData.channels[key] = make(chan struct{})
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
		go scheduleStatus(provider, name, channelData.channels[key], mInformer)
	}
	return nil
}

// StopClusterWatcher stop watching a cluster
func StopClusterWatcher(provider, name string) {
	key := provider + "+" + name
	if channelData.channels != nil {
		c, ok := channelData.channels[key]
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
func scheduleStatus(provider, name string, c <-chan struct{}, s cache.SharedIndexInformer) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			v, ok := obj.(*v1alpha1.ResourceBundleState)
			if ok {
				labels := v.GetLabels()
				l, ok := labels[monitorLabel]
				if ok {
					HandleStatusUpdate(provider, name, l, v)
				}
			}
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			v, ok := obj.(*v1alpha1.ResourceBundleState)
			if ok {
				labels := v.GetLabels()
				l, ok := labels[monitorLabel]
				if ok {
					HandleStatusUpdate(provider, name, l, v)
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
