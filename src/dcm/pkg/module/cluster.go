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

package module

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	clm "github.com/onap/multicloud-k8s/src/clm/pkg/cluster"
	rb "github.com/onap/multicloud-k8s/src/monitor/pkg/apis/k8splugin/v1alpha1"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// Cluster contains the parameters needed for a Cluster
type Cluster struct {
	MetaData      ClusterMeta `json:"metadata"`
	Specification ClusterSpec `json:"spec"`
}

type ClusterMeta struct {
	ClusterReference string `json:"name"`
	Description      string `json:"description"`
	UserData1        string `json:"userData1"`
	UserData2        string `json:"userData2"`
}

type ClusterSpec struct {
	ClusterProvider string `json:"cluster-provider"`
	ClusterName     string `json:"cluster-name"`
	LoadBalancerIP  string `json:"loadbalancer-ip"`
	Certificate     string `json:"certificate"`
}

type ClusterKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logical-cloud-name"`
	ClusterReference string `json:"clname"`
}

type KubeConfig struct {
	ApiVersion     string            `yaml:"apiVersion"`
	Kind           string            `yaml:"kind"`
	Clusters       []KubeCluster     `yaml:"clusters"`
	Contexts       []KubeContext     `yaml:"contexts"`
	CurrentContext string            `yaml:"current-context"`
	Preferences    map[string]string `yaml:"preferences"`
	Users          []KubeUser        `yaml:"users"`
}

type KubeCluster struct {
	ClusterDef  KubeClusterDef `yaml:"cluster"`
	ClusterName string         `yaml:"name"`
}

type KubeClusterDef struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type KubeContext struct {
	ContextDef  KubeContextDef `yaml:"context"`
	ContextName string         `yaml:"name"`
}

type KubeContextDef struct {
	Cluster   string `yaml:"cluster"`
	Namespace string `yaml:"namespace,omitempty"`
	User      string `yaml:"user"`
}

type KubeUser struct {
	UserName string      `yaml:"name"`
	UserDef  KubeUserDef `yaml:"user"`
}

type KubeUserDef struct {
	ClientCertificateData string `yaml:"client-certificate-data"`
	ClientKeyData         string `yaml:"client-key-data"`
	// client-certificate and client-key are NOT implemented
}

// ClusterManager is an interface that exposes the connection
// functionality
type ClusterManager interface {
	CreateCluster(project, logicalCloud string, c Cluster) (Cluster, error)
	GetCluster(project, logicalCloud, name string) (Cluster, error)
	GetAllClusters(project, logicalCloud string) ([]Cluster, error)
	DeleteCluster(project, logicalCloud, name string) error
	UpdateCluster(project, logicalCloud, name string, c Cluster) (Cluster, error)
	GetClusterConfig(project, logicalcloud, name string) (string, error)
}

// ClusterClient implements the ClusterManager
// It will also be used to maintain some localized state
type ClusterClient struct {
	storeName string
	tagMeta   string
	util      Utility
}

// ClusterClient returns an instance of the ClusterClient
// which implements the ClusterManager
func NewClusterClient() *ClusterClient {
	service := DBService{}
	return &ClusterClient{
		storeName: "orchestrator",
		tagMeta:   "cluster",
		util:      service,
	}
}

// Create entry for the cluster reference resource in the database
func (v *ClusterClient) CreateCluster(project, logicalCloud string, c Cluster) (Cluster, error) {

	//Construct key consisting of name
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: c.MetaData.ClusterReference,
	}

	//Check if project exists
	err := v.util.CheckProject(project)
	if err != nil {
		return Cluster{}, pkgerrors.New("Unable to find the project")
	}
	//check if logical cloud exists
	err = v.util.CheckLogicalCloud(project, logicalCloud)
	if err != nil {
		return Cluster{}, pkgerrors.New("Unable to find the logical cloud")
	}
	//Check if this Cluster reference already exists
	_, err = v.GetCluster(project, logicalCloud, c.MetaData.ClusterReference)
	if err == nil {
		return Cluster{}, pkgerrors.New("Cluster reference already exists")
	}

	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Creating DB Entry")
	}

	return c, nil
}

// Get returns  Cluster for corresponding cluster reference
func (v *ClusterClient) GetCluster(project, logicalCloud, clusterReference string) (Cluster, error) {

	//Construct the composite key to select the entry
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	value, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Get Cluster reference")
	}

	//value is a byte array
	if value != nil {
		cl := Cluster{}
		err = v.util.DBUnmarshal(value[0], &cl)
		if err != nil {
			return Cluster{}, pkgerrors.Wrap(err, "Unmarshaling value")
		}
		return cl, nil
	}

	return Cluster{}, pkgerrors.New("Cluster Reference does not exist")
}

// GetAll returns all cluster references in the logical cloud
func (v *ClusterClient) GetAllClusters(project, logicalCloud string) ([]Cluster, error) {
	//Construct the composite key to select clusters
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: "",
	}
	var resp []Cluster
	values, err := v.util.DBFind(v.storeName, key, v.tagMeta)
	if err != nil {
		return []Cluster{}, pkgerrors.Wrap(err, "Get All Cluster references")
	}
	if len(values) == 0 {
		return []Cluster{}, pkgerrors.New("No Cluster References associated")
	}

	for _, value := range values {
		cl := Cluster{}
		err = v.util.DBUnmarshal(value, &cl)
		if err != nil {
			return []Cluster{}, pkgerrors.Wrap(err, "Unmarshaling values")
		}
		resp = append(resp, cl)
	}

	return resp, nil
}

// Delete the Cluster reference entry from database
func (v *ClusterClient) DeleteCluster(project, logicalCloud, clusterReference string) error {
	//Construct the composite key to select the entry
	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}
	err := v.util.DBRemove(v.storeName, key)
	if err != nil {
		return pkgerrors.Wrap(err, "Delete Cluster Reference")
	}
	return nil
}

// Update an entry for the Cluster reference in the database
func (v *ClusterClient) UpdateCluster(project, logicalCloud, clusterReference string, c Cluster) (Cluster, error) {

	key := ClusterKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
		ClusterReference: clusterReference,
	}

	//Check for name mismatch in cluster reference
	if c.MetaData.ClusterReference != clusterReference {
		return Cluster{}, pkgerrors.New("Cluster Reference mismatch")
	}
	//Check if this Cluster reference exists
	_, err := v.GetCluster(project, logicalCloud, clusterReference)
	if err != nil {
		return Cluster{}, pkgerrors.New("Cluster Reference does not exist")
	}
	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}

// Get returns Cluster's kubeconfig for corresponding cluster reference
func (v *ClusterClient) GetClusterConfig(project, logicalCloud, clusterReference string) (string, error) {
	lcClient := NewLogicalCloudClient()
	lckey := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
	}
	context, ctxVal, err := lcClient.util.GetLogicalCloudContext(lcClient.storeName, lckey, lcClient.tagMeta, project, logicalCloud)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Error getting logical cloud context.")
	}
	if ctxVal == "" {
		return "", pkgerrors.New("Logical Cloud hasn't been applied yet")
	}

	// get logical cloud resource
	lc, err := lcClient.Get(project, logicalCloud)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed getting logical cloud")
	}
	// get user's private key
	privateKeyData, err := v.util.DBFind(v.storeName, lckey, "privatekey")
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed getting private key from logical cloud")
	}

	// get cluster from dcm (need provider/name)
	cluster, err := v.GetCluster(project, logicalCloud, clusterReference)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed getting cluster")
	}

	// before attempting to generate a kubeconfig,
	// check if certificate has been issued and copy it from etcd to mongodb
	if cluster.Specification.Certificate == "" {
		log.Info("Certificate not yet in MongoDB, checking etcd.", log.Fields{})

		// access etcd
		clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")

		// get the app context handle for the status of this cluster (which should contain the certificate inside, if already issued)
		statusHandle, err := context.GetClusterStatusHandle("logical-cloud", clusterName)

		if err != nil {
			return "", pkgerrors.New("The cluster doesn't contain status, please check if all services are up and running.")
		}
		statusRaw, err := context.GetValue(statusHandle)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while reading the cluster status.")
		}

		var rbstatus rb.ResourceBundleStatus
		err = json.Unmarshal([]byte(statusRaw.(string)), &rbstatus)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while parsing the cluster status.")
		}

		if len(rbstatus.CsrStatuses) == 0 {
			return "", pkgerrors.New("The certificate for this cluster hasn't been issued yet. Please try later.")
		}

		// validate that we indeed obtained a certificate before persisting it in the database:
		approved := false
		for _, c := range rbstatus.CsrStatuses[0].Status.Conditions {
			if c.Type == "Denied" {
				return "", pkgerrors.Wrap(err, "Certificate was denied!")
			}
			if c.Type == "Failed" {
				return "", pkgerrors.Wrap(err, "Certificate issue failed.")
			}
			if c.Type == "Approved" {
				approved = true
			}
		}
		if approved {
			//just double-check certificate field contents aren't empty:
			cert := rbstatus.CsrStatuses[0].Status.Certificate
			if len(cert) > 0 {
				cluster.Specification.Certificate = base64.StdEncoding.EncodeToString([]byte(cert))
			} else {
				return "", pkgerrors.Wrap(err, "Certificate issued was invalid.")
			}
		}

		// copy key to MongoDB
		// func (v *ClusterClient)
		// UpdateCluster(project, logicalCloud, clusterReference string, c Cluster) (Cluster, error) {
		_, err = v.UpdateCluster(project, logicalCloud, clusterReference, cluster)
		if err != nil {
			return "", pkgerrors.Wrap(err, "An error occurred while storing the certificate.")
		}
	} else {
		// certificate is already in MongoDB so just hand it over to create the API response
		log.Info("Certificate already in MongoDB, pass it to API.", log.Fields{})
	}

	// contact clm about admins cluster kubeconfig (to retrieve CA cert)
	clusterContent, err := clm.NewClusterClient().GetClusterContent(cluster.Specification.ClusterProvider, cluster.Specification.ClusterName)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed getting cluster content from CLM")
	}
	adminConfig, err := base64.StdEncoding.DecodeString(clusterContent.Kubeconfig)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed decoding CLM kubeconfig from base64")
	}

	// unmarshall clm kubeconfig into struct
	adminKubeConfig := KubeConfig{}
	err = yaml.Unmarshal(adminConfig, &adminKubeConfig)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed parsing CLM kubeconfig yaml")
	}

	// all data needed for final kubeconfig:
	privateKey := string(privateKeyData[0])
	signedCert := cluster.Specification.Certificate
	clusterCert := adminKubeConfig.Clusters[0].ClusterDef.CertificateAuthorityData
	clusterAddr := adminKubeConfig.Clusters[0].ClusterDef.Server
	namespace := lc.Specification.NameSpace
	userName := lc.Specification.User.UserName
	contextName := userName + "@" + clusterReference

	kubeconfig := KubeConfig{
		ApiVersion: "v1",
		Kind:       "Config",
		Clusters: []KubeCluster{
			KubeCluster{
				ClusterName: clusterReference,
				ClusterDef: KubeClusterDef{
					CertificateAuthorityData: clusterCert,
					Server:                   clusterAddr,
				},
			},
		},
		Contexts: []KubeContext{
			KubeContext{
				ContextName: contextName,
				ContextDef: KubeContextDef{
					Cluster:   clusterReference,
					Namespace: namespace,
					User:      userName,
				},
			},
		},
		CurrentContext: contextName,
		Preferences:    map[string]string{},
		Users: []KubeUser{
			KubeUser{
				UserName: userName,
				UserDef: KubeUserDef{
					ClientCertificateData: signedCert,
					ClientKeyData:         privateKey,
				},
			},
		},
	}

	yaml, err := yaml.Marshal(&kubeconfig)
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed marshaling user kubeconfig into yaml")
	}

	return string(yaml), nil
}
