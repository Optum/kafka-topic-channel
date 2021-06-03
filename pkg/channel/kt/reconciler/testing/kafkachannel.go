package testing

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/types"

	v1 "knative.dev/eventing/pkg/apis/duck/v1"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/optum/kafka-topic-channel/pkg/apis/messaging/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

// KafkaChannelOption enables further configuration of a KafkaChannel.
type KafkaChannelOption func(*v1alpha1.KafkaTopicChannel)

// NewKafkaChannel creates an KafkaChannel with KafkaChannelOptions.
func NewKafkaChannel(name, namespace string, ncopt ...KafkaChannelOption) *v1alpha1.KafkaTopicChannel {
	nc := &v1alpha1.KafkaTopicChannel{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			UID:       types.UID("abc-xyz"),
		},
		Spec: v1alpha1.KafkaTopicChannelSpec{},
	}
	for _, opt := range ncopt {
		opt(nc)
	}
	nc.SetDefaults(context.Background())
	return nc
}

func WithInitKafkaChannelConditions(nc *v1alpha1.KafkaTopicChannel) {
	nc.Status.InitializeConditions()
}

func WithKafkaChannelDeleted(nc *v1alpha1.KafkaTopicChannel) {
	deleteTime := metav1.NewTime(time.Unix(1e9, 0))
	nc.ObjectMeta.SetDeletionTimestamp(&deleteTime)
}

func WithKafkaChannelTopicReady() KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkTopicTrue()
	}
}

func WithKafkaChannelConfigReady() KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkConfigTrue()
	}
}

func WithKafkaChannelDeploymentNotReady(reason, message string) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkDispatcherFailed(reason, message)
	}
}

func WithKafkaChannelDeploymentReady() KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.PropagateDispatcherStatus(&appsv1.DeploymentStatus{Conditions: []appsv1.DeploymentCondition{{Type: appsv1.DeploymentAvailable, Status: corev1.ConditionTrue}}})
	}
}

func WithKafkaChannelServicetNotReady(reason, message string) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkServiceFailed(reason, message)
	}
}

func WithKafkaChannelServiceReady() KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkServiceTrue()
	}
}

func WithKafkaChannelChannelServicetNotReady(reason, message string) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkChannelServiceFailed(reason, message)
	}
}

func WithKafkaChannelChannelServiceReady() KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkChannelServiceTrue()
	}
}

func WithKafkaChannelEndpointsNotReady(reason, message string) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkEndpointsFailed(reason, message)
	}
}

func WithKafkaChannelEndpointsReady() KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.MarkEndpointsTrue()
	}
}

func WithKafkaChannelAddress(a string) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Status.SetAddress(&apis.URL{
			Scheme: "http",
			Host:   a,
		})
	}
}

func WithKafkaFinalizer(finalizerName string) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		finalizers := sets.NewString(nc.Finalizers...)
		finalizers.Insert(finalizerName)
		nc.SetFinalizers(finalizers.List())
	}
}

func WithKafkaChannelSubscribers(subs []v1.SubscriberSpec) KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		nc.Spec.Subscribers = subs
	}
}

func WithKafkaChannelStatusSubscribers() KafkaChannelOption {
	return func(nc *v1alpha1.KafkaTopicChannel) {
		ss := make([]v1.SubscriberStatus, len(nc.Spec.Subscribers))
		for _, s := range nc.Spec.Subscribers {
			ss = append(ss, v1.SubscriberStatus{
				UID:                s.UID,
				ObservedGeneration: s.Generation,
				Ready:              corev1.ConditionTrue,
			})
		}
		nc.Status.Subscribers = ss
	}
}
