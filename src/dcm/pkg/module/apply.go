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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/grpc/installappclient"
	log "github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/module/controller"
	pkgerrors "github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// rsyncName denotes the name of the rsync controller
const rsyncName = "rsync"

type Resource struct {
	ApiVersion    string         `yaml:"apiVersion"`
	Kind          string         `yaml:"kind"`
	MetaData      MetaDatas      `yaml:"metadata"`
	Specification Specs          `yaml:"spec,omitempty"`
	Rules         []RoleRules    `yaml:"rules,omitempty"`
	Subjects      []RoleSubjects `yaml:"subjects,omitempty"`
	RoleRefs      RoleRef        `yaml:"roleRef,omitempty"`
}

type MetaDatas struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace,omitempty"`
}

type Specs struct {
	Request string   `yaml:"request,omitempty"`
	Usages  []string `yaml:"usages,omitempty"`
	//Hard           logicalcloud.QSpec    `yaml:"hard,omitempty"`
	Hard QSpec `yaml:"hard,omitempty"`
}

type RoleRules struct {
	ApiGroups []string `yaml:"apiGroups"`
	Resources []string `yaml:"resources"`
	Verbs     []string `yaml:"verbs"`
}

type RoleSubjects struct {
	Kind     string `yaml:"kind"`
	Name     string `yaml:"name"`
	ApiGroup string `yaml:"apiGroup"`
}

type RoleRef struct {
	Kind     string `yaml:"kind"`
	Name     string `yaml:"name"`
	ApiGroup string `yaml:"apiGroup"`
}

func createNamespace(logicalcloud LogicalCloud) (string, error) {

	namespace := Resource{
		ApiVersion: "v1",
		Kind:       "Namespace",
		MetaData: MetaDatas{
			Name: logicalcloud.Specification.NameSpace,
		},
	}

	nsData, err := yaml.Marshal(&namespace)
	if err != nil {
		return "", err
	}

	return string(nsData), nil
}

func createRole(logicalcloud LogicalCloud) (string, error) {

	userPermissions := logicalcloud.Specification.User.UserPermissions[0]

	role := Resource{
		ApiVersion: "rbac.authorization.k8s.io/v1beta1",
		Kind:       "Role",
		MetaData: MetaDatas{
			Name:      strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-role"}, ""),
			Namespace: logicalcloud.Specification.NameSpace,
		},
		Rules: []RoleRules{RoleRules{
			ApiGroups: userPermissions.APIGroups,
			Resources: userPermissions.Resources,
			Verbs:     userPermissions.Verbs,
		},
		},
	}

	roleData, err := yaml.Marshal(&role)
	if err != nil {
		return "", err
	}

	return string(roleData), nil
}

func createRoleBinding(logicalcloud LogicalCloud) (string, error) {

	roleBinding := Resource{
		ApiVersion: "rbac.authorization.k8s.io/v1beta1",
		Kind:       "RoleBinding",
		MetaData: MetaDatas{
			Name:      strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-roleBinding"}, ""),
			Namespace: logicalcloud.Specification.NameSpace,
		},
		Subjects: []RoleSubjects{RoleSubjects{
			Kind:     "User",
			Name:     logicalcloud.Specification.User.UserName,
			ApiGroup: "",
		},
		},

		RoleRefs: RoleRef{
			Kind:     "Role",
			Name:     strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-role"}, ""),
			ApiGroup: "",
		},
	}

	rBData, err := yaml.Marshal(&roleBinding)
	if err != nil {
		return "", err
	}

	return string(rBData), nil

}

func createQuota(quota []Quota, namespace string) (string, error) {
	lcQuota := quota[0]

	q := Resource{
		ApiVersion: "v1",
		Kind:       "ResourceQuota",
		MetaData: MetaDatas{
			Name:      lcQuota.MetaData.QuotaName,
			Namespace: namespace,
		},
		Specification: Specs{
			Hard: lcQuota.Specification,
		},
	}

	qData, err := yaml.Marshal(&q)
	if err != nil {
		return "", err
	}

	return string(qData), nil

}

func createUserCSR(logicalcloud LogicalCloud) (string, error) {
	KEYSIZE := 4096
	userName := logicalcloud.Specification.User.UserName

	key, err := rsa.GenerateKey(rand.Reader, KEYSIZE)
	if err != nil {
		return "", err
	}

	csrTemplate := x509.CertificateRequest{Subject: pkix.Name{CommonName: userName}}

	csrCert, err := x509.CreateCertificateRequest(rand.Reader, &csrTemplate, key)
	if err != nil {
		return "", err
	}

	//Encode csr
	csr := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE REQUEST",
		Bytes: csrCert,
	})

	csrObj := Resource{
		ApiVersion: "certificates.k8s.io/v1beta1",
		Kind:       "CertificateSigningRequest",
		MetaData: MetaDatas{
			Name:      strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "-user-csr"}, ""),
			Namespace: logicalcloud.Specification.NameSpace,
		},
		Specification: Specs{
			Request: base64.StdEncoding.EncodeToString(csr),
			Usages:  []string{"digital signature", "key encipherment"},
		},
	}

	csrData, err := yaml.Marshal(&csrObj)
	if err != nil {
		return "", err
	}

	return string(csrData), nil

}

/*
queryDBAndSetRsyncInfo queries the MCO db to find the record the sync controller
and then sets the RsyncInfo global variable.
*/
func queryDBAndSetRsyncInfo() (installappclient.RsyncInfo, error) {
	client := controller.NewControllerClient()
	vals, _ := client.GetControllers()
	for _, v := range vals {
		if v.Metadata.Name == rsyncName {
			log.Info("Initializing RPC connection to resource synchronizer", log.Fields{
				"Controller": v.Metadata.Name,
			})
			rsyncInfo := installappclient.NewRsyncInfo(v.Metadata.Name, v.Spec.Host, v.Spec.Port)
			return rsyncInfo, nil
		}
	}
	return installappclient.RsyncInfo{}, pkgerrors.Errorf("queryRsyncInfoInMCODB Failed - Could not get find rsync by name : %v", rsyncName)
}

/*
callRsyncUninstall method shall take in the app context id and invoke the rsync service via grpc
*/
func callRsyncUninstall(contextid interface{}) error {
	rsyncInfo, err := queryDBAndSetRsyncInfo()
	log.Info("Calling the Rsync ", log.Fields{
		"RsyncName": rsyncInfo.RsyncName,
	})
	if err != nil {
		return err
	}

	appContextID := fmt.Sprintf("%v", contextid)
	err = installappclient.InvokeUninstallApp(appContextID)
	if err != nil {
		return err
	}
	return nil
}

// TODO:
// Install istio
// Store user key for user creation
// Code to run kubectl commands for user
// kubectl certificate approve lc1-user-cert
// kubectl get csr lc1-user-cert -o jsonpath='{.status.certificate}' | base64 --decode > user.crt
// kubectl config set-credentials user --client-certificate=<user.crt>  --client-key=<user.key>
// kubectl config set-context user-context --cluster=cluster-name --namespace=lc1 --user=user

func CreateEtcdContext(logicalcloud LogicalCloud, clusterList []Cluster,
	quotaList []Quota) error {

	APP := "logical-cloud"
	logicalCloudName := logicalcloud.MetaData.LogicalCloudName

	//Resource Names
	namespaceName := strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "+namespace"}, "")
	roleName := strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "+role"}, "")
	roleBindingName := strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "+roleBinding"}, "")
	quotaName := strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "+quota"}, "")
	csrName := strings.Join([]string{logicalcloud.MetaData.LogicalCloudName, "+CertificateSigningRequest"}, "")

	// Get resources to be added
	namespace, err := createNamespace(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Namespace YAML for logical cloud")
	}

	role, err := createRole(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Role YAML for logical cloud")
	}

	roleBinding, err := createRoleBinding(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating RoleBinding YAML for logical cloud")
	}

	quota, err := createQuota(quotaList, logicalcloud.Specification.NameSpace)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating Quota YAML for logical cloud")
	}

	csr, err := createUserCSR(logicalcloud)
	if err != nil {
		return pkgerrors.Wrap(err, "Error Creating User CSR for logical cloud")
	}

	context := appcontext.AppContext{}
	ctxVal, err := context.InitAppContext()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext")
	}

	fmt.Printf("%v\n", ctxVal)

	handle, err := context.CreateCompositeApp()
	if err != nil {
		return pkgerrors.Wrap(err, "Error creating AppContext CompositeApp")
	}

	appHandle, err := context.AddApp(handle, APP)
	if err != nil {
		cleanuperr := context.DeleteCompositeApp()
		if cleanuperr != nil {
			log.Warn("Error cleaning AppContext CompositeApp create failure", log.Fields{
				"logical-cloud": logicalCloudName,
			})
		}
		return pkgerrors.Wrap(err, "Error adding App to AppContext")
	}

	// Iterate through cluster list and add all the clusters
	for _, cluster := range clusterList {
		clusterName := strings.Join([]string{cluster.Specification.ClusterProvider, "+", cluster.Specification.ClusterName}, "")
		clusterHandle, err := context.AddCluster(appHandle, clusterName)

		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add cluster failure", log.Fields{
					"cluster-provider": cluster.Specification.ClusterProvider,
					"cluster":          cluster.Specification.ClusterName,
					"logical-cloud":    logicalCloudName,
				})
			}
			return pkgerrors.Wrap(err, "Error adding Cluster to AppContext")
		}

		// Add namespace resource to each cluster
		_, err = context.AddResource(clusterHandle, namespaceName, namespace)
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add namespace resource failure", log.Fields{
					"cluster-provider": cluster.Specification.ClusterProvider,
					"cluster":          cluster.Specification.ClusterName,
					"logical-cloud":    logicalCloudName,
				})
			}
			return pkgerrors.Wrap(err, "Error adding Namespace Resource to AppContext")
		}

		// Add csr resource to each cluster
		_, err = context.AddResource(clusterHandle, csrName, csr)
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add CSR resource failure", log.Fields{
					"cluster-provider": cluster.Specification.ClusterProvider,
					"cluster":          cluster.Specification.ClusterName,
					"logical-cloud":    logicalCloudName,
				})
			}
			return pkgerrors.Wrap(err, "Error adding CSR Resource to AppContext")
		}

		// Add Role resource to each cluster
		_, err = context.AddResource(clusterHandle, roleName, role)
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add role resource failure", log.Fields{
					"cluster-provider": cluster.Specification.ClusterProvider,
					"cluster":          cluster.Specification.ClusterName,
					"logical-cloud":    logicalCloudName,
				})
			}
			return pkgerrors.Wrap(err, "Error adding role Resource to AppContext")
		}

		// Add RoleBinding resource to each cluster
		_, err = context.AddResource(clusterHandle, roleBindingName, roleBinding)
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add roleBinding resource failure", log.Fields{
					"cluster-provider": cluster.Specification.ClusterProvider,
					"cluster":          cluster.Specification.ClusterName,
					"logical-cloud":    logicalCloudName,
				})
			}
			return pkgerrors.Wrap(err, "Error adding roleBinding Resource to AppContext")
		}

		// Add quota resource to each cluster
		_, err = context.AddResource(clusterHandle, quotaName, quota)
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add quota resource failure", log.Fields{
					"cluster-provider": cluster.Specification.ClusterProvider,
					"cluster":          cluster.Specification.ClusterName,
					"logical-cloud":    logicalCloudName,
				})
			}
			return pkgerrors.Wrap(err, "Error adding quota Resource to AppContext")
		}

		// Add Resource Order and Resource Dependency
		resOrder, err := json.Marshal(map[string][]string{"resorder": []string{namespaceName, quotaName, csrName, roleName, roleBindingName}})
		if err != nil {
			return pkgerrors.Wrap(err, "Error creating resource order JSON")
		}

		resDependency, err := json.Marshal(map[string]map[string]string{"resdependency": map[string]string{namespaceName: "go",
			quotaName: strings.Join([]string{"wait on ", namespaceName}, ""), csrName: strings.Join([]string{"wait on ", quotaName}, ""),
			roleName: strings.Join([]string{"wait on ", csrName}, ""), roleBindingName: strings.Join([]string{"wait on ", roleName}, "")}})

		if err != nil {
			return pkgerrors.Wrap(err, "Error creating resource dependency JSON")
		}

		_, err = context.AddInstruction(clusterHandle, "resource", "order", string(resOrder))
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add instruction  failure", log.Fields{
					"cluster-provider": cluster.Specification.ClusterProvider,
					"cluster":          cluster.Specification.ClusterName,
					"logical-cloud":    logicalCloudName,
				})
			}
			return pkgerrors.Wrap(err, "Error adding instruction order to AppContext")
		}

		_, err = context.AddInstruction(clusterHandle, "resource", "dependency", string(resDependency))
		if err != nil {
			cleanuperr := context.DeleteCompositeApp()
			if cleanuperr != nil {
				log.Warn("Error cleaning AppContext after add instruction  failure", log.Fields{
					"cluster-provider": cluster.Specification.ClusterProvider,
					"cluster":          cluster.Specification.ClusterName,
					"logical-cloud":    logicalCloudName,
				})
			}
			return pkgerrors.Wrap(err, "Error adding instruction dependency to AppContext")
		}

	}

	return nil

}

// TODO: rename these methods
// DestroyEtcdContext remove from rsync then delete appcontext and all resources
func DestroyEtcdContext(logicalcloud LogicalCloud, clusterList []Cluster,
	quotaList []Quota) error {

	logicalCloudName := logicalcloud.MetaData.LogicalCloudName
	// project := "test-project" // FIXME(igordc): temporary, need to do some rework in the LC structs

	_, ctxVal, err := NewLogicalCloudClient().GetLogicalCloudContext(logicalCloudName)
	if err != nil {
		return pkgerrors.Wrapf(err, "Error finding AppContext for Logical Cloud: %v", logicalCloudName)
	}

	// call resource synchronizer to delete the CRs from every cluster of the logical cloud
	err = callRsyncUninstall(ctxVal)
	if err != nil {
		return err
	}

	// TODO: status handling for logical cloud after terminate:
	// rsync updates the status of the appcontext to Terminated
	// dcm should launch thread to observe status of appcontext before concluding logical cloud is terminated
	// dcm should somewhat mimic the status tracking of rsync
	// logical cloud might be in a non-applied non-terminated state for a long period of time.........

	// // remove the app context
	// err = context.DeleteCompositeApp()
	// if err != nil {
	// 	return pkgerrors.Wrap(err, "Error deleting AppContext CompositeApp")
	// }

	// remove the app context field from the cluster db record
	// lckey := LogicalCloudKey{
	// 	LogicalCloudName: logicalcloud.MetaData.LogicalCloudName,
	// 	Project:          project,
	// }
	// err = db.DBconn.RemoveTag("orchestrator", lckey, "lccontext")
	// if err != nil {
	// 	log.Warn("Error removing AppContext from Logical Cloud", log.Fields{
	// 		"logical-cloud": logicalCloudName,
	// 	})
	// }

	return nil
}
