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

package connector

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/onap/multicloud-k8s/src/clm/pkg/cluster"
	kubeclient "github.com/onap/multicloud-k8s/src/rsync/pkg/client"
	pkgerrors "github.com/pkg/errors"
)

type Connector struct {
	cid     string
	Clients map[string]*kubeclient.Client
	sync.Mutex
}

const basePath string = "/tmp/rsync/"

// Init connector for an app context
func Init(id interface{}) *Connector {
	c := make(map[string]*kubeclient.Client)
	str := fmt.Sprintf("%v", id)
	return &Connector{
		Clients: c,
		cid:     str,
	}
}

// getKubeConfig uses the connectivity client to get the kubeconfig based on the name
// of the clustername.
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

// GetClient returns client for the cluster
func (c *Connector) GetClient(cluster string) (*kubeclient.Client, error) {
	c.Lock()
	defer c.Unlock()

	client, ok := c.Clients[cluster]
	if !ok {
		// Get file from DB
		dec, err := getKubeConfig(cluster)
		if err != nil {
			return nil, err
		}
		var kubeConfigPath string = basePath + c.cid + "/" + cluster + "/"
		if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
			err = os.MkdirAll(kubeConfigPath, 0755)
			if err != nil {
				return nil, err
			}
		}
		kubeConfig := kubeConfigPath + "config"
		f, err := os.Create(kubeConfig)
		if err != nil {
			return nil, err
		}
		_, err = f.Write(dec)
		if err != nil {
			return nil, err
		}
		client = kubeclient.New("", kubeConfig, "default")
		if client != nil {
			c.Clients[cluster] = client
		}
	}
	return client, nil
}

func (c *Connector) GetClientWithRetry(cluster string) (*kubeclient.Client, error) {
	client, err := c.GetClient(cluster)
	if err != nil {
		return nil, err
	}
	if err = client.IsReachable(); err != nil {
		return nil, err // TODO: Add retry
	}
	return client, nil
}

func (c *Connector) RemoveClient() {
	c.Lock()
	defer c.Unlock()
	err := os.RemoveAll(basePath + "/" + c.cid)
	if err != nil {
		log.Printf("Warning: Deleting kubepath %s", err)
	}
}
