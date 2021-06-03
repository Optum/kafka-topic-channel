package messaging

import "k8s.io/apimachinery/pkg/runtime/schema"

const (
	GroupName = "messaging.optum.dev"
)

var (
	// KafkaChannelsResource represents a KafkaChannel
	KafkaChannelsResource = schema.GroupResource{
		Group:    GroupName,
		Resource: "kafkatopicchannels",
	}
)
