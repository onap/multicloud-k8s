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

package subresources

// The ApprovalSubresource type defines the 4 necessary parameters
// that the "approval" subresource of a CertificateSigningRequest in K8s
// requires, in a forma tto be exchanged over AppContext
type ApprovalSubresource struct {
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	Message        string `json:"message,omitempty"`
	Reason         string `json:"reason,omitempty"`
	Type           string `json:"type,omitempty"`
}
