package v1alpha1

import (
	"context"

	"knative.dev/eventing-kafka/pkg/common/constants"
	"knative.dev/eventing/pkg/apis/messaging"
)

func (c *KafkaTopicChannel) SetDefaults(ctx context.Context) {
	// Set the duck subscription to the stored version of the duck
	// we support. Reason for this is that the stored version will
	// not get a chance to get modified, but for newer versions
	// conversion webhook will be able to take a crack at it and
	// can modify it to match the duck shape.
	if c.Annotations == nil {
		c.Annotations = make(map[string]string)
	}

	// TODO this should be wrapped in a if to avoid conversion of ducks and let some other duck controller do that?!
	//   BUT we need this hack to properly update from v1alpha1 to v1beta1, so:
	//   * KEEP THIS FOR THE WHOLE 0.22 LIFECYCLE
	//   * REMOVE THIS BEFORE 0.23 RELEASE
	//if _, ok := c.Annotations[messaging.SubscribableDuckVersionAnnotation]; !ok {
	c.Annotations[messaging.SubscribableDuckVersionAnnotation] = "v1"
	//}

	c.Spec.SetDefaults(ctx)
}

func (cs *KafkaTopicChannelSpec) SetDefaults(ctx context.Context) {
	if cs.NumPartitions == 0 {
		cs.NumPartitions = constants.DefaultNumPartitions
	}
	if cs.ReplicationFactor == 0 {
		cs.ReplicationFactor = constants.DefaultReplicationFactor
	}
}
