package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/batch/v1"
	certsapi "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceBundleState is the Schema for the ResourceBundleStatees API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +genclient
type ResourceBundleState struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata" yaml:"metadata"`

	Spec   ResourceBundleStateSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ResourceBundleStatus    `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// ResourceBundleStateSpec defines the desired state of ResourceBundleState
// +k8s:openapi-gen=true
type ResourceBundleStateSpec struct {
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,1,opt,name=selector"`
}

// ResourceBundleStatus defines the observed state of ResourceBundleState
// +k8s:openapi-gen=true
type ResourceBundleStatus struct {
	Ready               bool                                 `json:"ready" protobuf:"varint,1,opt,name=ready"`
	ResourceCount       int32                                `json:"resourceCount" protobuf:"varint,2,opt,name=resourceCount"`
	PodStatuses         []corev1.Pod                         `json:"podStatuses" protobuf:"varint,3,opt,name=podStatuses"`
	ServiceStatuses     []corev1.Service                     `json:"serviceStatuses" protobuf:"varint,4,opt,name=serviceStatuses"`
	ConfigMapStatuses   []corev1.ConfigMap                   `json:"configMapStatuses" protobuf:"varint,5,opt,name=configMapStatuses"`
	DeploymentStatuses  []appsv1.Deployment                  `json:"deploymentStatuses" protobuf:"varint,6,opt,name=deploymentStatuses"`
	SecretStatuses      []corev1.Secret                      `json:"secretStatuses" protobuf:"varint,7,opt,name=secretStatuses"`
	DaemonSetStatuses   []appsv1.DaemonSet                   `json:"daemonSetStatuses" protobuf:"varint,8,opt,name=daemonSetStatuses"`
	IngressStatuses     []v1beta1.Ingress                    `json:"ingressStatuses" protobuf:"varint,11,opt,name=ingressStatuses"`
	JobStatuses         []v1.Job                             `json:"jobStatuses" protobuf:"varint,12,opt,name=jobStatuses"`
	StatefulSetStatuses []appsv1.StatefulSet                 `json:"statefulSetStatuses" protobuf:"varint,13,opt,name=statefulSetStatuses"`
	CsrStatuses         []certsapi.CertificateSigningRequest `json:"csrStatuses" protobuf:"varint,3,opt,name=csrStatuses"`
}

// PodStatus defines the observed state of ResourceBundleState
// +k8s:openapi-gen=true
type PodStatus struct {
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Ready             bool             `json:"ready" protobuf:"varint,2,opt,name=ready"`
	Status            corev1.PodStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceBundleStateList contains a list of ResourceBundleState
type ResourceBundleStateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceBundleState `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceBundleState{}, &ResourceBundleStateList{})
}
