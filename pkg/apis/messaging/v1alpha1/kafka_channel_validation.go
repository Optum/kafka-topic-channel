package v1alpha1

import (
	"context"
	"fmt"

	"knative.dev/eventing/pkg/apis/eventing"
	"knative.dev/pkg/apis"
)

func (c *KafkaTopicChannel) Validate(ctx context.Context) *apis.FieldError {
	errs := c.Spec.Validate(ctx).ViaField("spec")

	// Validate annotations
	if c.Annotations != nil {
		if scope, ok := c.Annotations[eventing.ScopeAnnotationKey]; ok {
			if scope != "namespace" && scope != "cluster" {
				iv := apis.ErrInvalidValue(scope, "")
				iv.Details = "expected either 'cluster' or 'namespace'"
				errs = errs.Also(iv.ViaFieldKey("annotations", eventing.ScopeAnnotationKey).ViaField("metadata"))
			}
		}
	}

	return errs
}

func (cs *KafkaTopicChannelSpec) Validate(ctx context.Context) *apis.FieldError {
	var errs *apis.FieldError

	if cs.NumPartitions <= 0 {
		fe := apis.ErrInvalidValue(cs.NumPartitions, "numPartitions")
		errs = errs.Also(fe)
	}

	if cs.ReplicationFactor <= 0 {
		fe := apis.ErrInvalidValue(cs.ReplicationFactor, "replicationFactor")
		errs = errs.Also(fe)
	}

	for i, subscriber := range cs.SubscribableSpec.Subscribers {
		if subscriber.ReplyURI == nil && subscriber.SubscriberURI == nil {
			fe := apis.ErrMissingField("replyURI", "subscriberURI")
			fe.Details = "expected at least one of, got none"
			errs = errs.Also(fe.ViaField(fmt.Sprintf("subscriber[%d]", i)).ViaField("subscribable"))
		}
	}
	return errs
}
