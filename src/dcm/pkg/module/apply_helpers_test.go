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
	"testing"

	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext/subresources"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

// helper: build a representative LogicalCloud with a namespace, user and permissions
func sampleLogicalCloud() LogicalCloud {
	return LogicalCloud{
		MetaData: MetaDataList{
			LogicalCloudName: "test-lc",
		},
		Specification: Spec{
			NameSpace: "test-ns",
			User: UserData{
				UserName: "test-user",
				Type:     "certificate",
				UserPermissions: []UserPerm{
					{
						PermName:  "test-perm",
						APIGroups: []string{"", "apps"},
						Resources: []string{"pods", "deployments"},
						Verbs:     []string{"get", "list"},
					},
				},
			},
		},
	}
}

func TestCreateNamespace(t *testing.T) {
	lc := sampleLogicalCloud()

	nsData, nsName, err := createNamespace(lc)
	assert.NoError(t, err)
	assert.NotEmpty(t, nsData)
	assert.Equal(t, "test-ns+Namespace", nsName)

	// Round-trip the YAML to verify it is well-formed and carries the right fields.
	var res Resource
	err = yaml.Unmarshal([]byte(nsData), &res)
	assert.NoError(t, err)
	assert.Equal(t, "v1", res.ApiVersion)
	assert.Equal(t, "Namespace", res.Kind)
	assert.Equal(t, "test-ns", res.MetaData.Name)
}

func TestCreateRole(t *testing.T) {
	lc := sampleLogicalCloud()

	roleData, roleName, err := createRole(lc)
	assert.NoError(t, err)
	assert.NotEmpty(t, roleData)
	assert.Equal(t, "test-lc-role+Role", roleName)

	var res Resource
	err = yaml.Unmarshal([]byte(roleData), &res)
	assert.NoError(t, err)
	assert.Equal(t, "rbac.authorization.k8s.io/v1beta1", res.ApiVersion)
	assert.Equal(t, "Role", res.Kind)
	assert.Equal(t, "test-lc-role", res.MetaData.Name)
	assert.Equal(t, "test-ns", res.MetaData.Namespace)
	assert.Len(t, res.Rules, 1)
	assert.Equal(t, []string{"pods", "deployments"}, res.Rules[0].Resources)
	assert.Equal(t, []string{"get", "list"}, res.Rules[0].Verbs)
}

func TestCreateRoleBinding(t *testing.T) {
	lc := sampleLogicalCloud()

	rbData, rbName, err := createRoleBinding(lc)
	assert.NoError(t, err)
	assert.NotEmpty(t, rbData)
	assert.Equal(t, "test-lc-roleBinding+RoleBinding", rbName)

	var res Resource
	err = yaml.Unmarshal([]byte(rbData), &res)
	assert.NoError(t, err)
	assert.Equal(t, "rbac.authorization.k8s.io/v1beta1", res.ApiVersion)
	assert.Equal(t, "RoleBinding", res.Kind)
	assert.Equal(t, "test-lc-roleBinding", res.MetaData.Name)
	assert.Len(t, res.Subjects, 1)
	assert.Equal(t, "User", res.Subjects[0].Kind)
	assert.Equal(t, "test-user", res.Subjects[0].Name)
	assert.Equal(t, "Role", res.RoleRefs.Kind)
	assert.Equal(t, "test-lc-role", res.RoleRefs.Name)
}

func TestCreateQuotaResource(t *testing.T) {
	quotaList := []Quota{
		{
			MetaData: QMetaDataList{
				QuotaName: "test-quota",
			},
			Specification: map[string]string{
				"limits.cpu":    "400",
				"limits.memory": "1000Gi",
			},
		},
	}

	qData, qName, err := createQuota(quotaList, "test-ns")
	assert.NoError(t, err)
	assert.NotEmpty(t, qData)
	assert.Equal(t, "test-quota+ResourceQuota", qName)

	var res Resource
	err = yaml.Unmarshal([]byte(qData), &res)
	assert.NoError(t, err)
	assert.Equal(t, "v1", res.ApiVersion)
	assert.Equal(t, "ResourceQuota", res.Kind)
	assert.Equal(t, "test-quota", res.MetaData.Name)
	assert.Equal(t, "test-ns", res.MetaData.Namespace)
	assert.Equal(t, "400", res.Specification.Hard["limits.cpu"])
	assert.Equal(t, "1000Gi", res.Specification.Hard["limits.memory"])
}

func TestCreateUserCSR(t *testing.T) {
	lc := sampleLogicalCloud()

	csrData, keyData, csrName, err := createUserCSR(lc)
	assert.NoError(t, err)
	assert.NotEmpty(t, csrData)
	assert.NotEmpty(t, keyData)
	assert.Equal(t, "test-lc-user-csr+CertificateSigningRequest", csrName)

	// The CSR resource is YAML: round-trip it and validate the key fields.
	var res Resource
	err = yaml.Unmarshal([]byte(csrData), &res)
	assert.NoError(t, err)
	assert.Equal(t, "certificates.k8s.io/v1beta1", res.ApiVersion)
	assert.Equal(t, "CertificateSigningRequest", res.Kind)
	assert.Equal(t, "test-lc-user-csr", res.MetaData.Name)
	assert.Contains(t, res.Specification.Usages, "digital signature")
	assert.Contains(t, res.Specification.Usages, "key encipherment")

	// The embedded CSR request is base64-encoded PEM.
	assert.NotEmpty(t, res.Specification.Request)
	decodedCSR, err := base64.StdEncoding.DecodeString(res.Specification.Request)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(decodedCSR), "CERTIFICATE REQUEST"))

	// The private key is base64-encoded PEM as well.
	decodedKey, err := base64.StdEncoding.DecodeString(keyData)
	assert.NoError(t, err)
	assert.True(t, strings.Contains(string(decodedKey), "RSA PRIVATE KEY"))
}

func TestCreateUserCSRUniqueKeys(t *testing.T) {
	lc := sampleLogicalCloud()

	_, keyData1, _, err := createUserCSR(lc)
	assert.NoError(t, err)
	_, keyData2, _, err := createUserCSR(lc)
	assert.NoError(t, err)

	// Each invocation generates a fresh RSA key pair.
	assert.NotEqual(t, keyData1, keyData2)
}

func TestCreateApprovalSubresource(t *testing.T) {
	lc := sampleLogicalCloud()

	approval, err := createApprovalSubresource(lc)
	assert.NoError(t, err)
	assert.NotEmpty(t, approval)

	var sub subresources.ApprovalSubresource
	err = json.Unmarshal([]byte(approval), &sub)
	assert.NoError(t, err)
	assert.Equal(t, "Approved for Logical Cloud authentication", sub.Message)
	assert.Equal(t, "LogicalCloud", sub.Reason)
	assert.Equal(t, "Approved", sub.Type)
	assert.NotEmpty(t, sub.LastUpdateTime)
}
