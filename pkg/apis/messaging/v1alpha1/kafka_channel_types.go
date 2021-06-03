package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	bindingsv1beta1 "knative.dev/eventing-kafka/pkg/apis/bindings/v1beta1"
	eventingduck "knative.dev/eventing/pkg/apis/duck/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaTopicChannel is a resource representing a Kafka Topic Channel.
type KafkaTopicChannel struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the Channel.
	Spec KafkaTopicChannelSpec `json:"spec,omitempty"`

	// Status represents the current state of the KafkaTopicChannel. This data may be out of
	// date.
	// +optional
	Status KafkaTopicChannelStatus `json:"status,omitempty"`
}

var (
	// Check that this channel can be validated and defaulted.
	_ apis.Validatable = (*KafkaTopicChannel)(nil)
	_ apis.Defaultable = (*KafkaTopicChannel)(nil)

	_ runtime.Object = (*KafkaTopicChannel)(nil)

	// Check that we can create OwnerReferences to an this channel.
	_ kmeta.OwnerRefable = (*KafkaTopicChannel)(nil)

	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*KafkaTopicChannel)(nil)
)

// KafkaTopicChannelSpec defines the specification for a KafkaTopicChannel.
type KafkaTopicChannelSpec struct {
	// NumPartitions is the number of partitions of a Kafka topic. By default, it is set to 1.
	NumPartitions int32 `json:"numPartitions"`

	// ReplicationFactor is the replication factor of a Kafka topic. By default, it is set to 1.
	ReplicationFactor int16 `json:"replicationFactor"`

	bindingsv1beta1.KafkaAuthSpec `json:",inline"`

	// Topic topics to consume messages from
	// +required
	Topic string `json:"topic"`

	// Channel conforms to Duck type Channelable.
	eventingduck.ChannelableSpec `json:",inline"`
}

// KafkaTopicChannelStatus represents the current state of a KafkaTopicChannel.
type KafkaTopicChannelStatus struct {
	// Channel conforms to Duck type Channelable.
	eventingduck.ChannelableStatus `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KafkaTopicChannelList is a collection of KafkaTopicChannels.
type KafkaTopicChannelList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaTopicChannel `json:"items"`
}

// GetGroupVersionKind returns GroupVersionKind for KafkaTopicChannels
func (c *KafkaTopicChannel) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("KafkaTopicChannel")
}

// GetStatus retrieves the duck status for this resource. Implements the KRShaped interface.
func (k *KafkaTopicChannel) GetStatus() *duckv1.Status {
	return &k.Status.Status
}
