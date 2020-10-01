/*
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
package client

import (
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/appcontext/subresources"
	pkgerrors "github.com/pkg/errors"
	"github.com/onap/multicloud-k8s/src/orchestrator/pkg/infra/logutils"
)

func (c *Client) Approve(name string, sa []byte) error {

	var a subresources.ApprovalSubresource
	err := json.Unmarshal(sa, &a)
	if err != nil {
		return pkgerrors.Wrap(err, "An error occurred while parsing the approval Subresource.")
	}
	csr, err := c.Clientset.CertificatesV1beta1().CertificateSigningRequests().Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	var timePtr metav1.Time
	str := []string{a.LastUpdateTime}
	if err = metav1.Convert_Slice_string_To_v1_Time(&str, &timePtr, nil); err != nil {
	    return pkgerrors.Wrap(err, "An error occurred while converting time from string.")
	}
	// Update CSR with Conditions
	csr.Status.Conditions = append(csr.Status.Conditions, certificatesv1beta1.CertificateSigningRequestCondition{
		Type:           certificatesv1beta1.RequestConditionType(a.Type),
		Reason:         a.Reason,
		Message:        a.Message,
		LastUpdateTime: timePtr,
	})
	// CSR Approval
	_, err = c.Clientset.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(csr)
	if err != nil {
		logutils.Error("Failed to UpdateApproval", logutils.Fields{
			"error":    err,
			"resource": name,
		})
		return err
	}
	return nil
}
