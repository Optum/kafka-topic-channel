/*
Copyright 2021 The Optum Authors
*/

// Code generated by injection-gen. DO NOT EDIT.

package kafkatopicchannel

import (
	context "context"

	v1alpha1 "github.com/optum/kafka-topic-channel/pkg/client/informers/externalversions/messaging/v1alpha1"
	factory "github.com/optum/kafka-topic-channel/pkg/client/injection/informers/factory"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterInformer(withInformer)
}

// Key is used for associating the Informer inside the context.Context.
type Key struct{}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := factory.Get(ctx)
	inf := f.Messaging().V1alpha1().KafkaTopicChannels()
	return context.WithValue(ctx, Key{}, inf), inf.Informer()
}

// Get extracts the typed informer from the context.
func Get(ctx context.Context) v1alpha1.KafkaTopicChannelInformer {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch github.com/optum/kafka-topic-channel/pkg/client/informers/externalversions/messaging/v1alpha1.KafkaTopicChannelInformer from context.")
	}
	return untyped.(v1alpha1.KafkaTopicChannelInformer)
}
