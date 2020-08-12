// +build !ignore_autogenerated

// Code generated by operator-sdk. DO NOT EDIT.

package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

/*
// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodStatus) DeepCopyInto(out *PodStatus) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodStatus.
func (in *PodStatus) DeepCopy() *PodStatus {
	if in == nil {
		return nil
	}
	out := new(PodStatus)
	in.DeepCopyInto(out)
	return out
}
*/

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceBundleState) DeepCopyInto(out *ResourceBundleState) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceBundleState.
func (in *ResourceBundleState) DeepCopy() *ResourceBundleState {
	if in == nil {
		return nil
	}
	out := new(ResourceBundleState)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceBundleState) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceBundleStateList) DeepCopyInto(out *ResourceBundleStateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ResourceBundleState, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceBundleStateList.
func (in *ResourceBundleStateList) DeepCopy() *ResourceBundleStateList {
	if in == nil {
		return nil
	}
	out := new(ResourceBundleStateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ResourceBundleStateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceBundleStateSpec) DeepCopyInto(out *ResourceBundleStateSpec) {
	*out = *in
	if in.Selector != nil {
		in, out := &in.Selector, &out.Selector
		*out = new(v1.LabelSelector)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceBundleStateSpec.
func (in *ResourceBundleStateSpec) DeepCopy() *ResourceBundleStateSpec {
	if in == nil {
		return nil
	}
	out := new(ResourceBundleStateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResourceBundleStatus) DeepCopyInto(out *ResourceBundleStatus) {
	*out = *in
	if in.PodStatuses != nil {
		in, out := &in.PodStatuses, &out.PodStatuses
		*out = make([]corev1.Pod, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ServiceStatuses != nil {
		in, out := &in.ServiceStatuses, &out.ServiceStatuses
		*out = make([]corev1.Service, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ConfigMapStatuses != nil {
		in, out := &in.ConfigMapStatuses, &out.ConfigMapStatuses
		*out = make([]corev1.ConfigMap, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DeploymentStatuses != nil {
		in, out := &in.DeploymentStatuses, &out.DeploymentStatuses
		*out = make([]appsv1.Deployment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SecretStatuses != nil {
		in, out := &in.SecretStatuses, &out.SecretStatuses
		*out = make([]corev1.Secret, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.DaemonSetStatuses != nil {
		in, out := &in.DaemonSetStatuses, &out.DaemonSetStatuses
		*out = make([]appsv1.DaemonSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.IngressStatuses != nil {
		in, out := &in.IngressStatuses, &out.IngressStatuses
		*out = make([]v1beta1.Ingress, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.JobStatuses != nil {
		in, out := &in.JobStatuses, &out.JobStatuses
		*out = make([]batchv1.Job, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.StatefulSetStatuses != nil {
		in, out := &in.StatefulSetStatuses, &out.StatefulSetStatuses
		*out = make([]appsv1.StatefulSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResourceBundleStatus.
func (in *ResourceBundleStatus) DeepCopy() *ResourceBundleStatus {
	if in == nil {
		return nil
	}
	out := new(ResourceBundleStatus)
	in.DeepCopyInto(out)
	return out
}
