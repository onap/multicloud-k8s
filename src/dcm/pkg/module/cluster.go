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
}

type ClusterKey struct {
	Project          string `json:"project"`
	LogicalCloudName string `json:"logical-cloud-name"`
	ClusterReference string `json:"clname"`
}

////////// .kube/config structs here:
type KubeConfig struct {
	ApiVersion     string            `yaml:"apiVersion"`
	Kind           string            `yaml:"kind"`
	Clusters       []KubeCluster     `yaml:"clusters"`
	Contexts       []KubeContext     `yaml:"contexts"`
	CurrentContext string            `yaml:"current-context`
	Preferences    map[string]string `yaml:"preferences"`
	Users          []KubeUser        `yaml":"users"`
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

////////////////////////////////////////

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

	return Cluster{}, pkgerrors.New("Error getting Cluster")
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
		return Cluster{}, pkgerrors.New("Update Error - Cluster reference mismatch")
	}
	//Check if this Cluster reference exists
	_, err := v.GetCluster(project, logicalCloud, clusterReference)
	if err != nil {
		return Cluster{}, pkgerrors.New("Update Error - Cluster reference doesn't exist")
	}
	err = v.util.DBInsert(v.storeName, key, nil, v.tagMeta, c)
	if err != nil {
		return Cluster{}, pkgerrors.Wrap(err, "Updating DB Entry")
	}
	return c, nil
}

// Get returns Cluster's kubeconfig for corresponding cluster reference
func (v *ClusterClient) GetClusterConfig(project, logicalCloud, clusterReference string) (string, error) {
	// private key comes from logical cloud
	lckey := LogicalCloudKey{
		Project:          project,
		LogicalCloudName: logicalCloud,
	}

	// TODO: signed cert comes from clusters
	// ckey := ClusterKey{
	// 	Project:          project,
	// 	LogicalCloudName: logicalCloud,
	// 	ClusterReference: clusterReference,
	// }

	// get user's private key
	value, err := v.util.DBFind(v.storeName, lckey, "privatekey")
	if err != nil {
		return "", pkgerrors.Wrap(err, "Failed getting private key from logical cloud")
	}

	// get logical cloud resource
	lcClient := NewLogicalCloudClient()
	lc, err := lcClient.Get(project, logicalCloud)

	privateKey := string(value[0])
	// TODO:
	signedCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUQ4ekNDQXR1Z0F3SUJBZ0lVRnE0Y0EveXBwZ3ZSM3llSUdvai91VUZrN0pzd0RRWUpLb1pJaHZjTkFRRUwKQlFBd0ZURVRNQkVHQTFVRUF4TUthM1ZpWlhKdVpYUmxjekFlRncweU1EQTRNRGN3TURRNU1EQmFGdzB5TVRBNApNRGN3TURRNU1EQmFNQkV4RHpBTkJnTlZCQU1UQm5WelpYSXRNVENDQWlJd0RRWUpLb1pJaHZjTkFRRUJCUUFECmdnSVBBRENDQWdvQ2dnSUJBSm5mWlFybGdIOXZmaUxPNThpaDRWNnRQL0RVMnp3V3hyV2ljeWhEMUFaUjNKNlgKV2dmcWxQVVpKQ0tDRHhLejg3Z3kzdEttMlFVU1dVaWhOVGd5bHJIRTAzbVVFZ3AwOVAwZDM3UVpicmVHWUttSApELzNiZE5oSjZBZGhHZVcxdFhSeU9YRDRGR0JnMmFYcmlpaFpnSng1ZFRMVVVwRnludVZEdjhhM0kwNmRWUnRWCkVUVmFUU1Q3K00wOEhZTDNIN2VZT0FFM3JXL2Q5K2Nqc3hnZ0VlcWNDTCtuREhpWTkwSC9DaS9pajZIWnNGU1cKRFd1R2l6Uk1MZ1oxNTlxdkd0bEtRcXgrRmNlWlFDZGJTQ1hHQ2Y3YnpZN3cvVmFzSnFJZXNuUlA2Z245eWJFVQpjd0ZrTDVJdUZxVUZFWHNQUm1HQ0RsUzFQVFpmVXVhKytqWjlNSjYvRmZZQ0c3L3dlcDRzd0dhblovOFB5bkpXCjhJRU8rekZSSGJMK0dkUmZDc2RiL1NWTjZRaUFuTHIxRXBjWjVPb0NnYm13dStQMm5wUCtJVHJ2OUhSdE1VVFMKSllXUzdRa2RWcmxOOEM0UlhBTFowQW5qTXZJUHozbk9GcTZoaDVpY0VMZUg0Y3Q4TXkvc2JBL2ZkVFp6R3hWUQp0cVdDZ3VqSGFtbWpxSGdBSDRDaENOYlAvWE96eEVVWTRRT0gxRXJteEREVUdKTzlCNGNGblZsZlh4V2lNeVpzCmE4K2Y4Q3J5SU13aVhBNFNxM1Q3NnBmVTVKcmVZTXFLbm5jb05hSkNoak0yUE9xRzZrTHBsQjVrb2N0czZsZDMKb1ptRG45a081SlpYTmVkUjU0SHZkWms1Q2tiVjNBSVpVYndVMTlicVVEUVVieG9UV1EzWUplSGl0dmFqQWdNQgpBQUdqUHpBOU1BNEdBMVVkRHdFQi93UUVBd0lGb0RBTUJnTlZIUk1CQWY4RUFqQUFNQjBHQTFVZERnUVdCQlIzCktTZk1QV1J4VmdGeWZWQ2o5dWJwMVVQRU5EQU5CZ2txaGtpRzl3MEJBUXNGQUFPQ0FRRUFSTldMNVBwLzR4TUYKVTE0WXRMa1NYYlF5bmNHaVFnSGMyVGVaTGtxMmRvZW16d1FtajBuckpmYm90Q3RHZHUrTHJTTVhoNCtuWmpsMQpob1A3UHk0SEh4cGtVK01nOVdUZmNLTVZYT1FONWE4eE1ScG1vanFPV0czY2VhdjQ3SW8wSitEVFBpTUI1N1BFCnhpMUFwMVVHN0NpRW9oeElhT0dtVEt3STdYdzVGaUJBc0VQalFrS3RuOHhYWUZ5OHp4bE51ZlVoOUx6UERuSkgKN3Vpa0w1NU9uREo2NEdFeVFjQjZCeVVjUlZkL1IvVHFtOUJHKzVBWGpqSVFua21Gcnp3bEdCdzlnMGZPWHphZwozVjdZYm9QNE4xMTdLZStwMlNLNmkyT0RSYUtWUzVsMnNCUWtRZm1wYlM3VDRvbTlTR1BJdnF0b012K2dTTFNwCis1dXhFZlViVlE9PQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg=="
	// TODO:
	clusterCert := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJd01EY3dNakl4TXpjeE0xb1hEVE13TURZek1ESXhNemN4TTFvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBUEZVCkVwZWMxSmRSSFlUYko1dHEwemM5aEpxR2l2QnlvUXlxZFJ2YjVQK2dCeG9VL01RMHhnSDhrbFA5Ykp5VitxZUsKYUI0S2lkSk1HYWdUV2FhdUpCYlFHYmdnaldGSzR3U3pXakExM0RCQ2xSRWlLVG5MSnkvK3gzK0lOODIrY2MvUApNTWhkSUE2aW9NMjRpa2JkQ3RobkRSVWNUNU4wdDBVaGxMc1p3cWM2bXhJTVcxSGM0ckpsRmY3UVpCbmh3N0lLCjIzRlREcWJ5Ykw2Y1Q2OHFOdEg3RFJaNURUV3lzOTJMR0NCcXlPUW9kZS9TQVFHcGpBRkt0L1FDM2lXUW43RXkKbkxSZjBIOWR6MEJVUUhaQUhMd21IVkdUdFJ5ZWMyMU5uQzZHcUsralY4d1BCcXR1Q2xWUGdOT0xLNTJxSkd5VgpqSE5DckR6TWVpRHR3dTMzL1VVQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFGRXJScFpGMnk4dDcxQkJCOXN4cC9kbGMwY3gKblVZRSt2alFnT2FrTUltVHcrTkEyR2l0a2NtNUtNeDBwQzNKUmg5MW1sYTFOZ28xZHcwRkE1ZE5BZW5Ua1lmSgpiVXhqM0gzZXhVSmtFekgxbSs4dTk3d0ZzVTNkZ3d6VWNQRkFUcmlVeEZvdjA0b1NzSUExa3ZLb1ZkaThlUnBKCjJXN1Y4bzNpdVFTR1lXZHNMcDhTbWJwazlhQS9qK2d4TXpNOFBpN1loQUpSM0lOcURsaU1CbnJaVTRzaU5LRTUKTDhNRVV5Y1dRbmNHTDJUbUZDU25vN1ZjTmtlTWdudFZpbGtoaUpjOEdNV1VzakVkSmNxbm9SOGs1ODdjTXBJZQorcm9kMlhsRTBmakZRRG9FSnRhYXlZQjNEWFRBSU4vMWNkRjE1bWdCN0Y3R2tnNGNPM0lPSmxZQXBEMD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	// TODO:
	clusterAddr := "https://192.168.121.203:6443"
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

	return string(yaml), nil

	// //value is a byte array
	// if value != nil {
	// 	cl := Cluster{}
	// 	err = v.util.DBUnmarshal(value[0], &cl)
	// 	if err != nil {
	// 		return Cluster{}, pkgerrors.Wrap(err, "Unmarshaling value")
	// 	}
	// 	return cl, nil
	// }

	return "", pkgerrors.New("Error getting Cluster's kubeconfig")
}
