// +build !ignore_autogenerated

/*
Copyright 2021 The Optum Authors
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicChannel) DeepCopyInto(out *KafkaTopicChannel) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicChannel.
func (in *KafkaTopicChannel) DeepCopy() *KafkaTopicChannel {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicChannel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KafkaTopicChannel) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicChannelList) DeepCopyInto(out *KafkaTopicChannelList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]KafkaTopicChannel, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicChannelList.
func (in *KafkaTopicChannelList) DeepCopy() *KafkaTopicChannelList {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicChannelList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *KafkaTopicChannelList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicChannelSpec) DeepCopyInto(out *KafkaTopicChannelSpec) {
	*out = *in
	in.KafkaAuthSpec.DeepCopyInto(&out.KafkaAuthSpec)
	in.ChannelableSpec.DeepCopyInto(&out.ChannelableSpec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicChannelSpec.
func (in *KafkaTopicChannelSpec) DeepCopy() *KafkaTopicChannelSpec {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicChannelSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KafkaTopicChannelStatus) DeepCopyInto(out *KafkaTopicChannelStatus) {
	*out = *in
	in.ChannelableStatus.DeepCopyInto(&out.ChannelableStatus)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KafkaTopicChannelStatus.
func (in *KafkaTopicChannelStatus) DeepCopy() *KafkaTopicChannelStatus {
	if in == nil {
		return nil
	}
	out := new(KafkaTopicChannelStatus)
	in.DeepCopyInto(out)
	return out
}
