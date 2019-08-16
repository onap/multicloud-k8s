package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ResourceBundleState is the Schema for the ResourceBundleStatees API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +genclient
type ResourceBundleState struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

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
	Ready           bool             `json:"ready" protobuf:"varint,1,opt,name=ready"`
	ResourceCount   int32            `json:"resourceCount" protobuf:"varint,2,opt,name=resourceCount"`
	PodStatuses     []PodStatus      `json:"podStatuses" protobuf:"varint,3,opt,name=podStatuses"`
	ServiceStatuses []corev1.Service `json:"serviceStatuses" protobuf:"varint,4,opt,name=serviceStatuses"`
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
